package node

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"math"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol/caller"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/router"
	"git.yayafish.com/nbagent/rpc"
	"git.yayafish.com/nbagent/taskworker"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

const (
	CONTEXT_KEY_NODE       = "CONTEXT_KEY_NODE"
	CONNECTION_TIMEOUT     = 60
	SIGN_TIMNEOUT          = 5
	NODE_TIMEOUT           = 15
	MANAGER_STATUS_STARTED = 1
	MANAGER_STATUS_STOPED  = 2
	GOOSIP_SIZE            = 5
)

type NodeManager struct {
	mTempConnection       map[*network.ServerConnection]int64
	objTempConnectionLock sync.Mutex
	ptrNodeRouter         *router.Router
	mNodeInfo             map[string]*Info
	objLock4NodeInfoMap   sync.RWMutex
	mRpcFunc2Node         map[string][]*Info
	objRpcFunc2NodeMap    sync.RWMutex
	ptrSelf               *Info
	nSequence             uint64
	aryNeighbour          []Neighbour
	ctx                   context.Context
	nStatus               int32
	ptrServer             *network.Server
	objAgentHandler       AgentHandler
}

type Neighbour struct {
	szName     string
	szIP       string
	nPort      uint16
	nAgentPort uint16
}

func GetNeighbour(szName, szIP string, nPort, nAgentPort uint16) Neighbour {
	return Neighbour{szName, szIP, nPort, nAgentPort}
}

