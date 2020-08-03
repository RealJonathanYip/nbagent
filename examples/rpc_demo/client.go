package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"git.yayafish.com/nbagent/examples/common"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"git.yayafish.com/nbagent/protocol/agent"
	"git.yayafish.com/nbagent/protocol/demo"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/rpc"
	"github.com/golang/protobuf/proto"
	_ "github.com/google/uuid"
)

type ClientConnectionHandler struct{}

func (this ClientConnectionHandler) OnRead(ptrConnection *network.ServerConnection,
	byteData []byte, ctx context.Context) {
	log.Infof("OnRead local: %v,data: %v", ptrConnection.LocalAddr(), len(byteData))

	var nMsgType uint8
	objReader := bytes.NewBuffer(byteData)
	if szUri, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("read uri fail!")
		return
	} else if anyErr := binary.Read(objReader, binary.LittleEndian, &nMsgType); anyErr != nil {
		log.Warningf("read nMsgType fail!")
		return
	} else if szRequestID, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("read szRequestID fail!")
		return
	} else {
		if nMsgType == protocol.MSG_TYPE_RESPONE {
			ptrServerConnection := network.GetServerConnection(ctx)
			ptrRpcContext, bSuccess := ptrServerConnection.GetRpcContext(szRequestID)
			if bSuccess {
				ptrRpcContext.Msg <- objReader.Bytes()
			} else {
				log.Warningf("can not find response context szRequestID:%v", szRequestID)
			}
			if szUri == rpc.NODE_RPC_RESP {
				objRpcCallResp := node.RpcCallResp{}
				if anyErr := proto.Unmarshal(objReader.Bytes(), &objRpcCallResp); anyErr != nil {
					log.Warningf("proto.Unmarshal error:%v", anyErr)
					return
				}
				log.Infof("RpcCallResp: %+v", objRpcCallResp)
			}
		} else if nMsgType == protocol.MSG_TYPE_REQUEST {

		} else if nMsgType == protocol.MSG_TYPE_CAST {

		}
	}
}

func (this ClientConnectionHandler) OnCloseConnection(ptrConnection *network.ServerConnection) {
	log.Infof("OnCloseConnection local: %v", ptrConnection.LocalAddr())
}

var (
	ptrAgentClient *network.ServerConnection = nil
)

func init() {
	log.SetLogLevel("info")
	log.InitLog(2, log.SetTarget("asyncfile"), log.SetEncode("json"),
		log.LogFilePath("./log/"+path.Base(os.Args[0])+"/"), log.LogFileRotate("date"))
}

func ClientAddAgent() {
	var bRet bool = false
	bRet, ptrAgentClient = network.NewClient(common.AgentIp, common.AgentPort, ClientConnectionHandler{})
	if !bRet {
		log.Errorf("network.NewClient error")
		return
	}
}

func ClientAgentRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCall *rpc.CallContext
	var objReq agent.AgentRegisterReq = agent.AgentRegisterReq{}
	objReq.InstanceID = fmt.Sprintf("%s_%d", common.AgentNameSrc, time.Now().UnixNano())
	objReq.TimeStamp = nTimeNow
	objReq.Sign = common.MakeSign(objReq.InstanceID, objReq.TimeStamp)
	var objRsp agent.AgentRegisterRsp = agent.AgentRegisterRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_REGISTER, &objReq, &objRsp, ptrAgentClient)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_REGISTER Req: %v, Rsp: %v", objReq, objRsp)
	}

	go func() {
		var objOnce sync.Once
		for {
			time.Sleep(1 * time.Second)
			ClientAgentKeepAlive()
			objOnce.Do(func() {
				time.Sleep(20 * time.Second)
				//ClienAgentUnRegister()
			})
		}
	}()
}

func TestRpcDemo() {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			AgentRpcDemo()
		}
	}()

}

func ClientLoop() {
	for {
		time.Sleep(3 * time.Second)
	}
}

func ClientAgentKeepAlive() {

	var bRet bool = false
	var objReq agent.AgentKeepAliveNotify = agent.AgentKeepAliveNotify{}
	var ptrCall *rpc.CallContext
	var objRsp agent.AgentKeepAliveRsp = agent.AgentKeepAliveRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_KEEP_ALIVE, &objReq, &objRsp, ptrAgentClient)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_KEEP_ALIVE success, Rsp: %+v", objRsp)
	} else {
		log.Warningf("AGENT_KEEP_ALIVE error,Req: %v", objReq)
	}
}

func ClienAgentUnRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCast *rpc.CastContext
	var objReq agent.AgentUnRegisterReq = agent.AgentUnRegisterReq{}
	objReq.TimeStamp = nTimeNow
	objReq.Sign = common.MakeSign(objReq.InstanceID, objReq.TimeStamp)

	ptrCast = rpc.NewCast(rpc.AGENT_UN_REGISTER, &objReq, ptrAgentClient)
	bRet = ptrCast.Start()
	if bRet {
		log.Infof("AGENT_UN_REGISTER Req: %v", objReq)
	}
}

func AgentRpcDemo() {

	var objDemoReq demo.TestMsgReq = demo.TestMsgReq{}
	objDemoReq.TestNumber = 1
	objDemoReq.TestString = "1"
	byteData, _ := proto.Marshal(&objDemoReq)

	var bRet bool = false
	var ptrCall *rpc.CallContext
	var objReq node.RpcCallReq = node.RpcCallReq{}
	objReq.Data = byteData
	objReq.URI = common.RPC_URI_DEMO
	objReq.EntryType = node.EntryType_RPC
	objReq.Caller = []string{}
	objReq.RequestMode = node.RequestMode_LOCAL_BETTER
	objReq.Key = ""
	var objRsp node.RpcCallResp = node.RpcCallResp{}

	ptrCall = rpc.NewCall(rpc.NODE_RPC_REQ, &objReq, &objRsp, ptrAgentClient)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		var objDemoRsp demo.TestMsgRsp = demo.TestMsgRsp{}
		proto.Unmarshal(objRsp.Data, &objDemoRsp)
		log.Infof("NODE_RPC_REQ success Req: %+v, Rsp: %+v", objDemoReq, objDemoRsp.TestReply)
	} else {
		log.Warningf("NODE_RPC_REQ error,Req: %v", objReq)
	}
}

func main() {

	ClientAddAgent()
	ClientAgentRegister()
	TestRpcDemo()

	ClientLoop()
}
