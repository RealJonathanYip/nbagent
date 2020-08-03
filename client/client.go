package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"git.yayafish.com/nbagent/protocol/agent"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/router"
	"git.yayafish.com/nbagent/rpc"
	"git.yayafish.com/nbagent/taskworker"
	"github.com/google/uuid"
	"github.com/micro/protobuf/proto"
	"os"
	"sync"
	"time"
)

// TODO: 切换agent；getAgent nil
// TODO: 之后的平滑切换由于将会有goodbye协议，需要将之前“将goodbye”的agent记录，避免又切回“将goodbye”agent，既然进行了记录，则需要在后续进行清除

type Client struct {
	// delete
	//ptrConnMgr *ConnectionManager

	// TODO: 以下移至client层
	szInstanceID string
	szSecretKey string

	ptrRouter *router.Router

	// 公共协议协程池
	ptrCommonTaskWorkerMgr *taskworker.TaskWorkerManager
	// 业务协程池
	ptrTaskWorkerMgr *taskworker.TaskWorkerManager

	ptrRequestContext *map[string]*network.RpcContext
	ptrRequestRWLock *sync.Mutex

	// 下面的可能会动态变化，需要使用锁
	objRWLock sync.RWMutex

	// 业务路由
	aryEntry []*agent.EntryInfo
	mRouter map[string]func(byteData []byte, ctx context.Context) []byte

	ptrAgentManager *AgentManager
}

// TODO: del ip:port
// TODO: add route say goodbye
func NewClient(szSecretKey string, aryAgentInfo []AgentInfo, nInitConnectNum, nMaxConnectNum int, nAgentGoodByeTimeoutS int) (*Client, error) {
	var anyErr error

	// TODO: instanceID to string
	objUUID := uuid.New()
	szInstanceID := objUUID.String()

	// 公共协议协程
	var objCommonTaskWorkerCfg taskworker.WorkerConfigs
	objCommonTaskWorkerCfg.Workers = append(objCommonTaskWorkerCfg.Workers, taskworker.WorkerConfig{
		// name="tcp_msg" size="100" queue="100"
		Name:  TASK_WORKER_TAG_TCP_MESSAGE,
		Size:  nMaxConnectNum,
	})
	ptrCommonTaskWorkerMgr := &taskworker.TaskWorkerManager{}
	ptrCommonTaskWorkerMgr.Init(objCommonTaskWorkerCfg.Workers)
	// 业务协程池
	ptrTaskWorkerMgr := &taskworker.TaskWorkerManager{}
	ptrTaskWorkerMgr.Init(objCommonTaskWorkerCfg.Workers)

	mRequestContext := make(map[string]*network.RpcContext)
	ptrRequestRWLock := &sync.Mutex{}

	ptrAgentManager := &AgentManager{
		mAgentConnectPool:     make(map[string]*AgentConnectPoolInfo),
		mAgentGoodByeInfo:     make(map[string]*AgentGoodByeInfo),
		nAgentGoodByeTimeoutS: nAgentGoodByeTimeoutS,
	}
	// 初始化各个agent的连接池
	for i := range aryAgentInfo {
		var ptrAgentConnectPool = &AgentConnectPoolInfo{}
		if anyErr = ptrAgentConnectPool.Init(aryAgentInfo[i], nInitConnectNum, nMaxConnectNum); anyErr != nil {
			log.Errorf("AgentConnectPool.Init err:%v, agent:%+v, maxConnectNum:%v", anyErr, aryAgentInfo[i], nMaxConnectNum)
			continue
		}

		ptrAgentManager.AddAgent(ptrAgentConnectPool)
	}
	if ptrAgentManager.getAgentSize() <= 0 {
		anyErr = fmt.Errorf("agent pool list is empty")
		log.Errorf(anyErr.Error())
		return nil, anyErr
	}

	ptrClient := &Client{
		szInstanceID:        szInstanceID,
		szSecretKey:         szSecretKey,
		ptrRouter:           nil, // 因为router创建后注册路由需要client的函数，所以先初始化client，后续对这个变量赋值
		ptrCommonTaskWorkerMgr: ptrCommonTaskWorkerMgr,
		ptrTaskWorkerMgr:    ptrTaskWorkerMgr,
		ptrRequestContext:   &mRequestContext,
		ptrRequestRWLock:    ptrRequestRWLock,
		objRWLock:           sync.RWMutex{},
		aryEntry:            nil, // 后续append，可以不new
		mRouter:             make(map[string]func(byteData []byte, ctx context.Context) []byte),
		ptrAgentManager:     ptrAgentManager,
	}

	// 公共协议路由
	ptrRouter := router.NewRouter()
	if ptrRouter == nil {
		anyErr = fmt.Errorf("client:%s create router is nil", ptrClient)
		log.Error(anyErr.Error())
		return nil, anyErr
	}
	if ! ptrRouter.RegisterRouter(rpc.NODE_RPC_REQ, ptrClient.handleRpcCallReq) {
		anyErr = fmt.Errorf("client:%v RegisterRouter:%v fail", ptrClient, rpc.NODE_RPC_REQ)
		return nil, anyErr
	}
	if ! ptrRouter.RegisterRouter(rpc.NODE_SAY_GOOBEY, ptrClient.handleSayGoodbyeNotify) {
		anyErr = fmt.Errorf("client:%v, RegisterRouter:%v fail", ptrClient, rpc.NODE_SAY_GOOBEY)
		return nil, anyErr
	}
	ptrClient.ptrRouter = ptrRouter

	return ptrClient, nil
}