func NewManager(szName, szIP string, nPort, nAgentPort uint16, aryNeighbour []Neighbour) *NodeManager {

	objID := uuid.New()
	if szName == "" {
		szName = "node_" + objID.String()
	}

	ptrRouter := router.NewRouter()
	nInstance := uint64(time.Now().UnixNano())
	ptrManager := &NodeManager{
		mTempConnection:       make(map[*network.ServerConnection]int64),
		objTempConnectionLock: sync.Mutex{},
		ptrNodeRouter:         ptrRouter,
		mNodeInfo:             make(map[string]*Info),
		objLock4NodeInfoMap:   sync.RWMutex{},
		mRpcFunc2Node:         make(map[string][]*Info),
		objRpcFunc2NodeMap:    sync.RWMutex{},
		ptrSelf:               nil,
		nSequence:             0,
		aryNeighbour:          aryNeighbour,
		ctx: context.WithValue(context.TODO(), NODE_INFO_KEY,
			&node.NodeInfo{Name: szName, IP: szIP, Port: uint32(nPort), AgentPort: uint32(nAgentPort), InstanceID: nInstance}),
		nStatus:         MANAGER_STATUS_STARTED,
		ptrServer:       nil,
		objAgentHandler: &DummyAgentHandler{},
	}

	ptrManager.ptrSelf = newNode(szName, szIP, nInstance, nPort, nAgentPort, ptrManager, ptrManager.ctx)
	ptrManager.mNodeInfo[ptrManager.ptrSelf.szName] = ptrManager.ptrSelf

	ptrRouter.RegisterRouter(rpc.NODE_SAY_HELLO, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleSayHello(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_SAY_GOOBEY, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleSayGoodbyeNotify(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_GET_ENTRY, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleGetEntryReq(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_ENTRY_NOTIFY, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleNodeEntryNotify(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_KEEP_ALIVE, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleKeepAlive(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_RPC_REQ, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleRpcCallReq(byteData, ctx)
	})

	ptrRouter.RegisterRouter(rpc.NODE_RPC_RESP, func(byteData []byte, ctx context.Context) string {
		return ptrManager.handleRpcCallResp(byteData, ctx)
	})

	go func() {
		ptrServer := network.NewServer(szIP, nPort, false, ptrManager, ptrManager)
		ptrManager.ptrServer = ptrServer
		go ptrServer.Start()

		for {
			if atomic.LoadInt32(&ptrManager.nStatus) != MANAGER_STATUS_STARTED {
				break
			}
			ptrManager.cleanTimeout()
			time.Sleep(time.Minute)
		}
	}()

	go func() {
		for {
			if atomic.LoadInt32(&ptrManager.nStatus) != MANAGER_STATUS_STARTED {
				break
			}

			ptrManager.checkNeighbour()
			time.Sleep(time.Minute * 3)
		}
	}()

	go func() {
		for {
			if atomic.LoadInt32(&ptrManager.nStatus) != MANAGER_STATUS_STARTED {
				break
			}

			time.Sleep(time.Second * 2)
			ptrManager.keepalive()
		}
	}()

	return ptrManager
}

func (ptrNodeManager *NodeManager) RegisterAgentHandler(objHandler AgentHandler) {
	ptrNodeManager.objAgentHandler = objHandler
}

func (ptrNodeManager *NodeManager) OnStart(ptrServer *network.Server) {
	log.Infof("server started ...")
}

func (ptrNodeManager *NodeManager) OnStop(ptrServer *network.Server) {
}

func (ptrNodeManager *NodeManager) OnNewConnection(ptrServerConnection *network.ServerConnection) {
	ptrNodeManager.objTempConnectionLock.Lock()
	defer ptrNodeManager.objTempConnectionLock.Unlock()

	ptrNodeManager.mTempConnection[ptrServerConnection] = time.Now().Unix()
}

func (ptrNodeManager *NodeManager) OnRead(ptrServerConnection *network.ServerConnection, aryBuffer []byte, ctx context.Context) {
	taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_TCP_MESSAGE, func(ctx context.Context) int {

		if !ptrNodeManager.ptrNodeRouter.Process(aryBuffer, ctx) {
			ptrServerConnection.Close()
		} else if ptrNode := getNode(ctx); ptrNode != nil {
			ptrNode.refreshKeepLiveTime()
		}

		return CODE_SUCCESS
	})
}

func (ptrNodeManager *NodeManager) OnCloseConnection(ptrServerConnection *network.ServerConnection) {
	if ptrNode := getNodeByConnection(ptrServerConnection); ptrNode != nil {
		if nLeft := ptrNode.removeConnection(ptrServerConnection); nLeft == 0 && !ptrNode.alive() {
			ptrNodeManager.objLock4NodeInfoMap.Lock()
			defer ptrNodeManager.objLock4NodeInfoMap.Unlock()
			delete(ptrNodeManager.mNodeInfo, ptrNode.szName)
			log.Infof("all connection of offline node: %v disconnected remove it", ptrNode.szName)
		}
	} else {
		ptrNodeManager.objTempConnectionLock.Lock()
		defer ptrNodeManager.objTempConnectionLock.Unlock()
		delete(ptrNodeManager.mTempConnection, ptrServerConnection)
		// TODO: 日志如何调整？
		//log.Infof("node: %v disconnected", ptrNode.szName)
	}
}

func getNodeByConnection(ptrServerConnection *network.ServerConnection) *Info {
	if ptrNode := ptrServerConnection.Ctx.Value(CONTEXT_KEY_NODE); ptrNode == nil {
		return nil
	} else if reflect.TypeOf(ptrNode).String() != "*node.Info" {
		return nil
	} else {
		return ptrNode.(*Info)
	}
}

func getNode(ctx context.Context) *Info {
	if ptrServerConnection := network.GetServerConnection(ctx); ptrServerConnection == nil {
		return nil
	} else {
		return getNodeByConnection(ptrServerConnection)
	}
}

func (ptrNodeManager *NodeManager) keepalive() {
	objNotify := node.KeepAliveNotify{DataVersion: atomic.LoadUint64(&ptrNodeManager.nSequence)}

	ptrNodeManager.objLock4NodeInfoMap.RLock()
	defer ptrNodeManager.objLock4NodeInfoMap.RUnlock()

	var aryNodeInfoTemp, aryNodeInfo []*node.NodeInfo
	for _, ptrNode := range ptrNodeManager.mNodeInfo {
		aryNodeInfoTemp = append(aryNodeInfoTemp,
			&node.NodeInfo{Name: ptrNode.szName,
				IP:         ptrNode.szIP,
				Port:       uint32(ptrNode.nPort),
				InstanceID: ptrNode.nInstanceID,
				AgentPort:  uint32(ptrNode.nAgentPort),
			})
	}

	nSize := len(aryNodeInfoTemp)
	nCount := int(math.Min(GOOSIP_SIZE, float64(nSize)))
	nStart := rand.Intn(len(aryNodeInfoTemp))

	for ; nCount > 0; nCount-- {
		aryNodeInfo = append(aryNodeInfo, aryNodeInfoTemp[nStart%nSize])
		nStart++
	}
	objNotify.Neighbours = aryNodeInfo
	for _, ptrNode := range ptrNodeManager.mNodeInfo {
		if ptrNode.szName == ptrNodeManager.ptrSelf.szName {
			continue
		}

		bCast := ptrNode.cast(rpc.NODE_KEEP_ALIVE, &objNotify, context.TODO()).Start()
		if bCast {
			log.Infof("keepalive cast node name: %v", ptrNode.szName)
		} else {
			log.Warningf("keepalive cast fail,node name: %v", ptrNode.szName)
		}
	}

}

func (ptrNodeManager *NodeManager) checkNeighbour() {
	ptrNodeManager.objLock4NodeInfoMap.Lock()

	for _, objNeighbour := range ptrNodeManager.aryNeighbour {
		if _, bExist := ptrNodeManager.mNodeInfo[objNeighbour.szName]; !bExist {
			ptrNode := newNode(objNeighbour.szName, objNeighbour.szIP, 0,
				objNeighbour.nPort, objNeighbour.nAgentPort, ptrNodeManager, ptrNodeManager.ctx)
			ptrNodeManager.mNodeInfo[objNeighbour.szName] = ptrNode

			taskworker.TaskWorkerManagerInstance().GoTask(context.TODO(), TASK_WORKER_TAG_NODE_MANAGER, func(i context.Context) int {
				ptrNode.reconnect()
				return CODE_SUCCESS
			})

			log.Infof("add neighbour:%v", objNeighbour.szName)
		}
	}
	ptrNodeManager.objLock4NodeInfoMap.Unlock()
}

func (ptrNodeManager *NodeManager) cleanTimeout() {
	nNow := time.Now().Unix()

	ptrNodeManager.objTempConnectionLock.Lock()

	var aryTimeoutConnection []*network.ServerConnection
	for ptrServerConnection, nTime := range ptrNodeManager.mTempConnection {
		if nNow > nTime+CONNECTION_TIMEOUT {
			aryTimeoutConnection = append(aryTimeoutConnection, ptrServerConnection)
		}
	}

	ptrNodeManager.objTempConnectionLock.Unlock()

	for _, ptrServerConnection := range aryTimeoutConnection {
		log.Infof("connection of %v timeout", ptrServerConnection.RemoteAddr())
		ptrServerConnection.Close()
	}

	var aryTimeoutNode []*Info
	ptrNodeManager.objLock4NodeInfoMap.RLock()

	for _, ptrNode := range ptrNodeManager.mNodeInfo {
		if ptrNode != ptrNodeManager.ptrSelf && ptrNode.timeout(NODE_TIMEOUT) {
			log.Infof("node time out:%v", ptrNode.szName)
			aryTimeoutNode = append(aryTimeoutNode, ptrNode)
		}
	}

	ptrNodeManager.objLock4NodeInfoMap.RUnlock()

	if len(aryTimeoutNode) > 0 {
		for _, ptrNode := range aryTimeoutNode {
			ptrNodeManager.deleteNode(ptrNode)
		}
	}
}

func (ptrNodeManager *NodeManager) handleSayHello(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"

	objNow := time.Now()
	objReq := node.SayHelloReq{}
	objRsp := node.SayHelloRsp{}
	if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}
	//验证签名
	if objReq.Sign != makeSign(objReq.NodeInfo.Name, objReq.TimeStamp) {
		objRsp.Result = node.ResultCode_SIGN_VALIDATE_FAIL
		szResult = "ResultCode_SIGN_VALIDATE_FAIL"
		goto Exit_Reply
	}

	if objNow.Unix() > objReq.TimeStamp+SIGN_TIMNEOUT {
		objRsp.Result = node.ResultCode_SIGN_TIMEOUT
		szResult = "ResultCode_SIGN_TIMEOUT"
		goto Exit_Reply
	} else if atomic.LoadInt32(&ptrNodeManager.nStatus) != MANAGER_STATUS_STARTED {
		szResult = "node stoped"
		goto Exit
	} else {
		objRsp.Sign = makeSign(ptrNodeManager.ptrSelf.szName, objNow.Unix())
		objRsp.TimeStamp = objNow.Unix()

		ptrNodeManager.objLock4NodeInfoMap.Lock()

		ptrServerConnection := network.GetServerConnection(ctx)
		if ptrServerConnection != nil {
			if ptrNode, bExist := ptrNodeManager.mNodeInfo[objReq.NodeInfo.Name]; bExist {
				ptrNode.addConnection(ptrServerConnection)
				ptrServerConnection.Ctx = context.WithValue(ptrServerConnection.Ctx, CONTEXT_KEY_NODE, ptrNode)
				if ptrNode.getInstanceID() != objReq.NodeInfo.InstanceID {
					log.Infof("data instance change:%v-%+v", ptrNode.getInstanceID(), objReq.NodeInfo)
					ptrNode.setInstanceID(objReq.NodeInfo.InstanceID)
				}

				taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_INTRODUCE, func(ctx context.Context) int {
					ptrNodeManager.trySyncAll(ptrNode, objReq.DataVersion, ctx)
					return CODE_SUCCESS
				})
			} else {
				ptrNode := newNode(objReq.NodeInfo.Name, objReq.NodeInfo.IP, objReq.NodeInfo.InstanceID,
					uint16(objReq.NodeInfo.Port), uint16(objReq.NodeInfo.AgentPort), ptrNodeManager, ptrNodeManager.ctx)
				ptrNode.addConnection(ptrServerConnection)
				ptrServerConnection.Ctx = context.WithValue(ptrServerConnection.Ctx, CONTEXT_KEY_NODE, ptrNode)
				ptrNodeManager.mNodeInfo[objReq.NodeInfo.Name] = ptrNode

				taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_INTRODUCE, func(ctx context.Context) int {
					ptrNodeManager.trySyncAll(ptrNode, objReq.DataVersion, ctx)
					return CODE_SUCCESS
				})
			}
		}

		ptrNodeManager.objLock4NodeInfoMap.Unlock()

		objRsp.NodeInfo = MyInfo(ptrNodeManager.ctx)

		ptrNodeManager.objTempConnectionLock.Lock()

		delete(ptrNodeManager.mTempConnection, ptrServerConnection)

		ptrNodeManager.objTempConnectionLock.Unlock()
	}

Exit_Reply:
	rpc.Reply(&objRsp, ctx)
	log.Infof("handleSayHello szResult:%v objReq:%+v objRsp:%+v", szResult, objReq, objRsp)

	if szResult == "SUCCESS" {
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_INTRODUCE, func(ctx context.Context) int {
			time.Sleep(time.Second * 2)

			var objNotify = node.KeepAliveNotify{}
			var aryNodeInfoTemp, aryNodeInfo []*node.NodeInfo

			ptrNodeManager.objLock4NodeInfoMap.RLock()
			for _, ptrNode := range ptrNodeManager.mNodeInfo {
				aryNodeInfoTemp = append(aryNodeInfoTemp, &node.NodeInfo{Name: ptrNode.szName, IP: ptrNode.szIP, Port: uint32(ptrNode.nPort), InstanceID: ptrNode.nInstanceID})
			}

			nSize := len(aryNodeInfoTemp)
			nCount := int(math.Min(GOOSIP_SIZE, float64(nSize)))
			nStart := rand.Intn(len(aryNodeInfoTemp))

			for ; nCount > 0; nCount-- {
				aryNodeInfo = append(aryNodeInfo, aryNodeInfoTemp[nStart%nSize])
				nStart++
			}
			objNotify.Neighbours = aryNodeInfo
			ptrNodeManager.objLock4NodeInfoMap.RUnlock()

			if len(aryNodeInfo) == 0 {
				return CODE_SUCCESS
			}

			objNotify.Neighbours = aryNodeInfo
			objNotify.DataVersion = atomic.LoadUint64(&ptrNodeManager.ptrSelf.nDataVersion)

			ptrNode := getNode(ctx)

			ptrNode.cast(rpc.NODE_KEEP_ALIVE, &objNotify, ctx).Start()

			return CODE_SUCCESS
		})
	}
