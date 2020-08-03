package cli_agent

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"git.yayafish.com/nbagent/reporter"

	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"git.yayafish.com/nbagent/protocol/agent"
	"git.yayafish.com/nbagent/protocol/caller"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/rpc"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

type ClientInfo struct {
	szInstanceID       string
	aryConnectionList  []*network.ServerConnection
	objLock4Connection sync.RWMutex
	mEntryInfo         map[string]*agent.EntryInfo
	objLock4Entry      sync.Mutex
	nStatus            uint32
	objHandler         network.ConnectionHandler
	nLastKeepAliveTime int64
	ctx                context.Context
}

const (
	AGENT_STATUS_DEAD  = 1
	AGENT_STATUS_ALIVE = 2
)

func NewClient(szInstanceID string, objHandler network.ConnectionHandler, ctx context.Context) *ClientInfo {
	return &ClientInfo{
		szInstanceID:       szInstanceID,
		nStatus:            AGENT_STATUS_ALIVE,
		objHandler:         objHandler,
		nLastKeepAliveTime: time.Now().Unix(),
		ctx:                ctx,
	}
}

func (ptrClientInfo *ClientInfo) getClientInfo() string {
	//return fmt.Sprintf("%s_%s_%d_%d", ptrClientInfo.szName, ptrClientInfo.szIP, ptrClientInfo.nPort, ptrClientInfo.nInstanceID)
	return ptrClientInfo.szInstanceID
}

func (ptrClientInfo *ClientInfo) onCloseConnection(ptrServerConnection *network.ServerConnection) int {
	ptrClientInfo.objLock4Connection.Lock()
	defer ptrClientInfo.objLock4Connection.Unlock()

	var aryConnectionList []*network.ServerConnection
	for _, ptrServerConnectionTmp := range ptrClientInfo.aryConnectionList {
		if ptrServerConnectionTmp == ptrServerConnection {
			continue
		}
		aryConnectionList = append(aryConnectionList, ptrServerConnectionTmp)
	}

	ptrClientInfo.aryConnectionList = aryConnectionList
	nLeft := len(ptrClientInfo.aryConnectionList)
	log.Infof("connection of agent:%v close, left:%v", ptrClientInfo.getClientInfo(), nLeft)
	return nLeft
}

func (ptrClientInfo *ClientInfo) addConnection(ptrServerConnection *network.ServerConnection) {

	reporter.Count("ClientInfo_addConnection")

	if !ptrClientInfo.alive() {
		log.Infof("add connection fail: agent:%v not alive", ptrClientInfo.getClientInfo())
		ptrServerConnection.Close()
		return
	}

	ptrClientInfo.objLock4Connection.RLock()
	bExist := false
	nSize := len(ptrClientInfo.aryConnectionList)
	for _, ptrServerConnectionTmp := range ptrClientInfo.aryConnectionList {
		if ptrServerConnectionTmp == ptrServerConnection {
			bExist = true
			break
		}
	}
	ptrClientInfo.objLock4Connection.RUnlock()
	if bExist {
		return
	}

	if nSize > AgentClientMaxConnection {
		//大于最大连接数的话则关闭
		ptrServerConnection.Close()
		log.Infof("agent :%v,exceed max connection limit: %v", ptrClientInfo.getClientInfo(), nSize)
	} else {
		ptrClientInfo.objLock4Connection.Lock()
		ptrClientInfo.aryConnectionList = append(ptrClientInfo.aryConnectionList, ptrServerConnection)
		ptrClientInfo.objLock4Connection.Unlock()
		log.Infof("add connection: %v, agent:%v", getConnectionID(ptrServerConnection), ptrClientInfo.getClientInfo())
	}
}

func (ptrClientInfo *ClientInfo) alive() bool {
	return atomic.LoadUint32(&ptrClientInfo.nStatus) != AGENT_STATUS_DEAD
}

func (ptrClientInfo *ClientInfo) markDead() (aryEntry []*agent.EntryInfo) {
	atomic.StoreUint32(&ptrClientInfo.nStatus, AGENT_STATUS_DEAD)

	var aryConnectionList []*network.ServerConnection

	ptrClientInfo.objLock4Connection.RLock()
	for _, ptrServerConnectionTmp := range ptrClientInfo.aryConnectionList {
		aryConnectionList = append(aryConnectionList, ptrServerConnectionTmp)
	}

	ptrClientInfo.objLock4Connection.RUnlock()

	for _, ptrServerConnectionTmp := range aryConnectionList {
		ptrServerConnectionTmp.Close()
	}

	log.Infof("agent: %v dead", ptrClientInfo.getClientInfo())

	return ptrClientInfo.getAllEntry()
}