// TODO: handle rpc handle http
func (ptrClient *Client) HandleRpc(szUri string, fnHandler func(byteData []byte, ctx context.Context) []byte) error {
	var anyErr error

	ptrClient.objRWLock.Lock()
	defer ptrClient.objRWLock.Unlock()

	ptrClient.aryEntry = append(ptrClient.aryEntry, &agent.EntryInfo{
		URI:       szUri,
		EntryType: agent.EntryType_RPC,
	})

	ptrClient.mRouter[szUri] = fnHandler

	return anyErr
}

//func (ptrClient *Client) HandleFunc(szUri string, nEntryType agent.EntryType, fnHandler func(byteData []byte, ctx context.Context) []byte) error {
//	var anyErr error
//
//	ptrClient.objRWLock.Lock()
//	defer ptrClient.objRWLock.Unlock()
//
//	ptrClient.aryEntry = append(ptrClient.aryEntry, &agent.EntryInfo{
//		URI:       szUri,
//		EntryType: nEntryType,
//	})
//
//	ptrClient.mRouter[szUri] = fnHandler
//
//	return anyErr
//}

func (ptrClient *Client) Start(chStopReason chan string) (anyErr error) {

	// 客户端启动时立即进行注册
	anyErr = ptrClient.createAgentIdleConnect()
	if anyErr != nil {
		log.Warningf("Client.createAgentIdleConnect err:%v", anyErr)
		ptrClient.unregisterSelf()
		return
	}

	// exit
	var anySignal = make(chan os.Signal)
	// 监听指定信号 ctrl+c kill
	//signal.Notify(anySignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ptrTicker := time.NewTicker(time.Second * 3) // keepalive
	defer ptrTicker.Stop()
	for {
		select {
		case <- ptrTicker.C:
			go ptrClient.timer()
		case szStopReason := <-chStopReason:
			log.Infof("client:%v, stop reason:%v", ptrClient, szStopReason)
			//fallthrough
			ptrClient.unregisterSelf()
			goto Exist
		case <- anySignal:
			ptrClient.unregisterSelf()
			goto Exist
		}
	}

Exist:
	return nil
}