Exit:
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

//其他node投递过来的rpc协议回调
func (ptrNodeManager *NodeManager) handleRpcCallReq(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	szNodeName := ""

	objReq := node.RpcCallReq{}
	if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrNode := getNode(ctx); ptrNode == nil {
		szResult = "node_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	} else {
		szNodeName = ptrNode.szName
		//agent_handler.HandleRequest(&objReq, ctx)
		ptrNodeManager.objAgentHandler.HandleRequest(&objReq, ctx)
	}
Exit:
	log.Infof("szNodeName:%v ,handleRpcCallReq objReq: uri: %v, type: %v,caller: %v",
		szNodeName, objReq.URI, objReq.EntryType, objReq.Caller)
	return szResult
}

func (ptrNodeManager *NodeManager) handleRpcCallResp(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	szNodeName := ""

	objResp := node.RpcCallResp{}
	if anyErr := proto.Unmarshal(byteData, &objResp); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrNode := getNode(ctx); ptrNode == nil {
		szResult = "node_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	} else {
		szNodeName = ptrNode.szName
		//agent_handler.HandleResponse(&objResp, ctx)
		ptrNodeManager.objAgentHandler.HandleResponse(&objResp, ctx)
	}
Exit:
	log.Infof("szNodeName:%v handleRpcCallResp objResp:%v-%v", szNodeName, objResp.URI, objResp.Caller)
	return szResult
}