func (ptrClientInfo *ClientInfo) connectionAvailable() bool {
	ptrClientInfo.objLock4Connection.RLock()
	defer ptrClientInfo.objLock4Connection.RUnlock()

	return len(ptrClientInfo.aryConnectionList) >= AgentClientMaxConnection/2
}

func (ptrClientInfo *ClientInfo) getConnection() *network.ServerConnection {

	reporter.Count("ClientInfo_getConnection")

	ptrClientInfo.objLock4Connection.RLock()
	defer ptrClientInfo.objLock4Connection.RUnlock()

	if len(ptrClientInfo.aryConnectionList) == 0 {
		return nil
	} else {
		return ptrClientInfo.aryConnectionList[rand.Intn(len(ptrClientInfo.aryConnectionList))]
	}
}

func (ptrClientInfo *ClientInfo) getRespConnection(objCaller caller.Caller) *network.ServerConnection {

	reporter.Count("ClientInfo_getRespConnection")

	ptrClientInfo.objLock4Connection.RLock()
	defer ptrClientInfo.objLock4Connection.RUnlock()

	var ptrConnection *network.ServerConnection
	if len(ptrClientInfo.aryConnectionList) <= 0 {
		log.Warningf("Connection list null ,%v", ptrClientInfo.getClientInfo())
		return nil
	}

	var szConnectionID string
	for _, ptrTmpConn := range ptrClientInfo.aryConnectionList {
		szConnectionID = getConnectionID(ptrTmpConn)
		if szConnectionID == objCaller.ConnectionID {
			ptrConnection = ptrTmpConn
			break
		}
	}
	//找不到原请求连接，随机选一个连接进行回复
	if ptrConnection == nil {
		ptrConnection = ptrClientInfo.aryConnectionList[rand.Intn(len(ptrClientInfo.aryConnectionList))]
		log.Warningf("client: %v,can not find caller connection: %v,use rand connection: %v",
			ptrClientInfo.getClientInfo(), objCaller.ConnectionID, getConnectionID(ptrConnection))
	}

	return ptrConnection
}

func (ptrClientInfo *ClientInfo) syncEntry(aryEntry []*agent.EntryInfo) {
	ptrClientInfo.objLock4Entry.Lock()
	defer ptrClientInfo.objLock4Entry.Unlock()

	mEntryInfo := make(map[string]*agent.EntryInfo)
	for _, ptrEntry := range aryEntry {
		szUri := getUri(ptrEntry)
		mEntryInfo[szUri] = ptrEntry
	}

	ptrClientInfo.mEntryInfo = mEntryInfo
	log.Debugf("entryInfo:%+v", ptrClientInfo.mEntryInfo)
	atomic.StoreUint32(&ptrClientInfo.nStatus, AGENT_STATUS_ALIVE)
}

func (ptrClientInfo *ClientInfo) getAllEntry() []*agent.EntryInfo {
	ptrClientInfo.objLock4Entry.Lock()
	defer ptrClientInfo.objLock4Entry.Unlock()

	var aryEntry []*agent.EntryInfo
	for _, ptrEntry := range ptrClientInfo.mEntryInfo {
		aryEntry = append(aryEntry, ptrEntry)
	}

	return aryEntry
}

func (ptrClientInfo *ClientInfo) getState() uint32 {
	return atomic.LoadUint32(&ptrClientInfo.nStatus)
}

func (ptrClientInfo *ClientInfo) refreshKeepLiveTime() {
	atomic.StoreInt64(&ptrClientInfo.nLastKeepAliveTime, time.Now().Unix())
}

func (ptrClientInfo *ClientInfo) timeout(nTimeout int64) bool {
	return time.Now().Unix() > atomic.LoadInt64(&ptrClientInfo.nLastKeepAliveTime)+nTimeout
}

func getUri(ptrEntry *agent.EntryInfo) string {
	return ptrEntry.URI + "_" + ptrEntry.EntryType.String()
}