func (ptrClient *Client) SyncCall(szUri string, ptrReq, ptrRsp proto.Message, nTimeoutMs int64) (anyErr error) {
	byteData, _ := proto.Marshal(ptrReq)

	var objAgentReq node.RpcCallReq
	objAgentReq.Data = byteData
	objAgentReq.URI = szUri
	objAgentReq.EntryType = node.EntryType_RPC
	objAgentReq.Caller = []string{}
	objAgentReq.RequestMode = node.RequestMode_LOCAL_BETTER
	//objReq.RequestMode = node.RequestMode_LOCAL_BETTER
	//objReq.RequestMode = node.RequestMode_DEFAULT
	objAgentReq.Key = ""
	var objAgentRsp node.RpcCallResp

	var ptrConnection *network.ServerConnection
	ptrConnection, anyErr = ptrClient.getConnect()
	if anyErr != nil {
		log.Warningf("Client.getConnect err:%v", anyErr)
		return
	}

	ptrCall := rpc.NewCall(rpc.NODE_RPC_REQ, &objAgentReq, &objAgentRsp, ptrConnection)
	if ptrCall == nil {
		anyErr = fmt.Errorf("rpc.NewCall is nil, uri:%v, req:%+v, rsp:%+v, conn:%+v",
			szUri, ptrReq, ptrRsp, ptrConnection)
		log.Warningf(anyErr.Error())
		return
	}
	if ! ptrCall.Timeout(nTimeoutMs).Start() {
		anyErr = fmt.Errorf("CallContext.Start fail, uri:%v, req:%+v, conn:%+v",
			szUri, ptrReq, ptrConnection)
		log.Warningf(anyErr.Error())
		ptrClient.clearConnect(ptrConnection)
		return
	}
	log.Debugf("agent Call success")
	if objAgentRsp.Result != node.ResultCode_SUCCESS {
		anyErr = fmt.Errorf("result not success, result:%v", objAgentRsp.Result)
		log.Warningf(anyErr.Error())
		return
	}
	log.Debugf("agent Call result success, rsp:%+v", objAgentRsp)
	if anyErr = proto.Unmarshal(objAgentRsp.Data, ptrRsp); anyErr != nil {
		log.Errorf("proto.Unmarshal err:%v", anyErr)
	}
	return
}

// 获取或建立连接
func (ptrClient *Client) getConnect() (*network.ServerConnection, error) {
	var anyErr error
	// 获取当前的agent，从agent中获取连接
	var ptrPool *AgentConnectPoolInfo
	ptrPool, anyErr = ptrClient.ptrAgentManager.GetActiveAgentConnectPool()
	if anyErr != nil {
		log.Warningf("AgentManager.GetActiveAgentConnectPool err:%v", anyErr)
		return nil, anyErr
	}

	ptrConnectionInfo := ptrPool.GetConnect()
	if ptrConnectionInfo == nil {
		// 立刻建立连接，注册，返回
		if ptrConnectionInfo, anyErr = ptrClient.createConnectAndRegister(ptrPool); anyErr != nil {
			log.Warningf("Client.createConnectAndRegister err:%v, agent:%+v", anyErr, ptrPool.AgentInfo)
			// 当前的agent已经不能建立连接或进行注册了，根据目前agent的连接数，考虑切换agent
			if ptrPool.GetConnectNum() <= 0 {
				ptrClient.ptrAgentManager.SwitchActiveAgent()
			}
			return nil, anyErr
		}
	}

	if ! ptrPool.IsFullConnect() {
		go func() {
			// create and register
			if _, anyErr := ptrClient.createConnectAndRegister(ptrPool); anyErr != nil {
				log.Warningf("Client.createConnectAndRegister err:%v, agent:%+v", anyErr, ptrPool.AgentInfo)
				return
			}
		}()
	}

	return ptrConnectionInfo.Connection, nil
}

func (ptrClient *Client) clearConnect(ptrConnection *network.ServerConnection) {
	aryAgentKey := ptrConnection.Ctx.Value(CONTEXT_KEY_NODE)
	if szAgentKey, bOK := aryAgentKey.(string); bOK {
		var ptrPool *AgentConnectPoolInfo
		var bExist bool
		if ptrPool, bExist = ptrClient.ptrAgentManager.GetAgent(szAgentKey); bExist {
			ptrPool.ClearConnect(ptrConnection)
		}
	} else {
		log.Warningf("aryAgentKey.(string) fail, aryAgentKey:%v", aryAgentKey)
	}
}

