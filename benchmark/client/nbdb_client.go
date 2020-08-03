package main

import (
	"flag"
	"git.yayafish.com/nbagent/benchmark/client/config"
	pb "git.yayafish.com/nbagent/benchmark/pb"
	"git.yayafish.com/nbagent/client"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/taskworker"
	"github.com/golang/protobuf/proto"
	"github.com/montanaflynn/stats"
	"os"
	"os/signal"
	"path"
	"reflect"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	nConcurrency int
	nInitTotal   int
	nTotal       int
	szHost       string
	nPort        int
	nInitConn int
	nMaxConn int
	nGoodByeTimeoutS int
	szSecretKey  string
)

func init() {
	if config.ServerConf.HasConfig {
		_ = log.InitLog(2, log.SetTarget(config.ServerConf.LogTarget), log.SetEncode(config.ServerConf.LogEncode),
			log.LogFilePath(config.ServerConf.LogPath+path.Base(os.Args[0])+"/"),
			log.LogFileRotate("date"),
		)
		_ = log.SetLogLevel("debug")
	} else {
		_ = log.InitLog(2, log.SetTarget("asyncfile"), log.SetEncode("json"),
			log.LogFilePath("./log/"+path.Base(os.Args[0])+"/"))
		_ = log.SetLogLevel("debug")
	}
}