func (ptrNodeManager *NodeManager) handleSayGoodbyeNotify(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	szNodeName := ""

	objNotify := node.SayGoodbyeNotify{}
	if anyErr := proto.Unmarshal(byteData, &objNotify); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrNode := getNode(ctx); ptrNode == nil {
		szResult = "node_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	} else {
		szNodeName = ptrNode.szName

		aryEntry := ptrNode.markDead()

		ptrNodeManager.deleteEntry(ptrNode, aryEntry)
	}

Exit:
	log.Infof("szNodeName:%v handleSayGoodbyeNotify objNotify:%+v", szNodeName, objNotify)
	return szResult
}

func (ptrNodeManager *NodeManager) deleteEntry(ptrNode *Info, aryEntry []*node.EntryInfo) {
	ptrNodeManager.objRpcFunc2NodeMap.Lock()
	for _, ptrEntry := range aryEntry {
		aryNodeList := ptrNodeManager.mRpcFunc2Node[getUri(ptrEntry)]
		var aryNodeListNew []*Info
		for _, ptrNodeTemp := range aryNodeList {
			if ptrNodeTemp.szName != ptrNode.szName {
				aryNodeListNew = append(aryNodeListNew, ptrNodeTemp)
			}
		}

		ptrNodeManager.mRpcFunc2Node[getUri(ptrEntry)] = aryNodeListNew
	}
	ptrNodeManager.objRpcFunc2NodeMap.Unlock()
}

func (ptrNodeManager *NodeManager) deleteNode(ptrNode *Info) {
	aryEntry := ptrNode.markDead()

	ptrNodeManager.deleteEntry(ptrNode, aryEntry)

	ptrNodeManager.objLock4NodeInfoMap.RLock()
	defer ptrNodeManager.objLock4NodeInfoMap.RUnlock()
	delete(ptrNodeManager.mNodeInfo, ptrNode.szName)
}

