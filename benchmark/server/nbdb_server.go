package main

import (
	"context"
	"flag"
	"git.yayafish.com/nbagent/benchmark/pb"
	"git.yayafish.com/nbagent/benchmark/server/config"
	"git.yayafish.com/nbagent/client"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/taskworker"
	"github.com/golang/protobuf/proto"
	"os"
	"path"
	"runtime"
	"time"
)

var (
	szIP string
	nPort int
	nDelay int
	nInitConn int
	nMaxConn int
	nGoodByeTimeoutS int
	szSecretKey string
)

func init() {
	//if config.ServerConf.HasConfig {
	//	_ = log.InitLog(2, log.SetTarget(config.ServerConf.LogTarget), log.SetEncode(config.ServerConf.LogEncode),
	//		log.LogFilePath(config.ServerConf.LogPath+path.Base(os.Args[0])+"/"))
	//	_ = log.SetLogLevel("debug")
	//} else {
		_ = log.InitLog(2, log.SetTarget("asyncfile"), log.SetEncode("json"),
			log.LogFilePath("./log/"+path.Base(os.Args[0])+"/"),
			log.LogFileRotate("date"))
		_ = log.SetLogLevel("debug")
	//}
}

func main() {

	chStop := make(chan string, 1)

	//d, _ := os.LookupEnv("DELAY")
	//if d != "" {
	//	nDelay, _ = strconv.Atoi(d)
	//}

	flag.StringVar(&szIP, "host", "127.0.0.1", "主机名，默认为127.0.0.1")
	flag.IntVar(&nPort, "port", 8800, "端口，默认8800")
	flag.IntVar(&nDelay, "delay", 0, "消息回复时延，默认0")
	flag.IntVar(&nInitConn, "initConn", 1, "初始连接数，默认1")
	flag.IntVar(&nMaxConn, "maxConn", 3, "最大连接数，默认3")
	flag.IntVar(&nGoodByeTimeoutS, "goodByeTimeout", 30, "good bye超时时间")
	flag.StringVar(&szSecretKey, "secretKey", "node", "秘钥，默认node")
	flag.Parse()

	var objTaskWorkerCfg taskworker.WorkerConfigs

	if config.ServerConf.HasConfig {
		szIP = config.ServerConf.Host
		nPort = config.ServerConf.Port
		nDelay = config.ServerConf.Delay
		nInitConn = config.ServerConf.InitConn
		nMaxConn = config.ServerConf.MaxConn
		nGoodByeTimeoutS = config.ServerConf.GoodByeTimeoutS
		objTaskWorkerCfg.Workers = config.ServerConf.NetWorkers
		szSecretKey = config.ServerConf.SecretKey
	} else {
		objTaskWorkerCfg.Workers = append(objTaskWorkerCfg.Workers, taskworker.WorkerConfig{
			// name="network_task_flag" size="100" queue="100"
			Name:  "network_task_flag",
			Size:  nMaxConn,
		})
	}

	taskworker.TaskWorkerManagerInstance().Init(objTaskWorkerCfg.Workers)

	aryAgentHostInfo := []client.AgentInfo{{szIP, uint16(nPort)}}
	ptrClient, anyErr := client.NewClient(szSecretKey, aryAgentHostInfo, nInitConn, nMaxConn, nGoodByeTimeoutS)
	if anyErr != nil {
		log.Errorf("NewClient err:%v", anyErr)
		return
	}

	anyTmpErr := ptrClient.HandleRpc("benchmark.test", func(byteData []byte, ctx context.Context) []byte {
		objReq := pb.BenchmarkMessage{}
		_ = proto.Unmarshal(byteData, &objReq)

		s := "OK"
		var i int32 = 100
		objReq.Field1 = &s
		objReq.Field2 = &i
		if nDelay > 0 {
			time.Sleep(time.Duration(nDelay))
		} else {
			runtime.Gosched()
		}

		retByteData, _ := proto.Marshal(&objReq)

		return retByteData
	})
	if anyTmpErr != nil {
		log.Errorf("HandleRpc err:%v", anyTmpErr)
	}

	anyErr = ptrClient.Start(chStop)
	if anyErr != nil {
		log.Errorf("Client.Start err:%v", anyErr)
		return
	}

}