func (ptrClient *Client) handleRpcCallReq(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"

	objReq := node.RpcCallReq{}
	if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		return szResult
	}

	log.Infof("handleRpcCallReq objReq: uri: %v, type: %v,caller: %v", objReq.URI, objReq.EntryType, objReq.Caller)
	if fnHandler, bExist := ptrClient.mRouter[objReq.URI]; ! bExist {
		log.Warningf("router not found of uri:%v Type:%v", objReq.URI, objReq.EntryType)
	} else {
		// 业务用的协程池
		anyErr := ptrClient.ptrTaskWorkerMgr.GoTask(ctx, TASK_WORKER_TAG_TCP_MESSAGE, func(ctx context.Context) int {

			aryData := fnHandler(objReq.Data, ctx)

			var objAgentRpcRsp= node.RpcCallResp{}
			objAgentRpcRsp.Data = aryData
			objAgentRpcRsp.URI = objReq.URI
			objAgentRpcRsp.EntryType = objReq.EntryType
			objAgentRpcRsp.Caller = objReq.Caller

			if byteByte, bSuccess := protocol.NewProtocol(rpc.NODE_RPC_RESP,
				protocol.MSG_TYPE_RESPONE, network.GetRequestID(ctx), &objAgentRpcRsp); bSuccess {
				ptrServerConnection := network.GetServerConnection(ctx)
				log.Debugf("client:%v send_rsp:[uri:%v] [req_id:%v] for conn(%v)", ptrClient, objReq.URI, network.GetRequestID(ctx), ConnectToString(ptrServerConnection))
				ptrServerConnection.Write(byteByte)
			} else {
				szResult = "protocol_NewProtocol_fail"
				log.Warningf("protocol.NewProtocol fail")
			}

			log.Debugf("uri:%v result:%v", objReq.URI, szResult)

			return CODE_SUCCESS
		})
		if anyErr != nil {
			szResult = "client_TaskWorkerToTask_fail"
			log.Warningf("TaskWorker.GoTask for tcp_msg err:%v", anyErr)
		}
	}

	return szResult
}

func (ptrClient *Client) handleSayGoodbyeNotify(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"

	//objReq := node.SayGoodbyeNotify{}
	//if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
	//	szResult = "proto_Unmarshal_fail"
	//	return szResult
	//}

	ptrServerConnection := network.GetServerConnection(ctx)
	// 获取remote ip:port
	szAgentHost := ptrServerConnection.RemoteAddr()
	log.Debugf("client:%v, rec agent:%v goodbye notify", ptrClient, szAgentHost)

	ptrClient.ptrAgentManager.AddGoodByeAgent(szAgentHost)

	ptrClient.ptrAgentManager.SwitchToOtherAgent(szAgentHost)
	// TODO: create connect
	anyErr := ptrClient.createAgentIdleConnect()
	if anyErr != nil {
		log.Warningf("Client.createAgentIdleConnect err:%v", anyErr)
	}

	return szResult
}

func (ptrClient *Client) OnRead(ptrConnection *network.ServerConnection, byteData []byte, ctx context.Context) {
	log.Debugf("client:%v handler onRead:%v", ptrClient, ptrConnection)

	log.Infof("client:%v OnRead local: %v,data: %v", ptrClient, ptrConnection.LocalAddr(), len(byteData))

	var nMsgType uint8
	objReader := bytes.NewBuffer(byteData)
	if szUri, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("client:%v read uri fail!", ptrClient)
		return
	} else if anyErr := binary.Read(objReader, binary.LittleEndian, &nMsgType); anyErr != nil {
		log.Warningf("client:%v read nMsgType fail!", ptrClient)
		return
	} else if szRequestID, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("client:%v read szRequestID fail!", ptrClient)
		return
	} else {
		if nMsgType == protocol.MSG_TYPE_RESPONE {
			ptrServerConnection := network.GetServerConnection(ctx)
			log.Debugf("client:%v rec_rsp:[uri:%v] [req_id:%v] for conn(%v)", ptrClient, szUri, szRequestID, ptrServerConnection.RemoteAddr() + "->" + ptrServerConnection.LocalAddr())
			ptrRpcContext, bSuccess := ptrServerConnection.GetRpcContext(szRequestID)
			if bSuccess {
				ptrRpcContext.Msg <- objReader.Bytes()
			} else {
				log.Warningf("client:%v can not find response context szRequestID:%v", ptrClient, szRequestID)
			}
		} else if nMsgType == protocol.MSG_TYPE_REQUEST || nMsgType == protocol.MSG_TYPE_CAST {
			// 公共协议用的协程池
			anyErr = ptrClient.ptrCommonTaskWorkerMgr.GoTask(ctx, TASK_WORKER_TAG_TCP_MESSAGE, func(ctx context.Context) int {

				if ! ptrClient.ptrRouter.Process(byteData, ctx) {
					log.Warningf("client:%v Router.Process failed", ptrClient)
					ptrClient.clearConnect(ptrConnection)
				} else {
					// 因为时间目前没用，也为了避免为了更新时间频繁加锁，这里先不更新
					//ptrClient.ptrAgentManager.GetActiveAgentConnectPool().refreshKeepAliveTime()
				}

				return CODE_SUCCESS
			})
			if anyErr != nil {
				log.Warningf("Client.CommonTaskWorkerMgr.GoTask err:%v", anyErr)
			}
		} else {
			log.Errorf("client:%v can not handle msgType:%v", ptrClient, nMsgType)
		}
	}

}