func (ptrNodeManager *NodeManager) handleNodeEntryNotify(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	szNodeName := ""

	objNotify := node.NewEntryNotify{}
	if anyErr := proto.Unmarshal(byteData, &objNotify); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrNode := getNode(ctx); ptrNode == nil {
		szResult = "node_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	} else if ptrNode.getState() == NODE_STATUS_SYNCING || ptrNode.getState() == NODE_STATUS_DEAD {
		szNodeName = ptrNode.szName
		szResult = "node_dead_or_syncing"
		goto Exit
	} else if bNeedSyncAll, aryChangeEntry := ptrNode.modifyEntry(objNotify.EntryInfo, objNotify.DataVersion); !bNeedSyncAll && len(aryChangeEntry) > 0 {
		szNodeName = ptrNode.szName

		ptrNodeManager.objRpcFunc2NodeMap.Lock()
		defer ptrNodeManager.objRpcFunc2NodeMap.Unlock()

		for _, ptrEntry := range aryChangeEntry {
			szUri := getUri(ptrEntry)

			if ptrEntry.IsNew {
				ptrNodeManager.mRpcFunc2Node[szUri] = append(ptrNodeManager.mRpcFunc2Node[szUri], ptrNode)
				log.Infof("add entry:%v-%+v", ptrNode.szName, ptrEntry)
			} else {
				aryNode := ptrNodeManager.mRpcFunc2Node[szUri]

				var aryNodeListNew []*Info
				for _, ptrNodeTemp := range aryNode {
					if ptrNodeTemp != ptrNode {
						aryNodeListNew = append(aryNodeListNew, ptrNodeTemp)
					}
				}

				ptrNodeManager.mRpcFunc2Node[szUri] = aryNodeListNew
				log.Infof("remove entry:%v-%+v", ptrNode.szName, ptrEntry)
			}
		}
	}
Exit:
	log.Infof("handleNodeEntryNotify szNodeName:%v, szResult:%v, objNotify:%+v", szNodeName, szResult, objNotify)
	return szResult
}

func (ptrNodeManager *NodeManager) handleGetEntryReq(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	szNodeName := ""

	objReq := node.GetEntryReq{}
	objResp := node.GetEntryResp{}
	if anyErr := proto.Unmarshal(byteData, &objReq); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrNode := getNode(ctx); ptrNode == nil {
		szResult = "node_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	} else if atomic.LoadInt32(&ptrNodeManager.nStatus) != MANAGER_STATUS_STARTED {
		szResult = "node stoped"
		goto Exit
	} else {
		szNodeName = ptrNode.szName
		objResp.DataVersion = atomic.LoadUint64(&ptrNodeManager.ptrSelf.nDataVersion)
		objResp.EntryInfo = ptrNodeManager.ptrSelf.getAllEntry()
		goto Exit_Reply
	}

Exit_Reply:
	rpc.Reply(&objResp, ctx)
	log.Infof("handleGetEntryReq szNodeName:%v szResult:%v objReq:%+v objResp:%+v",
		szNodeName, szResult, objReq, objResp)

	return szResult

Exit:
	log.Infof("objReq:%+v", objReq)
	return szResult
}

func (ptrNodeManager *NodeManager) makefriends(aryNode []*node.NodeInfo, ctx context.Context) {
	//开一个协程做流言
	taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_INTRODUCE, func(ctx context.Context) int {
		//避免过多使用写锁
		var aryNodeInfo []*node.NodeInfo
		ptrNodeManager.objLock4NodeInfoMap.RLock()

		for _, objNodeInfo := range aryNode {
			if _, bExist := ptrNodeManager.mNodeInfo[objNodeInfo.Name]; !bExist {
				aryNodeInfo = append(aryNodeInfo, objNodeInfo)
			}
		}

		ptrNodeManager.objLock4NodeInfoMap.RUnlock()

		if len(aryNodeInfo) > 0 {
			ptrNodeManager.objLock4NodeInfoMap.Lock()
			defer ptrNodeManager.objLock4NodeInfoMap.Unlock()

			for _, objNodeInfo := range aryNodeInfo {
				if _, bExist := ptrNodeManager.mNodeInfo[objNodeInfo.Name]; !bExist {
					ptrTempNode := newNode(objNodeInfo.Name, objNodeInfo.IP, objNodeInfo.InstanceID,
						uint16(objNodeInfo.Port), uint16(objNodeInfo.AgentPort), ptrNodeManager, ptrNodeManager.ctx)
					log.Infof("find new node:%+v", objNodeInfo)
					ptrNodeManager.mNodeInfo[objNodeInfo.Name] = ptrTempNode

					taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
						ptrTempNode.reconnect()
						return CODE_SUCCESS
					})
				}
			}
		}

		return CODE_SUCCESS
	})
}