func main() {
	//szHost, _ = os.LookupEnv("HOST")
	//if szHost == "" {
	//	szHost = "127.0.0.1:8972"
	//}
	//szTo, _ := os.LookupEnv("TOTAL")
	//if szTo != "" {
	//	nTotal, _ = strconv.Atoi(szTo)
	//} else {
	//	nTotal = 1
	//}
	//szConc, _ := os.LookupEnv("CONC")
	//if szConc != "" {
	//	nConcurrency, _ = strconv.Atoi(szConc)
	//} else {
	//	nConcurrency = 1
	//}

	flag.StringVar(&szHost, "host", "127.0.0.1", "主机名，默认为127.0.0.1")
	flag.IntVar(&nPort, "port", 8800, "端口，默认8800")
	flag.IntVar(&nTotal, "total", 1, "测试消息总数，默认1")
	flag.IntVar(&nConcurrency, "concurrency", 1, "并发数，默认1")
	flag.IntVar(&nInitConn, "initConn", 1, "初始连接数，默认1")
	flag.IntVar(&nMaxConn, "maxConn", 3, "最大连接数，默认3")
	flag.IntVar(&nGoodByeTimeoutS, "goodByeTimeout", 60, "agent goodbye超时时间")
	flag.StringVar(&szSecretKey, "secretKey", "node", "秘钥，默认node")
	flag.Parse()


	var objTaskWorkerCfg taskworker.WorkerConfigs
	var aryAgentHostInfo []client.AgentInfo

	if config.ServerConf.HasConfig {
		//szHost = config.ServerConf.Host
		nTotal = config.ServerConf.Total
		nConcurrency = config.ServerConf.Concurrency
		objTaskWorkerCfg.Workers = config.ServerConf.NetWorkers
		szSecretKey = config.ServerConf.SecretKey
		for _, objAgent := range config.ServerConf.Agents {
			aryAgentHostInfo = append(aryAgentHostInfo, client.AgentInfo{
				IP:   objAgent.IP,
				Port: objAgent.Port,
			})
		}
		nInitConn = config.ServerConf.InitConn
		nMaxConn = config.ServerConf.MaxConn
		nGoodByeTimeoutS = config.ServerConf.GoodByeTimeoutS
	} else {
		objTaskWorkerCfg.Workers = append(objTaskWorkerCfg.Workers, taskworker.WorkerConfig{
			// name="network_task_flag" size="100" queue="100"
			Name:  "network_task_flag",
			Size:  nMaxConn,
		})
		aryAgentHostInfo = []client.AgentInfo{{szHost, uint16(nPort)}}
	}
	nInitTotal = nTotal

	taskworker.TaskWorkerManagerInstance().Init(objTaskWorkerCfg.Workers)

	//for ; nConcurrency <= config.ServerConf.MaxConcurrency; nConcurrency += config.ServerConf.DeltaConcurrency {
	//	nTotal = nInitTotal
	//	for ; nTotal <= config.ServerConf.MaxTotal; nTotal += config.ServerConf.DeltaTotal {

			n := nConcurrency
			m := nTotal / n

			log.Infof("stat nConcurrency: %d", n)
			log.Infof("stat requests per client: %d", m)

			ptrArgs := prepareArgs()

			byteData, _ := proto.Marshal(ptrArgs)
			log.Infof("stat message size: %d bytes", len(byteData))

			var objWG sync.WaitGroup
			objWG.Add(n * m)

			var nTrans uint64
			var nTransOK uint64

			d := make([][]int64, n, n)

			//it contains warmup time but we can ignore it
			nTotalT := time.Now().UnixNano()
			for i := 0; i < n; i++ {
				dt := make([]int64, 0, m)
				d = append(d, dt)
				go func(i int) {
					chStop := make(chan string, 1)

					ptrClient, anyErr := client.NewClient(szSecretKey, aryAgentHostInfo, nInitConn, nMaxConn, nGoodByeTimeoutS)
					if anyErr != nil {
						log.Errorf("NewClient err:%v", anyErr)
						return
					}
					go func() {
						anyErr = ptrClient.Start(chStop)
						if anyErr != nil {
							log.Errorf("Client.Start err:%v", anyErr)
							return
						}
					}()

					for j := 0; j < m; j++ {
						t := time.Now().UnixNano()
						//ptrReply, anyErr := c.Say(context.Background(), ptrArgs)
						ptrReply := &pb.BenchmarkMessage{}
						anyErr := ptrClient.SyncCall("benchmark.test", ptrArgs, ptrReply, 3000)
						t = time.Now().UnixNano() - t

						d[i] = append(d[i], t)

						if anyErr == nil && *ptrReply.Field1 == "OK" {
							atomic.AddUint64(&nTransOK, 1)
						}

						atomic.AddUint64(&nTrans, 1)
						objWG.Done()
					}

					chStop <- "stop"

				}(i)

			}

			objWG.Wait()
			nTotalT = time.Now().UnixNano() - nTotalT
			nTotalT = nTotalT / 1000000
			log.Infof("stat took %d ms for %d requests", nTotalT, n*m)

			aryTotalD := make([]int64, 0, n*m)
			for _, k := range d {
				aryTotalD = append(aryTotalD, k...)
			}
			aryTotalD2 := make([]float64, 0, n*m)
			for _, k := range aryTotalD {
				aryTotalD2 = append(aryTotalD2, float64(k))
			}

			mean, _ := stats.Mean(aryTotalD2)
			median, _ := stats.Median(aryTotalD2)
			max, _ := stats.Max(aryTotalD2)
			min, _ := stats.Min(aryTotalD2)
			p99, _ := stats.Percentile(aryTotalD2, 99.9)

			log.Infof("stat sent     requests    : %d", n*m)
			log.Infof("stat concurrency num      : %d", n)
			log.Infof("stat received requests    : %d", atomic.LoadUint64(&nTrans))
			log.Infof("stat received requests_OK : %d", atomic.LoadUint64(&nTransOK))
			log.Infof("stat QPS                  : %d", int64(n*m)*1000/nTotalT)
			log.Infof("stat mean: %.f ns, median: %.f ns, max: %.f ns, min: %.f ns, p99: %.f ns", mean, median, max, min, p99)
			log.Infof("stat mean: %d ms, median: %d ms, max: %d ms, min: %d ms, p99: %d ms", int64(mean/1000000), int64(median/1000000), int64(max/1000000), int64(min/1000000), int64(p99/1000000))


			time.Sleep(time.Duration(20 * time.Second))
		//}
	//}

	if ! config.ServerConf.HasConfig {
		chProcStop := make(chan string, 1)
		// exit
		var anySignal= make(chan os.Signal)
		// 监听指定信号 ctrl+c kill
		signal.Notify(anySignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		for {
			select {
			case szStopReason := <-chProcStop:
				log.Infof("stop reason:%v", szStopReason)
				//fallthrough
				return
			case <-anySignal:
				return
			}
		}
	}
}

func prepareArgs() *pb.BenchmarkMessage {
	var b = true
	var i int32 = 100000
	var i64 int64 = 100000
	var s = "stringValue"

	var objArgs pb.BenchmarkMessage

	objVal := reflect.ValueOf(&objArgs).Elem()
	nNum := objVal.NumField()
	for k := 0; k < nNum; k++ {
		field := objVal.Field(k)
		if field.Type().Kind() == reflect.Ptr {
			switch objVal.Field(k).Type().Elem().Kind() {
			case reflect.Int, reflect.Int32:
				field.Set(reflect.ValueOf(&i))
			case reflect.Int64:
				field.Set(reflect.ValueOf(&i64))
			case reflect.Bool:
				field.Set(reflect.ValueOf(&b))
			case reflect.String:
				field.Set(reflect.ValueOf(&s))
			}
		} else {
			switch field.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				field.SetInt(100000)
			case reflect.Bool:
				field.SetBool(true)
			case reflect.String:
				field.SetString(s)
			}
		}

	}
	return &objArgs
}