func (ptrClient *Client) OnCloseConnection(ptrConnection *network.ServerConnection) {
	// TODO: 通过agent name clear
	ptrClient.clearConnect(ptrConnection)
	return
}

// 创建新连接，并注册
func (ptrClient *Client) createConnectAndRegister(ptrAgent *AgentConnectPoolInfo) (*ConnectionInfo, error) {
	log.Debugf("client:%v new connect(%v:%v) creating", ptrClient, ptrAgent.IP, ptrAgent.Port)
	if ptrConnect, anyErr := ptrAgent.CreateAgentConnect(ptrClient, func(szIP string, nPort uint16, ptrServerConnection *network.ServerConnection) error {
		ptrServerConnection.SetRequestContext(ptrClient.ptrRequestContext)
		ptrServerConnection.SetRWLock(ptrClient.ptrRequestRWLock)
		return nil
	}); anyErr != nil {
		anyErr := fmt.Errorf("client:%v Agent.NewConnect err:%v, ip:%v, port:%v", ptrClient, anyErr, ptrAgent.IP, ptrAgent.Port)
		return nil, anyErr
	} else {
		if anyErr := ptrClient.register(ptrConnect.Connection); anyErr != nil {
			anyErr = fmt.Errorf("client:%v connect register err:%v", ptrClient, anyErr)
			ptrConnect.Connection.Close()
			return nil, anyErr
		}
		ptrAgent.AddConnect(ptrConnect)
		// TODO: set agent name to ctx
		ptrConnect.Connection.Ctx = context.WithValue(ptrConnect.Connection.Ctx, CONTEXT_KEY_NODE, makeHostKeyByIpPort(ptrAgent.IP, ptrAgent.Port))
		return ptrConnect, nil
	}
}

func (ptrClient *Client) createAgentIdleConnect() (anyErr error) {
	var ptrPool *AgentConnectPoolInfo
	ptrPool, anyErr = ptrClient.ptrAgentManager.GetActiveAgentConnectPool()
	if anyErr != nil {
		log.Warningf("AgentManager.GetActiveAgentConnectPool err:%v", anyErr)
		return anyErr
	}
	nCreatingCount := ptrPool.GetInitConnectNum() - ptrPool.GetConnectNum()
	for nCreatingCount > 0 {
		_, anyErr = ptrClient.createConnectAndRegister(ptrPool)
		if anyErr != nil {
			log.Errorf("Client.createConnectAndRegister err:%v, agent:%+v", anyErr, ptrPool.AgentInfo)
			ptrClient.unregisterSelf()
			return anyErr
		}
		nCreatingCount--
		if nCreatingCount <= 0 {
			break
		}
	}
	return anyErr
}

func (ptrClient *Client) agentKeepalive() {
	var ptrPool *AgentConnectPoolInfo
	var anyTmpErr error
	ptrPool, anyTmpErr = ptrClient.ptrAgentManager.GetActiveAgentConnectPool()
	if anyTmpErr != nil {
		log.Warningf("AgentManager.GetActiveAgentConnectPool err:%v", anyTmpErr)
		return
	}
	var ptrConnection *network.ServerConnection
	var anyErr error
	ptrConnection, anyErr = ptrClient.getConnect()
	if anyErr != nil {
		log.Warningf("Client.getConnect err:%v", anyErr)
		return
	}
	var aryAgent []*agent.AgentInfo
	aryAgent, anyErr = ptrClient.keepalive(ptrConnection)
	if anyErr != nil {
		log.Warningf("client:%v Client.keepalive err:%v", ptrClient, anyErr)
		return
	}
	var aryUpdateAgent []AgentInfo
	for i := range aryAgent {
		aryUpdateAgent = append(aryUpdateAgent, AgentInfo{
			//Name: aryAgent[i].Name,
			IP:   aryAgent[i].IP,
			Port: uint16(aryAgent[i].Port),
		})
	}
	ptrClient.ptrAgentManager.UpdateAgent(aryUpdateAgent, ptrPool.GetInitConnectNum(), ptrPool.GetMaxConnectNum())
}

