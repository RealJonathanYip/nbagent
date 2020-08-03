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

type ServerConnectionHandler struct{}

func (this ServerConnectionHandler) OnRead(ptrConnection *network.ServerConnection,
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
			switch szUri {
			case rpc.NODE_RPC_REQ:
				{
					objRpcCallReq := node.RpcCallReq{}
					if anyErr := proto.Unmarshal(objReader.Bytes(), &objRpcCallReq); anyErr != nil {
						log.Warningf("proto.Unmarshal error:%v", anyErr)
						return
					}
					log.Infof("objRpcCallReq: %+v", objRpcCallReq)

					if objRpcCallReq.URI == common.RPC_URI_DEMO {
						objTestMsgReq := demo.TestMsgReq{}
						if anyErr := proto.Unmarshal(objRpcCallReq.Data, &objTestMsgReq); anyErr != nil {
							log.Warningf("proto.Unmarshal error:%v", anyErr)
							return
						}
						log.Infof("TestMsgReq: %+v", objTestMsgReq)

						objTestMsgRsp := demo.TestMsgRsp{}
						objTestMsgRsp.TestReply = "Reply: " + objTestMsgReq.TestString
						objRpcCallResp := node.RpcCallResp{}
						objRpcCallResp.Result = node.ResultCode_SUCCESS
						objRpcCallResp.Data, _ = proto.Marshal(&objTestMsgRsp)
						objRpcCallResp.URI = objRpcCallReq.URI
						objRpcCallResp.EntryType = objRpcCallReq.EntryType
						objRpcCallResp.Caller = objRpcCallReq.Caller
						bRet := false
						byteResp, bRet := protocol.NewProtocol(rpc.NODE_RPC_RESP, protocol.MSG_TYPE_RESPONE,
							szRequestID, &objRpcCallResp)
						if !bRet {
							log.Warningf("protocol.NewProtocol error, data: %v", objRpcCallResp)
							return
						}
						bRet = ptrConnection.Write(byteResp)
						if !bRet {
							log.Warningf("Write error, data: %v", objRpcCallResp)
							return
						}
						log.Infof("TestMsgRsp: %+v", objTestMsgRsp)
						log.Infof("objRpcCallResp: %+v", objRpcCallResp)
					}
				}
			}
		}
	}
}

func (this ServerConnectionHandler) OnCloseConnection(ptrConnection *network.ServerConnection) {
	log.Infof("OnCloseConnection local: %v", ptrConnection.LocalAddr())
}

var (
	ptrAgentServer *network.ServerConnection = nil
)

func init() {
	log.SetLogLevel("info")
	log.InitLog(2, log.SetTarget("asyncfile"), log.SetEncode("json"),
		log.LogFilePath("./log/"+path.Base(os.Args[0])+"/"), log.LogFileRotate("date"))
}

func ServerAddAgent() {
	var bRet bool = false
	bRet, ptrAgentServer = network.NewClient(common.AgentIp, common.AgentPort, ServerConnectionHandler{})
	if !bRet {
		log.Errorf("network.NewClient error")
		return
	}
}

func ServerAgentRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCall *rpc.CallContext
	var objReq agent.AgentRegisterReq = agent.AgentRegisterReq{}
	objReq.InstanceID = fmt.Sprintf("%s_%d", common.AgentNameDest, time.Now().UnixNano())
	objReq.TimeStamp = nTimeNow
	objReq.Sign = common.MakeSign(objReq.InstanceID, objReq.TimeStamp)
	objReq.AryEntry = append(objReq.AryEntry, &agent.EntryInfo{URI: common.RPC_URI_DEMO, EntryType: agent.EntryType_RPC})
	var objRsp agent.AgentRegisterRsp = agent.AgentRegisterRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_REGISTER, &objReq, &objRsp, ptrAgentServer)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_REGISTER Req: %v, Rsp: %v", objReq, objRsp)
	}

	go func() {
		var objOnce sync.Once
		for {
			time.Sleep(1 * time.Second)
			ServerAgentKeepAlive()
			objOnce.Do(func() {
				time.Sleep(20 * time.Second)
				//ServerAgentUnRegister()
			})
		}
	}()
}

func ServerLoop() {
	for {
		time.Sleep(3 * time.Second)
	}
}

func ServerAgentKeepAlive() {

	var bRet bool = false
	var objReq agent.AgentKeepAliveNotify = agent.AgentKeepAliveNotify{}
	var ptrCall *rpc.CallContext
	var objRsp agent.AgentKeepAliveRsp = agent.AgentKeepAliveRsp{}
	ptrCall = rpc.NewCall(rpc.AGENT_KEEP_ALIVE, &objReq, &objRsp, ptrAgentServer)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_KEEP_ALIVE success, Rsp: %+v", objRsp)
	} else {
		log.Warningf("AGENT_KEEP_ALIVE error,Req: %v", objReq)
	}
}

func ServerAgentUnRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCast *rpc.CastContext
	var objReq agent.AgentUnRegisterReq = agent.AgentUnRegisterReq{}
	objReq.TimeStamp = nTimeNow
	objReq.Sign = common.MakeSign(objReq.InstanceID, objReq.TimeStamp)

	ptrCast = rpc.NewCast(rpc.AGENT_UN_REGISTER, &objReq, ptrAgentServer)
	bRet = ptrCast.Start()
	if bRet {
		log.Infof("AGENT_UN_REGISTER Req: %v", objReq)
	}
}

func main() {

	ServerAddAgent()
	ServerAgentRegister()

	ServerLoop()
}