func (ptrNodeManager *NodeManager) trySyncAll(ptrNode *Info, nDataVersion uint64, ctx context.Context) {
	if ptrNode.needSyncAll(nDataVersion) {
		log.Infof("node:%v sync all data", ptrNode.szName)
		//如果发现数据版本号不连续 证明数据丢失 同步全部
		objReq := node.GetEntryReq{}
		objResp := node.GetEntryResp{}
		if ptrNode.call(rpc.NODE_GET_ENTRY, &objReq, &objResp, ctx).Start() {
			aryEntryOld := ptrNode.getAllEntry()
			ptrNode.syncEntry(objResp.EntryInfo, objResp.DataVersion)
			//先清空本地与node相关路由
			ptrNodeManager.deleteEntry(ptrNode, aryEntryOld)

			ptrNodeManager.objRpcFunc2NodeMap.Lock()
			defer ptrNodeManager.objRpcFunc2NodeMap.Unlock()

			//再重新添加
			for _, ptrEntry := range objResp.EntryInfo {
				szUri := getUri(ptrEntry)

				if ptrEntry.IsNew {
					aryNode := ptrNodeManager.mRpcFunc2Node[szUri]
					ptrNodeManager.mRpcFunc2Node[szUri] = append(aryNode, ptrNode)
					log.Infof("add entry:%v-%+v", ptrNode.szName, ptrEntry)
				} else {
					log.Warningf("sync all should not contain deleted entry:%v", szUri)
				}
			}
		} else {
			log.Warningf("sync call get entry fail! node:%v", ptrNode.szName)
		}
	}
}

func (ptrNodeManager *NodeManager) handleKeepAlive(byteData []byte, ctx context.Context) string {
	szResult := "SUCCESS"
	szNodeName := ""

	objNotify := node.KeepAliveNotify{}
	if anyErr := proto.Unmarshal(byteData, &objNotify); anyErr != nil {
		szResult = "proto_Unmarshal_fail"
		goto Exit
	}

	if ptrNode := getNode(ctx); ptrNode == nil {
		szResult = "node_not_auth"
		taskworker.TaskWorkerManagerInstance().GoTask(ctx, TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			network.GetServerConnection(ctx).Close()
			return CODE_SUCCESS
		})
		goto Exit
	} else if !ptrNode.alive() {
		szNodeName = ptrNode.szName
		szResult = "node_dead"
		goto Exit
	} else {
		szNodeName = ptrNode.szName
		ptrNodeManager.makefriends(objNotify.Neighbours, ctx)
		ptrNodeManager.trySyncAll(ptrNode, objNotify.DataVersion, ctx)
	}

Exit:
	log.Infof("handleKeepAlive szNodeName:%v szResult:%+v objNotify:%+v ", szNodeName, szResult, objNotify)
	return szResult
}

func (ptrNodeManager *NodeManager) RemoveEntry(szUri string, enumEntryType node.EntryType) {
	objEntry := node.EntryInfo{URI: szUri, EntryType: enumEntryType, IsNew: false}

	ptrNodeManager.objRpcFunc2NodeMap.Lock()

	var aryNewNodeInfo []*Info
	for _, ptrNode := range ptrNodeManager.mRpcFunc2Node[szUri] {
		if ptrNode.szName == ptrNodeManager.ptrSelf.szName {
			continue
		}

		aryNewNodeInfo = append(aryNewNodeInfo, ptrNode)
	}
	ptrNodeManager.mRpcFunc2Node[szUri] = aryNewNodeInfo
	log.Infof("RemoveEntry,uri: %v,left node: %v", szUri, len(aryNewNodeInfo))
	ptrNodeManager.objRpcFunc2NodeMap.Unlock()

	if _, aryChangeEntry := ptrNodeManager.ptrSelf.modifyEntry(
		[]*node.EntryInfo{&objEntry}, atomic.AddUint64(&ptrNodeManager.nSequence, 1)); len(aryChangeEntry) > 0 {

		taskworker.TaskWorkerManagerInstance().GoTask(context.TODO(), TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			var objNotify = node.NewEntryNotify{}
			objNotify.EntryInfo = append(objNotify.EntryInfo, &objEntry)
			objNotify.DataVersion = atomic.LoadUint64(&ptrNodeManager.ptrSelf.nDataVersion)

			ptrNodeManager.objLock4NodeInfoMap.RLock()
			defer ptrNodeManager.objLock4NodeInfoMap.RUnlock()

			for _, ptrNodeTemp := range ptrNodeManager.mNodeInfo {
				if ptrNodeTemp.szName == ptrNodeManager.ptrSelf.szName {
					continue
				}

				ptrNodeTemp.cast(rpc.NODE_ENTRY_NOTIFY, &objNotify, ctx).Start()
			}

			return CODE_SUCCESS
		})
	}
}