// 针对连接层的注册
func (ptrClient *Client) register(ptrConnection *network.ServerConnection) error {
	var nNow = time.Now().Unix()
	var objReq = agent.AgentRegisterReq{
		InstanceID: ptrClient.szInstanceID,
		AryEntry:   ptrClient.aryEntry,
		Sign:       makeSign(ptrClient.szInstanceID, nNow, ptrClient.szSecretKey),
		TimeStamp:  nNow,
	}
	var objRsp agent.AgentRegisterRsp

	log.Debugf("register req:%+v", objReq)

	var anyErr error
	ptrCallContext := rpc.NewCall(rpc.AGENT_REGISTER, &objReq, &objRsp, ptrConnection).Context(context.Background())
	if ! ptrCallContext.Timeout(3000).Start() {
		//ptrConnectPool.ClearConnect(ptrCliConn.mConnection) // 连接由外部管理
		anyErr = fmt.Errorf("client:%v agent register req start failed, req:%+v", ptrClient, objReq)
		log.Warningf(anyErr.Error())
		return anyErr
	}
	log.Debugf("register rsp:%+v", objRsp)
	if objRsp.Result != agent.ResultCode_SUCCESS {
		anyErr = fmt.Errorf("client:%v agent register req failed, req:%+v, rsp:%+v", ptrClient, objReq, objRsp)
		log.Warningf(anyErr.Error())
		return anyErr
	}

	log.Debugf("client:%v register success, rsp:%+v", ptrClient, objRsp)

	return nil
}

// 针对连接层的反注册
func (ptrClient *Client) unregister(ptrConnection *network.ServerConnection) error {
	var nNow = time.Now().Unix()
	var objReq = agent.AgentUnRegisterReq{
		InstanceID: ptrClient.szInstanceID,
		AryEntry:   ptrClient.aryEntry,
		Sign:       makeSign(ptrClient.szInstanceID, nNow, ptrClient.szSecretKey),
		TimeStamp:  nNow,
	}

	log.Debugf("unregister req:%+v", objReq)

	var anyErr error
	ptrCallContext := rpc.NewCast(rpc.AGENT_UN_REGISTER, &objReq, ptrConnection).Context(context.Background())
	if ! ptrCallContext.Start() {
		//ptrConnectPool.ClearConnect(ptrCliConn.mConnection) // 连接由外部管理
		anyErr = fmt.Errorf("client:%v agent unregister req start failed, req:%+v", ptrClient, objReq)
		log.Errorf(anyErr.Error())
		return anyErr
	}

	log.Debugf("client:%v unregister success, req:%+v", ptrClient, objReq)

	return nil
}

// 心跳
func (ptrClient *Client) keepalive(ptrConnection *network.ServerConnection) ([]*agent.AgentInfo, error) {
	var objReq = agent.AgentKeepAliveNotify{
		AryEntry: ptrClient.aryEntry,
	}
	var objRsp agent.AgentKeepAliveRsp

	log.Debugf("keepalive req:%+v", objReq)

	var anyErr error
	ptrCallContext := rpc.NewCall(rpc.AGENT_KEEP_ALIVE, &objReq, &objRsp, ptrConnection).Context(context.Background())
	if ! ptrCallContext.Timeout(3000).Start() {
		//ptrConnectPool.ClearConnect(ptrCliConn.mConnection) // 连接由外部管理
		anyErr = fmt.Errorf("client:%v agent unregister req start failed, req:%+v", ptrClient, objReq)
		log.Errorf(anyErr.Error())
		return nil, anyErr
	}

	log.Debugf("client:%v keepalive success, rsp:%+v", ptrClient, objRsp)

	return objRsp.AryAgent, anyErr
}

// 反注册
func (ptrClient *Client) unregisterSelf() {

	var ptrConnection *network.ServerConnection
	var anyErr error
	ptrConnection, anyErr = ptrClient.getConnect()
	if anyErr != nil {
		log.Warningf("Client.getConnect err:%v", anyErr)
		return
	}

	if anyErr = ptrClient.unregister(ptrConnection); anyErr != nil {
		log.Warningf("Client.unregister err:%v", anyErr)
	}

	return
}

// 定时器
func (ptrClient *Client) timer() {
	log.Debugf("client:%v timer running...", ptrClient)
	ptrClient.agentKeepalive()
	// timer clear timeout goodbye agent
	ptrClient.ptrAgentManager.ClearGoodbyeTimeout()
}

func (ptrClient *Client) String() string {
	return fmt.Sprintf("Client:%v", ptrClient.szInstanceID)
}
