package rpc

import (
	"context"
	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol/demo"
	"git.yayafish.com/nbagent/router"
	"github.com/golang/protobuf/proto"
	"os"
	"path"
	"sync"
	"testing"
)

type Test struct {
	ptrRouter *router.Router
}

var _objHandler = Test{router.NewRouter()}

func (this *Test) OnStart(ptrServer *network.Server) {
	log.Infof("server started ...")
	this.ptrRouter.RegisterRouter("test", func(byteData []byte, ctx context.Context) string {
		objMsg := demo.TestMsgReq{}
		_ = proto.Unmarshal(byteData, &objMsg)

		if objMsg.TestNumber > 0 {
			Reply(&objMsg, ctx)
		}
		return "SUCCESS"
	})
}

func (this *Test) OnStop(ptrServer *network.Server) {
}

func (this *Test) OnNewConnection(ptrServerConnection *network.ServerConnection) {
}

func (this *Test) OnRead(ptrServerConnection *network.ServerConnection, aryBuffer []byte, ctx context.Context) {
	if !this.ptrRouter.Process(aryBuffer, ctx) {
		ptrServerConnection.Close()
	}
}

func (s Test) OnCloseConnection(ptrServerConnection *network.ServerConnection) {
}

func init() {
	_ = log.SetLogLevel(config.ServerConf.LogLevel)

	_ = log.InitLog(2, log.SetTarget(config.ServerConf.LogMode), log.SetEncode("json"),
		log.LogFilePath(config.ServerConf.LogPath+path.Base(os.Args[0])+"/"))

	ptrServer := network.NewServer("127.0.0.1", 18886, false, &_objHandler, &_objHandler)

	go ptrServer.Start()
}

func callTest(ptrTest *testing.T, ptrServerConnection *network.ServerConnection) {
	objMsg := demo.TestMsgReq{}
	objMsg.TestString = "sss"
	objMsg.TestNumber = 888
	objMsgRsp := demo.TestMsgReq{}

	NewCall("test", &objMsg, &objMsgRsp, ptrServerConnection).Start()

	if objMsgRsp.TestNumber != objMsg.TestNumber {
		ptrTest.Fatalf("rpc fail")
	}
}

func TestCallNormal(ptrTest *testing.T) {
	if bSuccess, ptrServerConnection := network.NewClient("127.0.0.1", 18886, &_objHandler); bSuccess {
		callTest(ptrTest, ptrServerConnection)
	} else {
		ptrTest.Fatalf("connect server fail")
	}
}

func TestStreet(ptrTest *testing.T) {
	ptrWG := &sync.WaitGroup{}
	ptrWG.Add(300 * 1000)
	for i := 0; i < 300; i++ {
		go func() {
			if bSuccess, ptrServerConnection := network.NewClient("127.0.0.1", 18886, &_objHandler); bSuccess {
				for i := 0; i < 1000; i++ {
					callTest(ptrTest, ptrServerConnection)
					ptrWG.Done()
				}
			} else {
				ptrTest.Fatalf("connect server fail")
				ptrWG.Add(-1000)
			}
		}()
	}
	ptrWG.Wait()
}