func (ptrNodeManager *NodeManager) AddEntry(szUri string, enumEntryType node.EntryType) {
	objEntry := node.EntryInfo{URI: szUri, EntryType: enumEntryType, IsNew: true}

	if ptrNodeManager.ptrSelf.entryExist(&objEntry) || !ptrNodeManager.ptrSelf.alive() {
		return
	}

	szUriSave := getUri(&objEntry)

	ptrNodeManager.objRpcFunc2NodeMap.Lock()
	ptrNodeManager.mRpcFunc2Node[szUriSave] = append(ptrNodeManager.mRpcFunc2Node[szUriSave], ptrNodeManager.ptrSelf)
	ptrNodeManager.objRpcFunc2NodeMap.Unlock()

	if _, aryChangeEntry := ptrNodeManager.ptrSelf.modifyEntry(
		[]*node.EntryInfo{&objEntry}, atomic.AddUint64(&ptrNodeManager.nSequence, 1)); len(aryChangeEntry) > 0 {
		taskworker.TaskWorkerManagerInstance().GoTask(context.TODO(), TASK_WORKER_TAG_NODE_MANAGER, func(ctx context.Context) int {
			var objNotify = node.NewEntryNotify{}
			objNotify.EntryInfo = append(objNotify.EntryInfo, &objEntry)
			objNotify.DataVersion = atomic.LoadUint64(&ptrNodeManager.ptrSelf.nDataVersion)

			ptrNodeManager.objLock4NodeInfoMap.RLock()
			defer ptrNodeManager.objLock4NodeInfoMap.RUnlock()

			for _, ptrNodeTemp := range ptrNodeManager.mNodeInfo {
				if ptrNodeTemp.szName == ptrNodeManager.ptrSelf.szName {
					continue
				}

				ptrNodeTemp.cast(rpc.NODE_ENTRY_NOTIFY, &objNotify, ctx).Start()
			}

			return CODE_SUCCESS
		})
	}
}

func (ptrNodeManager *NodeManager) Stop() {
	objNotify := node.SayGoodbyeNotify{}

	atomic.StoreInt32(&ptrNodeManager.nStatus, MANAGER_STATUS_STOPED)

	aryEntry := ptrNodeManager.ptrSelf.markDead()
	ptrNodeManager.deleteEntry(ptrNodeManager.ptrSelf, aryEntry)

	ptrNodeManager.objLock4NodeInfoMap.RLock()
	defer ptrNodeManager.objLock4NodeInfoMap.RUnlock()

	for _, ptrNodeTemp := range ptrNodeManager.mNodeInfo {
		if ptrNodeTemp.szName == ptrNodeManager.ptrSelf.szName {
			continue
		}

		ptrNodeTemp.cast(rpc.NODE_SAY_GOOBEY, &objNotify, context.TODO()).Start()
	}
}

func (ptrNodeManager *NodeManager) DispatchRsp(ptrRsp *node.RpcCallResp, ctx context.Context) node.ResultCode {

	var anyErr error
	objCaller := caller.Caller{}
	szCaller, bSuccess := caller.GetTopStack(ptrRsp)
	log.Infof("DispatchRsp Caller: %v", szCaller)

	if bSuccess {
		anyErr = json.Unmarshal([]byte(szCaller), &objCaller)
		if anyErr != nil {
			log.Warningf("json.Unmarshal error: %v,data: %v", anyErr, szCaller)
			return node.ResultCode_ERROR
		}
		if objCaller.Type != caller.Caller_Node {
			return node.ResultCode_CALL_YOUR_SELF
		}

		if ptrNode, bExist := ptrNodeManager.mNodeInfo[objCaller.Node]; bExist {
			caller.PopStack(ptrRsp) //pop caller stack
			ptrNode.cast(rpc.NODE_RPC_RESP, ptrRsp, ctx).Start()
			return node.ResultCode_SUCCESS
		} else {
			log.Errorf("node:%v not exist", objCaller.Node) //todo error log : done
			return node.ResultCode_NODE_NOT_EXIST
		}
	} else {
		log.Warningf("caller stack is nil!:%+v", ptrRsp)
		return node.ResultCode_STACK_EMPTY
	}
}

