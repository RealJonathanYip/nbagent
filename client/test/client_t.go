package main

import (
	"context"
	"git.yayafish.com/nbagent/client"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/protocol/demo"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/taskworker"
	"github.com/golang/protobuf/proto"
	"os"
	"path"
	"time"
)

func init()  {
	//_ = log.InitLog(2, log.SetTarget("stdout"), log.SetEncode("json"),
	//	log.LogFilePath("..\\log\\"+path.Base(os.Args[0])+"/"))
	_ = log.InitLog(2, log.SetTarget("asyncfile"), log.SetEncode("json"),
		log.LogFilePath(".\\log\\"+path.Base(os.Args[0])+"/"))

	_ = log.SetLogLevel("debug")

	log.Debugf("log debug")
	log.Infof("log info")
	log.Warningf("log warning")
	log.Errorf("log error")
}

func createClient(szName string, chStop chan string, fn func(ptrClient *client.Client) error) (*client.Client, error) {

	var objTaskWorkerCfg taskworker.WorkerConfigs
	objTaskWorkerCfg.Workers = append(objTaskWorkerCfg.Workers, taskworker.WorkerConfig{
		// name="network_task_flag" size="100" queue="100"
		Name:  "network_task_flag",
		Size:  3,
	})
	taskworker.TaskWorkerManagerInstance().Init(objTaskWorkerCfg.Workers)

	aryAgentHostInfo := []client.AgentInfo{{"127.0.0.1", 8800}}
	ptrClient, anyErr := client.NewClient("node", aryAgentHostInfo, 5, 10, 30)
	if anyErr != nil {
		log.Errorf("NewClient err:%v", anyErr)
		return nil, anyErr
	}

	if fn != nil {
		if anyErr = fn(ptrClient); anyErr != nil {
			log.Errorf("fn err:%v", anyErr)
			return nil, anyErr
		}
	}

	go func() {
		anyErr = ptrClient.Start(chStop)
		if anyErr != nil {
			log.Errorf("Client.Start err:%v", anyErr)
			return
		}
	}()

	return ptrClient, nil
}

//func Te1st_Client_Create(t *testing.T) {
func T1111()  {

	stopCh := make(chan string, 1)

	szClientName := "test_client"
	if _, anyErr := createClient(szClientName, stopCh, nil); anyErr != nil {
		//t.Fatalf("createClient:%v err:%v", szClientName, anyErr)
		log.Fatalf("createClient:%v err:%v", szClientName, anyErr)
	}
	time.Sleep(time.Second * 3)

	time.Sleep(time.Second * 3)
	stopCh <- "test stop"
	time.Sleep(time.Second * 3)
}

//func Test_Client_Rpc(t *testing.T) {
func main() {

	stopCh1 := make(chan string, 1)
	stopCh2 := make(chan string, 1)

	{
		szClientName := "test_client1"
		if _, anyErr := createClient(szClientName, stopCh1, func(ptrCli *client.Client) error {
			anyTmpErr := ptrCli.HandleRpc("test_client1.test", func(byteData []byte, ctx context.Context) []byte {
				objReq := demo.TestMsgReq{}
				_ = proto.Unmarshal(byteData, &objReq)
				log.Debugf("client1 req:%+v", objReq)

				objRsp := node.SayHelloRsp{
					Result:               0,
					Sign:                 "rsp_sign",
					TimeStamp:            time.Now().Unix(),
					NodeInfo:             &node.NodeInfo{
						Name:                 "test_client1",
						IP:                   "",
						Port:                 0,
						InstanceID:           123,
						AgentPort:            0,
					},
					DataVersion:          1,
				}

				log.Debugf("client1 rsp:%+v", objRsp)
				retByteData, _ := proto.Marshal(&objRsp)

				return retByteData
			})
			if anyTmpErr != nil {
				log.Errorf("handleFunc err:%v", anyTmpErr)
			}
			return anyTmpErr
		}); anyErr != nil {
			//t.Fatalf("createClient:%v err:%v", szClientName, anyErr)
			log.Fatalf("createClient:%v err:%v", szClientName, anyErr)
		}
	}

	time.Sleep(time.Second * 10)

	{
		szClientName := "test_client2"
		if ptrClient, anyErr := createClient(szClientName, stopCh2, nil); anyErr != nil {
			//t.Fatalf("createClient:%v err:%v", szClientName, anyErr)
			log.Fatalf("createClient:%v err:%v", szClientName, anyErr)
		} else {
			objReq := demo.TestMsgReq{
				TestNumber:           213,
				TestString:           "235434",
			}
			objRsp := node.SayHelloRsp{}
			log.Debugf("req:%+v", objReq)
			if anyErr = ptrClient.SyncCall("test_client1.test", &objReq, &objRsp, 3000); anyErr != nil {
				//t.Fatalf("Client.SyncCall err:%v, req:%+v", anyErr, objReq)
				log.Fatalf("Client.SyncCall err:%v, req:%+v", anyErr, objReq)
			}
			//t.Logf("Client.SyncCall success, rsp:%+v", objRsp)
			log.Infof("Client.SyncCall success, rsp:%+v", objRsp)
		}
	}

	//for {
	//	time.Sleep(time.Second)
	//}

	time.Sleep(time.Second * 9)
	stopCh1 <- "test client1 stop"
	stopCh2 <- "test client2 stop"
	time.Sleep(time.Second * 3)
}
