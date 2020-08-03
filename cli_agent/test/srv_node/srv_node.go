package main

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	//"git.yayafish.com/nbagent/agent_handler"
	"git.yayafish.com/nbagent/cli_agent"
	"git.yayafish.com/nbagent/cli_agent/test"
	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	node_mgr "git.yayafish.com/nbagent/node"
	"git.yayafish.com/nbagent/protocol/agent"
	"git.yayafish.com/nbagent/rpc"
	"git.yayafish.com/nbagent/taskworker"
	_ "github.com/google/uuid"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var (
	AGENT_NAME_SRC  string = "agent_src"
	AGENT_NAME_DEST string = "agent_dest"

	RPC_URI_DEMO string = "rpc.demo"
)

var (
	ptrNodeManager   = node_mgr.NewManager("node_2", "127.0.0.1", 8901, 8801, []node_mgr.Neighbour{})
	ptrClientManager = cli_agent.NewClientManager("agent_mgr", "127.0.0.1", 8801)

	ptrAgentClient *network.ServerConnection = nil
)

func init() {
	log.SetLogLevel("info")

	rand.Seed(time.Now().UnixNano())
	taskworker.TaskWorkerManagerInstance().Init(config.ServerConf.WorkerConfigs.Workers)

	//agent_handler.RegisterAgentHandler(ptrClientManager.HandleRpcCallReqFromNode, ptrClientManager.HandleRpcCallRespFromNode)
	//agent_handler.RegisterNodeDispatch(ptrNodeManager.DispatchReq, ptrNodeManager.DispatchRsp)
	//agent_handler.RegisterHandleEntry(ptrNodeManager.AddEntry, ptrNodeManager.RemoveEntry)
	//agent_handler.RegisterNodeInfo(ptrNodeManager.NodeList)
	ptrNodeManager.RegisterAgentHandler(ptrClientManager)
	ptrClientManager.RegisterNodeHandler(ptrNodeManager)
}

func TestAddAgent() {
	var bRet bool = false
	bRet, ptrAgentClient = network.NewClient("127.0.0.1", 8801, test.ClientConnectionHandler{})
	if !bRet {
		log.Errorf("network.NewClient error")
		return
	}
}

func TestAgentRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCall *rpc.CallContext
	var objReq agent.AgentRegisterReq = agent.AgentRegisterReq{}
	objReq.InstanceID = fmt.Sprintf("%s_%d", AGENT_NAME_DEST, time.Now().UnixNano())
	objReq.TimeStamp = nTimeNow
	objReq.Sign = makeSign(objReq.InstanceID, objReq.TimeStamp)
	objReq.AryEntry = append(objReq.AryEntry, &agent.EntryInfo{URI: RPC_URI_DEMO, EntryType: agent.EntryType_RPC})
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
			AgentKeepAlive()
			objOnce.Do(func() {
				time.Sleep(20 * time.Second)
				//AgentUnRegister()
			})
		}
	}()
}

func TestLoop() {
	for {
		time.Sleep(3 * time.Second)
		log.Infof("AllEntry: %+v", ptrClientManager.GetAllEntry())
	}
}

func AgentKeepAlive() {

	var bRet bool = false
	var objReq agent.AgentKeepAliveNotify = agent.AgentKeepAliveNotify{}

	//var ptrCast *rpc.CastContext
	//ptrCast = rpc.NewCast(rpc.AGENT_KEEP_ALIVE, &objReq, ptrAgentClient)
	//bRet = ptrCast.Start()
	//if bRet {
	//	log.Infof("AGENT_KEEP_ALIVE")
	//}

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

func AgentUnRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCast *rpc.CastContext
	var objReq agent.AgentUnRegisterReq = agent.AgentUnRegisterReq{}
	objReq.InstanceID = fmt.Sprintf("%s_%d", AGENT_NAME_DEST, time.Now().UnixNano())
	objReq.TimeStamp = nTimeNow
	objReq.Sign = makeSign(objReq.InstanceID, objReq.TimeStamp)

	ptrCast = rpc.NewCast(rpc.AGENT_UN_REGISTER, &objReq, ptrAgentClient)
	bRet = ptrCast.Start()
	if bRet {
		log.Infof("AGENT_UN_REGISTER Req: %v", objReq)
	}
}

func makeSign(szID string, nNow int64) string {
	objMac := hmac.New(md5.New, []byte(config.ServerConf.SecretKey))
	objMac.Write([]byte(szID + strconv.FormatInt(nNow, 10)))

	return hex.EncodeToString(objMac.Sum(nil))
}

func main() {

	TestAddAgent()
	TestAgentRegister()

	TestLoop()
}
