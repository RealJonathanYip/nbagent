package cli_agent

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	node_mgr "git.yayafish.com/nbagent/node"
	"git.yayafish.com/nbagent/protocol"
	"git.yayafish.com/nbagent/protocol/agent"
	"git.yayafish.com/nbagent/protocol/demo"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/rpc"
	"git.yayafish.com/nbagent/taskworker"
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
			switch szUri {
			case rpc.NODE_RPC_REQ:
				{
					objRpcCallReq := node.RpcCallReq{}
					if anyErr := proto.Unmarshal(objReader.Bytes(), &objRpcCallReq); anyErr != nil {
						log.Warningf("proto.Unmarshal error:%v", anyErr)
						return
					}
					log.Infof("objRpcCallReq: %+v", objRpcCallReq)

					if objRpcCallReq.URI == RPC_URI_DEMO {
						objTestMsgReq := demo.TestMsgReq{}
						if anyErr := proto.Unmarshal(objRpcCallReq.Data, &objTestMsgReq); anyErr != nil {
							log.Warningf("proto.Unmarshal error:%v", anyErr)
							return
						}
						log.Infof("TestMsgReq: %+v", objTestMsgReq)

						objTestMsgRsp := demo.TestMsgRsp{}
						objTestMsgRsp.TestReply = "Reply " + objTestMsgReq.TestString
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
					}

				}

			}
		}
	}
}

func (this ClientConnectionHandler) OnCloseConnection(ptrConnection *network.ServerConnection) {
	log.Infof("OnCloseConnection local: %v", ptrConnection.LocalAddr())
}

var (
	ptrNode          *node_mgr.NodeManager
	ptrClientManager *ClientManager

	ptrAgentClientSrc  *network.ServerConnection = nil
	ptrAgentClientDest *network.ServerConnection = nil
)

var (
	AGENT_NAME_SRC  string = "agent_src"
	AGENT_NAME_DEST string = "agent_dest"

	RPC_URI_DEMO string = "rpc.demo"
)

var (
	InstanceID_Src  string = fmt.Sprintf("%s_%d", AGENT_NAME_SRC, time.Now().UnixNano())
	InstanceID_Dest string = fmt.Sprintf("%s_%d", AGENT_NAME_DEST, time.Now().UnixNano())
)

func init() {
	log.SetLogLevel("info")

	rand.Seed(time.Now().UnixNano())
	taskworker.TaskWorkerManagerInstance().Init(config.ServerConf.WorkerConfigs.Workers)

	ptrNode = node_mgr.NewManager("node_1", "127.0.0.1", 8900, 8800, []node_mgr.Neighbour{})
	ptrClientManager = NewClientManager("agent_mgr", "127.0.0.1", 8800)
	//agent_handler.RegisterAgentHandler(ptrClientManager.HandleRpcCallReqFromNode, ptrClientManager.HandleRpcCallRespFromNode)
	//agent_handler.RegisterNodeDispatch(ptrNode.DispatchReq, ptrNode.DispatchRsp)
	//agent_handler.RegisterHandleEntry(ptrNode.AddEntry, ptrNode.RemoveEntry)

	ptrNode.RegisterAgentHandler(ptrClientManager)
	ptrClientManager.RegisterNodeHandler(ptrNode)
}

func TestAddAgent(ptrTest *testing.T) {
	var bRet bool = false
	bRet, ptrAgentClientSrc = network.NewClient("127.0.0.1", 8800, ClientConnectionHandler{})
	if !bRet {
		log.Errorf("network.NewClient error")
		return
	}

	bRet, ptrAgentClientDest = network.NewClient("127.0.0.1", 8800, ClientConnectionHandler{})
	if !bRet {
		log.Errorf("network.NewClient error")
		return
	}
}