func (ptrNodeManager *NodeManager) DispatchReq(ptrReq *node.RpcCallReq, ctx context.Context) node.ResultCode {

	log.Infof("DispatchReq RequestMode: %v", ptrReq.RequestMode.String())

	switch ptrReq.RequestMode {
	case node.RequestMode_LOCAL_FORCE:
		if ptrNodeManager.ptrSelf.entryExist2(ptrReq.URI, ptrReq.EntryType, true) {
			return node.ResultCode_CALL_YOUR_SELF
		} else {
			return node.ResultCode_ENTRY_NOT_FOUND
		}
	case node.RequestMode_LOCAL_BETTER:
		if ptrNodeManager.ptrSelf.entryExist2(ptrReq.URI, ptrReq.EntryType, false) {
			return node.ResultCode_CALL_YOUR_SELF
		}
		//本地找不到，就找远端的
		szUri := getUri2(ptrReq.URI, ptrReq.EntryType)
		aryNode := ptrNodeManager.getRemoteNodeByUri(szUri)
		if len(aryNode) == 0 {
			return node.ResultCode_ENTRY_NOT_FOUND
		} else if ptrReq.Key == "" {
			ptrNode := aryNode[rand.Intn(len(aryNode))]
			caller.PushStack(ptrReq, caller.GenNodeCaller(ptrNodeManager.ptrSelf.szName))
			ptrNode.cast(rpc.NODE_RPC_REQ, ptrReq, ctx).Start()
			log.Debugf("cast node: %v", ptrNode.szName) //todo debug log:done
			return node.ResultCode_SUCCESS
		} else {
			ptrNode := aryNode[string2HashCode(ptrReq.Key)%len(aryNode)]
			caller.PushStack(ptrReq, caller.GenNodeCaller(ptrNodeManager.ptrSelf.szName)) //todo node name:done
			ptrNode.cast(rpc.NODE_RPC_REQ, ptrReq, ctx).Start()
			log.Infof("cast node: %v", ptrNode.szName)
			return node.ResultCode_SUCCESS
		}

	case node.RequestMode_DEFAULT: //todo 逻辑封装
		ptrNodeManager.objRpcFunc2NodeMap.RLock()
		defer ptrNodeManager.objRpcFunc2NodeMap.RUnlock()

		szUri := getUri2(ptrReq.URI, ptrReq.EntryType)
		if aryNode, bExist := ptrNodeManager.mRpcFunc2Node[szUri]; !bExist || len(aryNode) == 0 {
			return node.ResultCode_ENTRY_NOT_FOUND
		} else if ptrReq.Key == "" {
			ptrNode := aryNode[rand.Intn(len(aryNode))]

			caller.PushStack(ptrReq, caller.GenNodeCaller(ptrNodeManager.ptrSelf.szName))
			ptrNode.cast(rpc.NODE_RPC_REQ, ptrReq, ctx).Start()

			return node.ResultCode_SUCCESS
		} else {
			ptrNode := aryNode[string2HashCode(ptrReq.Key)%len(aryNode)]
			caller.PushStack(ptrReq, caller.GenNodeCaller(ptrNodeManager.ptrSelf.szName)) //todo node name:done
			ptrNode.cast(rpc.NODE_RPC_REQ, ptrReq, ctx).Start()

			return node.ResultCode_SUCCESS
		}
	default:
		log.Warningf("invalidate request mode:%v", ptrReq.RequestMode)
		return node.ResultCode_ENTRY_NOT_FOUND
	}
}

func (ptrNodeManager *NodeManager) EntryCheck(szEntryUri string, nEntryType node.EntryType) node.ResultCode {

	szUri := getUri2(szEntryUri, nEntryType)

	ptrNodeManager.objRpcFunc2NodeMap.RLock()
	defer ptrNodeManager.objRpcFunc2NodeMap.RUnlock()

	aryNode, bExist := ptrNodeManager.mRpcFunc2Node[szUri]
	if bExist && len(aryNode) > 0 {
		return node.ResultCode_SUCCESS
	}
	return node.ResultCode_ENTRY_NOT_FOUND
}

func (ptrNodeManager *NodeManager) getRemoteNodeByUri(szUri string) []*Info {
	ptrNodeManager.objRpcFunc2NodeMap.RLock()
	defer ptrNodeManager.objRpcFunc2NodeMap.RUnlock()

	var aryRpcNode []*Info
	aryNode, bExist := ptrNodeManager.mRpcFunc2Node[szUri]
	if !bExist {
		return aryRpcNode
	}
	for _, ptrNodeTmp := range aryNode {
		if ptrNodeTmp.szName == ptrNodeManager.ptrSelf.szName {
			continue
		}

		aryRpcNode = append(aryRpcNode, ptrNodeTmp)
	}

	return aryRpcNode
}

func (ptrNodeManager *NodeManager) NodeList() []*node.NodeInfo {

	ptrNodeManager.objLock4NodeInfoMap.RLock()
	defer ptrNodeManager.objLock4NodeInfoMap.RUnlock()

	var aryNodeInfoTemp, aryNodeInfo []*node.NodeInfo
	for _, ptrNode := range ptrNodeManager.mNodeInfo {
		if !ptrNode.alive() {
			continue
		}

		aryNodeInfoTemp = append(aryNodeInfoTemp,
			&node.NodeInfo{Name: ptrNode.szName,
				IP:         ptrNode.szIP,
				Port:       uint32(ptrNode.nPort),
				InstanceID: ptrNode.nInstanceID,
				AgentPort:  uint32(ptrNode.nAgentPort),
			})
	}

	nSize := len(aryNodeInfoTemp)
	if nSize == 0 {
		return aryNodeInfo
	}

	nCount := int(math.Min(GOOSIP_SIZE, float64(nSize)))
	nStart := rand.Intn(len(aryNodeInfoTemp))

	for ; nCount > 0; nCount-- {
		aryNodeInfo = append(aryNodeInfo, aryNodeInfoTemp[nStart%nSize])
		nStart++
	}
	return aryNodeInfo
}