func (ptrClientInfo *ClientInfo) setInstanceID(szInstanceID string) {
	//atomic.StoreUint64(&ptrClientInfo.nInstanceID, nInstanceID)
	ptrClientInfo.szInstanceID = szInstanceID
}

func (ptrClientInfo *ClientInfo) getInstanceID() string {
	//return atomic.LoadUint64(&ptrClientInfo.nInstanceID)
	return ptrClientInfo.szInstanceID
}

func (ptrClientInfo *ClientInfo) rpcCall(ptrReq *node.RpcCallReq) bool {

	var bRet bool
	var szResult string = "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("ClientInfo_rpcCall")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	if atomic.LoadUint32(&ptrClientInfo.nStatus) != AGENT_STATUS_ALIVE {
		log.Warningf("client not alive: %v", ptrClientInfo.getClientInfo())
		bRet = false
		szResult = "client not alive"
		return bRet
	}

	var ptrConnection *network.ServerConnection
	ptrConnection = ptrClientInfo.getConnection()
	if ptrConnection == nil {
		log.Warningf("can not find able connection")
		bRet = false
		szResult = "not find able connection"
		return bRet
	}

	objUUID := uuid.New()
	bRet = ptrClientInfo.sendMessage(rpc.NODE_RPC_REQ, protocol.MSG_TYPE_REQUEST,
		objUUID.String(), ptrReq, ptrConnection)
	if !bRet {
		szResult = "sendMessage fail"
	}
	return bRet
}

func (ptrClientInfo *ClientInfo) rpcResp(ptrResp *node.RpcCallResp, objCaller caller.Caller) bool {

	var bRet bool
	var szResult string = "SUCCESS"
	objReportHelper := reporter.NewAutoReportHelper("ClientInfo_rpcResp")
	defer func() {
		objReportHelper.Report(szResult)
	}()

	if atomic.LoadUint32(&ptrClientInfo.nStatus) != AGENT_STATUS_ALIVE {
		log.Warningf("client not alive ,%v", ptrClientInfo.getClientInfo())
		bRet = false
		szResult = "client not alive"
		return bRet
	}

	var ptrConnection *network.ServerConnection
	ptrConnection = ptrClientInfo.getRespConnection(objCaller)
	if ptrConnection == nil {
		log.Warningf("can not find response connection")
		bRet = false
		szResult = "not find response connection"
		return bRet
	}

	bRet = ptrClientInfo.sendMessage(rpc.NODE_RPC_RESP, protocol.MSG_TYPE_RESPONE,
		objCaller.RequestID, ptrResp, ptrConnection)
	if !bRet {
		szResult = "sendMessage fail"
	}

	return bRet
}

func (ptrClientInfo *ClientInfo) sendMessage(szUri string, nMsgType uint8,
	szRequestID string, objMsg proto.Message,
	ptrConnection *network.ServerConnection) bool {

	var bRet bool
	objReportHelper := reporter.NewAutoReportHelper("ClientInfo_sendMessage")
	defer func() {
		szResult := "SUCCESS"
		if !bRet {
			szResult = "FAIL"
		}
		objReportHelper.Report(szResult)
	}()

	var byteData []byte
	byteData, bRet = protocol.NewProtocol(szUri, nMsgType,
		szRequestID, objMsg)
	if !bRet {
		log.Warningf("protocol.NewProtocol error, data: %+v", objMsg)
		return bRet
	}
	bRet = ptrConnection.Write(byteData)
	if !bRet {
		log.Warningf("Write error, data: %+v", objMsg)
		return bRet
	}

	return bRet
}

func (ptrClientInfo *ClientInfo) cast(szUri string, objReq proto.Message, ctx context.Context) bool {

	var bRet bool
	objReportHelper := reporter.NewAutoReportHelper("ClientInfo_cast")
	defer func() {
		szResult := "SUCCESS"
		if !bRet {
			szResult = "FAIL"
		}
		objReportHelper.Report(szResult)
	}()

	var ptrConnection *network.ServerConnection
	ptrConnection = ptrClientInfo.getConnection()
	if ptrConnection == nil {
		log.Warningf("can not find able connection")
		bRet = false
		return bRet
	}

	ptrCastContext := rpc.NewCast(szUri, objReq, ptrConnection).Context(ctx)
	ptrCastContext.RequestID(uuid.New().String())
	bRet = ptrCastContext.Start()
	return bRet
}