func TestAgentRegister(ptrTest *testing.T) {
	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCall *rpc.CallContext
	var objReq agent.AgentRegisterReq = agent.AgentRegisterReq{}
	objReq.InstanceID = InstanceID_Src
	objReq.TimeStamp = nTimeNow
	objReq.Sign = makeSign(objReq.InstanceID, objReq.TimeStamp)
	var objRsp agent.AgentRegisterRsp = agent.AgentRegisterRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_REGISTER, &objReq, &objRsp, ptrAgentClientSrc)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_REGISTER Req: %v, Rsp: %v", objReq, objRsp)
	}

	objReq = agent.AgentRegisterReq{}
	objReq.InstanceID = InstanceID_Dest
	objReq.TimeStamp = nTimeNow
	objReq.Sign = makeSign(objReq.InstanceID, objReq.TimeStamp)
	objReq.AryEntry = append(objReq.AryEntry, &agent.EntryInfo{URI: RPC_URI_DEMO, EntryType: agent.EntryType_RPC})
	objRsp = agent.AgentRegisterRsp{}
	ptrCall = rpc.NewCall(rpc.AGENT_REGISTER, &objReq, &objRsp, ptrAgentClientDest)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_REGISTER Req: %v, Rsp: %v", objReq, objRsp)
	}

	go func() {
		var objOnce sync.Once
		for {
			time.Sleep(1 * time.Second)
			AgentKeepAlive()
			objOnce.Do(func() {
				time.Sleep(20 * time.Second)
				//AgentUnRegister()
			})
		}
	}()
}

func TestRpcDemo(ptrTest *testing.T) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			AgentRpcDemo()
		}
	}()

}

func TestLoop(ptrTest *testing.T) {
	for {
		time.Sleep(3 * time.Second)
		log.Infof("AllEntry: %+v", ptrClientManager.GetAllEntry())
	}
}

func AgentKeepAlive() {

	var bRet bool = false
	var objReq agent.AgentKeepAliveNotify = agent.AgentKeepAliveNotify{}
	var ptrCast *rpc.CastContext

	ptrCast = rpc.NewCast(rpc.AGENT_KEEP_ALIVE, &objReq, ptrAgentClientSrc)
	bRet = ptrCast.Start()
	if bRet {
		log.Infof("AGENT_KEEP_ALIVE")
	}

	ptrCast = rpc.NewCast(rpc.AGENT_KEEP_ALIVE, &objReq, ptrAgentClientDest)
	bRet = ptrCast.Start()
	if bRet {
		log.Infof("AGENT_KEEP_ALIVE")
	}
}

func AgentUnRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCast *rpc.CastContext
	var objReq agent.AgentUnRegisterReq = agent.AgentUnRegisterReq{}
	objReq.InstanceID = InstanceID_Src
	objReq.TimeStamp = nTimeNow
	objReq.Sign = makeSign(objReq.InstanceID, objReq.TimeStamp)

	ptrCast = rpc.NewCast(rpc.AGENT_UN_REGISTER, &objReq, ptrAgentClientSrc)
	bRet = ptrCast.Start()
	if bRet {
		log.Infof("AGENT_UN_REGISTER Req: %v", objReq)
	}

	objReq = agent.AgentUnRegisterReq{}
	objReq.InstanceID = InstanceID_Dest
	objReq.TimeStamp = nTimeNow
	objReq.Sign = makeSign(objReq.InstanceID, objReq.TimeStamp)
	ptrCast = rpc.NewCast(rpc.AGENT_UN_REGISTER, &objReq, ptrAgentClientDest)
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
	objReq.URI = RPC_URI_DEMO
	objReq.EntryType = node.EntryType_RPC
	objReq.Caller = []string{}
	objReq.RequestMode = node.RequestMode_LOCAL_FORCE
	//objReq.RequestMode = node.RequestMode_LOCAL_BETTER
	//objReq.RequestMode = node.RequestMode_DEFAULT
	objReq.Key = ""
	var objRsp node.RpcCallResp = node.RpcCallResp{}

	ptrCall = rpc.NewCall(rpc.NODE_RPC_REQ, &objReq, &objRsp, ptrAgentClientSrc)
	ptrCall.Timeout(10 * 1000)
	bRet = ptrCall.Start()
	if bRet {
		var objDemoRsp demo.TestMsgRsp = demo.TestMsgRsp{}
		proto.Unmarshal(objRsp.Data, &objDemoRsp)
		log.Infof("NODE_RPC_REQ success Rsp: %+v", objDemoRsp.TestReply)
	} else {
		log.Warningf("NODE_RPC_REQ error,Req: %v", objReq)
	}
}
