package cli_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol/agent"
	"git.yayafish.com/nbagent/protocol/caller"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/reporter"
	"git.yayafish.com/nbagent/router"
	"git.yayafish.com/nbagent/rpc"
	"git.yayafish.com/nbagent/taskworker"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

const (
	CONTEXT_KEY_AGENT      = "CONTEXT_KEY_AGENT"
	NODE_TIMEOUT           = 15
	MANAGER_STATUS_STARTED = 1
	MANAGER_STATUS_STOPED  = 2
)

//todo func name agent-->client //done
type ClientManager struct {
	szName                string
	mTempConnection       map[*network.ServerConnection]int64
	objTempConnectionLock sync.Mutex
	ptrClientRouter       *router.Router
	mClientInfo           map[string]*ClientInfo
	objLock4ClientInfoMap sync.RWMutex
	mUri2Client           map[string][]*ClientInfo
	objLock4UriMap        sync.RWMutex
	ctx                   context.Context
	nStatus               int32
	ptrServer             *network.Server
	objNodeHandler        NodeHandler
}

func NewClientManager(szName, szIP string, nPort uint16) *ClientManager {

	if len(szName) == 0 {
		objID := uuid.New()
		szName = "agent_" + objID.String()
	}

	ptrRouter := router.NewRouter()
	ptrManager := &ClientManager{
		szName:                szName,
		mTempConnection:       make(map[*network.ServerConnection]int64),
		objTempConnectionLock: sync.Mutex{},
		ptrClientRouter:       ptrRouter,
		mClientInfo:           make(map[string]*ClientInfo),
		objLock4ClientInfoMap: sync.RWMutex{},
		mUri2Client:           make(map[string][]*ClientInfo),
		objLock4UriMap:        sync.RWMutex{},
		ctx:                   context.TODO(),
		nStatus:               MANAGER_STATUS_STARTED,
		ptrServer:             nil,
		objNodeHandler:        &DummyNodeHandler{},
	}

	ptrRouter.RegisterRouter(rpc.AGENT_REGISTER, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleAgentRegister(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.AGENT_UN_REGISTER, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleAgentUnRegister(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.AGENT_KEEP_ALIVE, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleKeepAlive(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_RPC_REQ, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleRpcCallReq(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_RPC_RESP, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleRpcCallResp(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.AGENT_ENTRY_CHECK, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleEntryCheckReq(byteData, ctx)
	})

	go func() {
		ptrServer := network.NewServer(szIP, nPort, false, ptrManager, ptrManager)
		ptrManager.ptrServer = ptrServer
		go ptrServer.Start()

		//定时检查心跳，清理心跳超时的 client
		for {
			if atomic.LoadInt32(&ptrManager.nStatus) != MANAGER_STATUS_STARTED {
				break
			}
			ptrManager.cleanTimeout()
			time.Sleep(time.Duration(config.ServerConf.TimeoutCheckInterval) * time.Second) //todo interval //done
		}
	}()

	return ptrManager
}

func (ptrClientManager *ClientManager) RegisterNodeHandler(objNodeHandler NodeHandler) {
	ptrClientManager.objNodeHandler = objNodeHandler
}

func (ptrClientManager *ClientManager) OnStart(ptrServer *network.Server) {
	log.Infof("server started ...")
}

func (ptrClientManager *ClientManager) OnStop(ptrServer *network.Server) {
}

func (ptrClientManager *ClientManager) OnNewConnection(ptrServerConnection *network.ServerConnection) {

	if atomic.LoadInt32(&ptrClientManager.nStatus) != MANAGER_STATUS_STARTED {
		log.Warningf("agent manager stoped,refuse new connection")
		taskworker.TaskWorkerManagerInstance().GoTask(context.TODO(), TASK_WORKER_TAG_AGENT_MANAGER, func(ctx context.Context) int {
			ptrServerConnection.Close()
			return CODE_SUCCESS
		})
	}

	ptrClientManager.objTempConnectionLock.Lock()
	defer ptrClientManager.objTempConnectionLock.Unlock()

	ptrClientManager.mTempConnection[ptrServerConnection] = time.Now().Unix()
}

func (ptrClientManager *ClientManager) OnRead(ptrServerConnection *network.ServerConnection, aryBuffer []byte, ctx context.Context) {

	taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_TCP_MESSAGE, func(ctx context.Context) int {
		if !ptrClientManager.ptrClientRouter.Process(aryBuffer, ctx) {
			ptrServerConnection.Close()
		} else if ptrClient := getClient(ctx); ptrClient != nil {
			ptrClient.refreshKeepLiveTime()
		}

		return CODE_SUCCESS
	})
}

func (ptrClientManager *ClientManager) OnCloseConnection(ptrServerConnection *network.ServerConnection) {

	ptrClient := getClientByConnection(ptrServerConnection)
	if ptrClient != nil {
		log.Infof("client: %v disconnected", ptrClient.getClientInfo())
		nLeft := ptrClient.onCloseConnection(ptrServerConnection)
		if nLeft <= 0 {
			ptrClientManager.deleteClient(ptrClient)
		}
	} else {
		ptrClientManager.objTempConnectionLock.Lock()
		delete(ptrClientManager.mTempConnection, ptrServerConnection)
		ptrClientManager.objTempConnectionLock.Unlock()
	}
}

func getClientByConnection(ptrServerConnection *network.ServerConnection) *ClientInfo {
	ptrClient := ptrServerConnection.Ctx.Value(CONTEXT_KEY_AGENT)
	if ptrClient == nil {
		return nil
	} else if reflect.TypeOf(ptrClient).String() != "*cli_agent.ClientInfo" {
		return nil
	} else {
		return ptrClient.(*ClientInfo)
	}
}

func getClient(ctx context.Context) *ClientInfo {
	if ptrServerConnection := network.GetServerConnection(ctx); ptrServerConnection == nil {
		return nil
	} else {
		return getClientByConnection(ptrServerConnection)
	}
}

func (ptrClientManager *ClientManager) cleanTimeout() {
	nNow := time.Now().Unix()

	ptrClientManager.objTempConnectionLock.Lock()
	var aryTimeoutConnection []*network.ServerConnection
	for ptrServerConnection, nTime := range ptrClientManager.mTempConnection {
		if nNow > nTime+config.ServerConf.ConnectTimeout { //todo config //done
			aryTimeoutConnection = append(aryTimeoutConnection, ptrServerConnection)
			delete(ptrClientManager.mTempConnection, ptrServerConnection)
		}
	}
	ptrClientManager.objTempConnectionLock.Unlock()

	for _, ptrServerConnection := range aryTimeoutConnection {
		log.Infof("connection of %v timeout", ptrServerConnection.RemoteAddr())
		ptrServerConnection.Close()
	}

	var aryTimeoutClient []*ClientInfo
	ptrClientManager.objLock4ClientInfoMap.RLock()
	for _, ptrClient := range ptrClientManager.mClientInfo { //todo ptrClient //done
		if ptrClient.timeout(config.ServerConf.ClientTimeout) {
			log.Infof("client time out:%v", ptrClient.szInstanceID)
			aryTimeoutClient = append(aryTimeoutClient, ptrClient)
		}
	}
	ptrClientManager.objLock4ClientInfoMap.RUnlock()

	if len(aryTimeoutClient) > 0 {
		for _, ptrClient := range aryTimeoutClient {
			ptrClientManager.deleteClient(ptrClient)
		}
	}
}

func (ptrClientManager *ClientManager) deleteClient(ptrClient *ClientInfo) {

	aryEntry := ptrClient.markDead()
	ptrClientManager.RemoveEntry(aryEntry, ptrClient)

	//todo modify //done
	var szClientKey string = ptrClient.szInstanceID
	ptrClientManager.objLock4ClientInfoMap.Lock() //todo modify lock //done
	delete(ptrClientManager.mClientInfo, szClientKey)
	ptrClientManager.objLock4ClientInfoMap.Unlock()
}

func getClientMapKey(szName, szIP string, nPort uint32, nInstanceID uint64) string {
	return fmt.Sprintf("%s_%s_%d_%d", szName, szIP, nPort, nInstanceID)
}

func (ptrClientManager *ClientManager) handleAgentRegister(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("handleAgentRegister")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	objNow := time.Now()
	objReq := agent.AgentRegisterReq{}
	objRsp := agent.AgentRegisterRsp{}
	if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}
	//验证签名
	if objReq.Sign != makeSign(objReq.InstanceID, objReq.TimeStamp) {
		objRsp.Result = agent.ResultCode_SIGN_VALIDATE_FAIL
		szResult = agent.ResultCode_SIGN_VALIDATE_FAIL.String()
		goto ExitReply
	}

	if objNow.Unix() > objReq.TimeStamp+config.ServerConf.SignTimeout {
		objRsp.Result = agent.ResultCode_SIGN_TIMEOUT
		szResult = agent.ResultCode_SIGN_TIMEOUT.String()
		goto ExitReply
	} else if atomic.LoadInt32(&ptrClientManager.nStatus) != MANAGER_STATUS_STARTED {
		szResult = "agent manager stoped"
		goto Exit
	} else {
		objRsp.Result = agent.ResultCode_SUCCESS
		objRsp.Sign = makeSign(ptrClientManager.szName, objNow.Unix())
		objRsp.TimeStamp = objNow.Unix()

		var ptrClient *ClientInfo
		var szClientKey string = objReq.InstanceID
		var bExist bool
		ptrServerConnection := network.GetServerConnection(ctx)

		ptrClientManager.objLock4ClientInfoMap.Lock()
		if ptrServerConnection != nil {
			ptrClient, bExist = ptrClientManager.mClientInfo[szClientKey]
			if bExist {
				ptrClient.addConnection(ptrServerConnection)
				ptrClient.syncEntry(objReq.AryEntry) //todo: 是否支持增量注册URI
				ptrServerConnection.Ctx = context.WithValue(ptrServerConnection.Ctx, CONTEXT_KEY_AGENT, ptrClient)
			} else {
				ptrClient = NewClient(objReq.InstanceID, ptrClientManager, ptrClientManager.ctx)
				ptrClient.addConnection(ptrServerConnection)
				ptrClient.syncEntry(objReq.AryEntry)
				ptrServerConnection.Ctx = context.WithValue(ptrServerConnection.Ctx, CONTEXT_KEY_AGENT, ptrClient)
				ptrClientManager.mClientInfo[szClientKey] = ptrClient
			}
		}
		ptrClientManager.objLock4ClientInfoMap.Unlock()

		ptrClientManager.objTempConnectionLock.Lock()
		delete(ptrClientManager.mTempConnection, ptrServerConnection)
		ptrClientManager.objTempConnectionLock.Unlock()

		ptrClientManager.AddEntry(objReq.AryEntry, ptrClient)
	}

ExitReply:
	rpc.Reply(&objRsp, ctx)
	log.Infof("handleAgentRegister szResult:%v objReq:%+v objRsp:%+v", szResult, objReq, objRsp)
Exit:
	return szResult
}

func (ptrClientManager *ClientManager) handleAgentUnRegister(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("handleAgentUnRegister")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	szAgentName := ""
	ptrClient := getClient(ctx)

	objNotify := agent.AgentUnRegisterReq{}
	if anyErr := proto.Unmarshal(byteData, &objNotify); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrClient == nil {
		szResult = "agent_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_AGENT_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	} else {
		szAgentName = ptrClient.getClientInfo()
		ptrClientManager.deleteClient(ptrClient)
	}
Exit:
	log.Infof("client: %v handleAgentUnRegister req :%+v", szAgentName, objNotify)
	return szResult
}

func (ptrClientManager *ClientManager) handleKeepAlive(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("handleKeepAlive")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	szClient := ""
	ptrClient := getClient(ctx)

	objNotify := agent.AgentKeepAliveNotify{}
	if anyErr := proto.Unmarshal(byteData, &objNotify); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrClient == nil {
		szResult = "agent_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_AGENT_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})

		goto Exit
	} else if ptrClient.getState() == AGENT_STATUS_DEAD {
		szClient = ptrClient.getClientInfo()
		szResult = "agent_dead"
		goto Exit
	} else { //todo refresh //done
		ptrClient.refreshKeepLiveTime()
		szClient = ptrClient.getClientInfo()
		//aryNode := agent_handler.NodeList()
		aryNode := ptrClientManager.objNodeHandler.NodeList()
		var objRsp agent.AgentKeepAliveRsp
		for _, ptrNode := range aryNode {
			objRsp.AryAgent = append(objRsp.AryAgent, &agent.AgentInfo{Name: ptrNode.Name, IP: ptrNode.IP, Port: ptrNode.AgentPort})
		}
		rpc.Reply(&objRsp, ctx)
	}

Exit:
	log.Infof("handleKeepAlive client:%v szResult:%+v objNotify:%+v ",
		szClient, szResult, objNotify)
	return szResult
}

//客户端直连agent投递的rpc协议
func (ptrClientManager *ClientManager) handleRpcCallReq(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("handleRpcCallReq")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	szClient := ""
	objRsp := node.RpcCallResp{}
	var ptrClient *ClientInfo
	var ptrDestClient *ClientInfo
	var nResultCode node.ResultCode
	var szCaller string

	objReq := node.RpcCallReq{}
	if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	ptrClient = getClient(ctx)
	if ptrClient == nil {
		szResult = "agent_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_AGENT_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	}

	szCaller = caller.GenAgentCaller(ptrClient.getClientInfo(), network.GetRequestID(ctx), getConnectionIDFromCtx(ctx))
	caller.PushStack(&objReq, szCaller)

	//nResultCode = agent_handler.DispatchReq(&objReq, ctx)
	nResultCode = ptrClientManager.objNodeHandler.DispatchReq(&objReq, ctx)
	log.Infof("DispatchReq URI: %v,EntryType: %v,ResultCode: %v", objReq.URI, objReq.EntryType.String(), nResultCode.String())
	switch nResultCode { //todo 代码调整 //done
	case node.ResultCode_CALL_YOUR_SELF:
		{
			//ResultCode_CALL_YOUR_SELF 根据方法路由，使用相应的 client
			ptrDestClient = ptrClientManager.getClient4RpcCallReq(&objReq)
			if ptrDestClient == nil {
				szResult = node.ResultCode_ENTRY_NOT_FOUND.String()
				objRsp.Result = node.ResultCode_ENTRY_NOT_FOUND
				rpc.Reply(&objRsp, ctx)
				goto Exit
			} else {
				szClient = ptrDestClient.getClientInfo()
				if !ptrDestClient.rpcCall(&objReq) {
					szResult = "agent call fail"
					goto Exit
				}
			}
		}
	case node.ResultCode_ENTRY_NOT_FOUND:
		{
			szResult = nResultCode.String()
			objRsp.Result = nResultCode
			rpc.Reply(&objRsp, ctx)
			goto Exit
		}
	case node.ResultCode_SUCCESS:
		{
			szResult = nResultCode.String()
			goto Exit
		}
	default:
		{
			szResult = nResultCode.String()
			objRsp.Result = nResultCode
			rpc.Reply(&objRsp, ctx)
			goto Exit
		}
	}

Exit:
	log.Infof("szClient:%v handleRpcCallReq objReq:%v, Result: %v", szClient, objReq.URI, szResult)
	return szResult
}

//客户端直连agent投递的rpc协议
func (ptrClientManager *ClientManager) handleRpcCallResp(byteData []byte, ctx context.Context) string {

	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("handleRpcCallResp")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	szClient := ""
	objCaller := caller.Caller{}
	var anyErr error
	var ptrClientReq *ClientInfo
	var nResultCode node.ResultCode
	var bCaller bool = false
	var szCaller string

	objResp := node.RpcCallResp{}
	if anyErr := proto.Unmarshal(byteData, &objResp); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}
	//TODO:nResultCode
	//nResultCode = agent_handler.DispatchRsp(&objResp, ctx)
	nResultCode = ptrClientManager.objNodeHandler.DispatchRsp(&objResp, ctx)
	log.Infof("DispatchRsp URI: %v,EntryType: %v,ResultCode: %v", objResp.URI, objResp.EntryType.String(), nResultCode.String())
	switch nResultCode {
	case node.ResultCode_SUCCESS:
		{
			//other node resp
			szResult = nResultCode.String()
			goto Exit
		}
	case node.ResultCode_STACK_EMPTY:
		{
			//caller stack error
			szResult = nResultCode.String()
			goto Exit
		}
	case node.ResultCode_NODE_NOT_EXIST:
		{
			szResult = nResultCode.String()
			goto Exit
		}
	case node.ResultCode_CALL_YOUR_SELF: //todo //done
		{
			goto LocalResp
		}
	default:
		{
			szResult = nResultCode.String()
			goto Exit
		}
	}

LocalResp:
	if ptrClient := getClient(ctx); ptrClient == nil {
		szResult = "agent_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_AGENT_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	}

	//根据Caller 调用链路，找到对应的 client 和连接
	szCaller, bCaller = caller.GetTopStack(&objResp)
	if !bCaller {
		log.Warningf("can not get top caller,resp: %v", objResp.String())
		szResult = "can not get top caller"
		goto Exit
	}
	szCaller, bCaller = caller.PopStack(&objResp)
	if !bCaller {
		log.Warningf("can not get top caller,resp: %v", objResp.String())
		szResult = "can not get top caller"
		goto Exit
	}

	anyErr = json.Unmarshal([]byte(szCaller), &objCaller)
	if anyErr != nil {
		log.Warningf("json.Unmarshal error: %v,data: %v", anyErr, szCaller)
		szResult = "json.Unmarshal error: " + anyErr.Error()
		goto Exit
	}
	if objCaller.Type == caller.Caller_Agent {
		ptrClientManager.objLock4ClientInfoMap.RLock() //todo read lock //done
		ptrClientReq, _ = ptrClientManager.mClientInfo[objCaller.AgentID]
		ptrClientManager.objLock4ClientInfoMap.RUnlock()
		if ptrClientReq != nil {
			szClient = objCaller.AgentID
			log.Infof("rpcResp AgentID: %v", szClient)
			if !ptrClientReq.rpcResp(&objResp, objCaller) {
				log.Warningf("rpcResp fail,AgentID: %v", szClient)
			}
			goto Exit
		} else {
			log.Errorf("can not find client: %v", objCaller.AgentID)
		}
	} else {
		log.Errorf("not agent caller: %+v", objCaller)
	}

Exit:
	log.Infof("handleRpcCallResp szClient:%v, Resp: %v,szResult: %v", szClient, objResp.URI, szResult)
	return szResult
}

func (ptrClientManager *ClientManager) handleEntryCheckReq(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("handleEntryCheckReq")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	szClient := ""
	objRsp := agent.AgentEntryCheckRsp{}
	var ptrClient *ClientInfo
	var nResultCode node.ResultCode

	objReq := agent.AgentEntryCheckReq{}
	if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	ptrClient = getClient(ctx)
	if ptrClient == nil {
		szResult = "agent_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_AGENT_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	}

	nResultCode = ptrClientManager.objNodeHandler.EntryCheck(objReq.URI, node.EntryType(objReq.EntryType))
	log.Infof("EntryCheck URI: %v,EntryType: %v,ResultCode: %v", objReq.URI, objReq.EntryType.String(), nResultCode.String())
	rpc.Reply(&objRsp, ctx)

Exit:
	log.Infof("szClient:%v handleEntryCheckReq objReq:%v, Result: %v", szClient, objReq.URI, szResult)
	return szResult
}

func string2HashCode(szInput string) int {
	nResult := int(crc32.ChecksumIEEE([]byte(szInput)))

	if nResult >= 0 {
		return nResult
	}

	if -nResult >= 0 {
		return -nResult
	}

	return 0
}

func getConnectionIDFromCtx(ctx context.Context) string {

	var szConnectionID string
	ptrServerConnection := network.GetServerConnection(ctx)
	if ptrServerConnection != nil {
		szConnectionID = fmt.Sprintf("%s_%s", ptrServerConnection.LocalAddr(), ptrServerConnection.RemoteAddr())
	}

	return szConnectionID
}

func getConnectionID(ptrConnection *network.ServerConnection) string {

	var szConnectionID string
	if ptrConnection != nil {
		szConnectionID = fmt.Sprintf("%s_%s", ptrConnection.LocalAddr(), ptrConnection.RemoteAddr())
	}

	return szConnectionID
}

func (ptrClientManager *ClientManager) CheckAgentUri(ptrEntry *agent.EntryInfo) bool {
	_, bExist := ptrClientManager.mUri2Client[getUri(ptrEntry)]
	return bExist
}

func (ptrClientManager *ClientManager) AddEntry(aryEntry []*agent.EntryInfo, ptrClient *ClientInfo) {
	ptrClientManager.objLock4UriMap.Lock()
	defer ptrClientManager.objLock4UriMap.Unlock()

	for _, ptrEntry := range aryEntry {
		szUri := getUri(ptrEntry)
		aryClient, bExist := ptrClientManager.mUri2Client[szUri]
		if bExist {
			bFind := false
			for _, ptrTmp := range aryClient {
				if ptrTmp == ptrClient {
					bFind = true
					break
				}
			}
			if !bFind {
				aryClient = append(aryClient, ptrClient)
			}
		} else {
			aryClient = []*ClientInfo{}
			aryClient = append(aryClient, ptrClient)

			//node add entry
			//agent_handler.NodeAddEntry(ptrEntry.URI, node.EntryType(ptrEntry.EntryType))
			ptrClientManager.objNodeHandler.AddEntry(ptrEntry.URI, node.EntryType(ptrEntry.EntryType))
		}
		ptrClientManager.mUri2Client[szUri] = aryClient
	}
	log.Infof("AddEntry mUri2Client: %v", len(ptrClientManager.mUri2Client))
}

func (ptrClientManager *ClientManager) RemoveEntry(aryEntry []*agent.EntryInfo, ptrClient *ClientInfo) {
	ptrClientManager.objLock4UriMap.Lock()
	defer ptrClientManager.objLock4UriMap.Unlock()

	for _, ptrEntry := range aryEntry {
		szUri := getUri(ptrEntry)
		aryNodeList := ptrClientManager.mUri2Client[szUri]
		var aryClientListNew []*ClientInfo
		for _, ptrClientTemp := range aryNodeList {
			if ptrClientTemp != ptrClient {
				aryClientListNew = append(aryClientListNew, ptrClientTemp)
			}
		}

		ptrClientManager.mUri2Client[szUri] = aryClientListNew
		if len(aryClientListNew) <= 0 {
			delete(ptrClientManager.mUri2Client, szUri)
			//agent_handler.NodeRemoveEntry(ptrEntry.URI, node.EntryType(ptrEntry.EntryType))
			ptrClientManager.objNodeHandler.RemoveEntry(ptrEntry.URI, node.EntryType(ptrEntry.EntryType))
			log.Infof("NodeRemoveEntry uri: %v", szUri)
		}
	}
}

func (ptrClientManager *ClientManager) getClient4RpcCallReq(ptrReq *node.RpcCallReq) *ClientInfo {
	ptrClientManager.objLock4UriMap.Lock()
	defer ptrClientManager.objLock4UriMap.Unlock()

	szUri := ptrReq.URI + "_" + ptrReq.EntryType.String()
	aryClient, bExist := ptrClientManager.mUri2Client[szUri]
	log.Infof("getClient4RpcCallReq uri: %v,agent num: %v", szUri, len(aryClient))
	if !bExist {
		return nil
	}
	if len(aryClient) == 0 {
		return nil
	}

	var ptrClient *ClientInfo
	if ptrReq.Key == "" {
		ptrClient = aryClient[rand.Intn(len(aryClient))]
	} else {
		ptrClient = aryClient[string2HashCode(ptrReq.Key)%len(aryClient)]
	}

	return ptrClient
}

func (ptrClientManager *ClientManager) GetAllEntry() []*agent.EntryInfo {
	ptrClientManager.objLock4ClientInfoMap.RLock()
	defer ptrClientManager.objLock4ClientInfoMap.RUnlock()

	var aryEntry []*agent.EntryInfo
	for _, objAgent := range ptrClientManager.mClientInfo {
		aryEntry = append(aryEntry, objAgent.getAllEntry()...)
	}

	return aryEntry
}

//平滑退出
//1.拒绝新的链接
//2.广播 Goodbye 的消息
//3.超时后stop
func (ptrClientManager *ClientManager) Stop() {
	log.Infof("ClientManager Stop...")

	atomic.StoreInt32(&ptrClientManager.nStatus, MANAGER_STATUS_STOPED)

	var objNotify node.SayGoodbyeNotify

	ptrClientManager.objLock4ClientInfoMap.RLock()
	defer ptrClientManager.objLock4ClientInfoMap.RUnlock()

	for _, ptrlientTemp := range ptrClientManager.mClientInfo {
		ptrlientTemp.cast(rpc.NODE_SAY_GOOBEY, &objNotify, context.TODO())
	}

	//ptrClientManager.ptrServer.Stop()
}

//来自其他node投递的rpc协议
func (ptrClientManager *ClientManager) HandleRequest(ptrReq *node.RpcCallReq, ctx context.Context) {
	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("HandleRequestFromNode")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	szClient := ""
	objRsp := node.RpcCallResp{}
	var ptrDestClient *ClientInfo

	//根据方法路由，使用相应的agent
	ptrDestClient = ptrClientManager.getClient4RpcCallReq(ptrReq)
	if ptrDestClient == nil {
		szResult = node.ResultCode_ENTRY_NOT_FOUND.String()
		objRsp.Result = node.ResultCode_ENTRY_NOT_FOUND
		rpc.Reply(&objRsp, ctx)
		goto Exit
	} else {
		szClient = ptrDestClient.getClientInfo()
		if !ptrDestClient.rpcCall(ptrReq) {
			szResult = "agent call fail"
			goto Exit
		}
	}

Exit:
	log.Infof("HandleRpcCallReqFromNode szClient: %v, objReq: %v,caller: %v,result: %v",
		szClient, ptrReq.URI, ptrReq.Caller, szResult)
	return
}

//来自其他node投递的rpc协议
func (ptrClientManager *ClientManager) HandleResponse(ptrResp *node.RpcCallResp, ctx context.Context) {
	szResult := "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("HandleResponseNode")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	szClient := ""
	var anyErr error
	var ptrClientReq *ClientInfo
	var objCaller caller.Caller
	var bCaller bool = false
	var szCaller string

	//根据 Caller 调用链路，找到对应的agent和连接
	_, bCaller = caller.GetTopStack(ptrResp)
	if !bCaller {
		log.Warningf("can not get top caller,resp: %v", ptrResp.String())
		szResult = "can not get top caller"
		goto Exit
	}
	szCaller, bCaller = caller.PopStack(ptrResp)
	if !bCaller {
		log.Warningf("can not get top caller,resp: %v", ptrResp.String())
		szResult = "can not get top caller"
		goto Exit
	}
	log.Infof("HandleRpcCallRespFromNode TopStack caller: %v", szCaller)

	anyErr = json.Unmarshal([]byte(szCaller), &objCaller)
	if anyErr != nil {
		log.Warningf("json.Unmarshal error: %v,data: %v", anyErr, szCaller)
		szResult = "json.Unmarshal error: " + anyErr.Error()
		goto Exit
	}
	if objCaller.Type == caller.Caller_Agent {
		ptrClientManager.objLock4ClientInfoMap.RLock() //todo read lock //done
		ptrClientReq, _ = ptrClientManager.mClientInfo[objCaller.AgentID]
		ptrClientManager.objLock4ClientInfoMap.RUnlock()
		if ptrClientReq != nil {
			szClient = objCaller.AgentID
			ptrClientReq.rpcResp(ptrResp, objCaller)
			goto Exit
		} else {
			//todo not find,log //done
			log.Errorf("can not find client: %v", objCaller.AgentID)
		}
	}

Exit:
	log.Infof("HandleRpcCallRespFromNode szClient:%v, objResp:%v, caller: %v, result: %v",
		szClient, ptrResp.URI, objCaller, szResult)
	return
}
