package node

import (
	"context"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/rpc"
	"git.yayafish.com/nbagent/taskworker"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Info struct {
	szName             string
	szIP               string
	nPort              uint16
	nAgentPort         uint16
	aryConnectionList  []*network.ServerConnection
	objLock4Connection sync.RWMutex
	nSequence          uint32
	nStatus            uint32
	objHandler         network.ConnectionHandler
	mEntryInfo         map[string]*node.EntryInfo
	objLock4Entry      sync.RWMutex
	nDataVersion       uint64
	nLastKeepAliveTime int64
	ctx                context.Context
	nInstanceID        uint64
}

const (
	NODE_STATUS_DEAD    = 1
	NODE_STATUS_ALIVE   = 2
	NODE_STATUS_SYNCING = 3
)

func newNode(szName, szIP string, nInstanceID uint64, nPort, nAgentPort uint16, objHandler network.ConnectionHandler, ctx context.Context) *Info {
	return &Info{
		szName:             szName,
		szIP:               szIP,
		nPort:              nPort,
		nAgentPort:         nAgentPort,
		nStatus:            NODE_STATUS_ALIVE,
		objHandler:         objHandler,
		mEntryInfo:         make(map[string]*node.EntryInfo),
		nLastKeepAliveTime: time.Now().Unix(),
		ctx:                ctx,
		nDataVersion:       0,
		nInstanceID:        nInstanceID,
	}
}

func (ptrNode *Info) removeConnection(ptrServerConnection *network.ServerConnection) int {
	var aryConnectionList []*network.ServerConnection

	ptrNode.objLock4Connection.Lock()
	defer ptrNode.objLock4Connection.Unlock()
	for _, ptrServerConnectionTmp := range ptrNode.aryConnectionList {
		if ptrServerConnectionTmp == ptrServerConnection {
			continue
		}

		aryConnectionList = append(aryConnectionList, ptrServerConnectionTmp)
	}

	ptrNode.aryConnectionList = aryConnectionList

	nLeft := len(ptrNode.aryConnectionList)

	log.Infof("connection of node:%v close, left:%v", ptrNode.szName, nLeft)
	return nLeft
}

func (ptrNode *Info) addConnection(ptrServerConnection *network.ServerConnection) {
	if !ptrNode.alive() {
		log.Infof("add connection fail: node:%v not alive", ptrNode.szName)
		ptrServerConnection.Close()
		return
	}

	ptrNode.objLock4Connection.RLock()
	nSize := len(ptrNode.aryConnectionList)
	ptrNode.objLock4Connection.RUnlock()

	if nSize > NeighbourMaxConnection {
		//大于最大连接数的话则关闭
		ptrServerConnection.Close()
		log.Infof("add connection fail: node:%v connection too mush", ptrNode.szName)
	} else {
		ptrNode.objLock4Connection.Lock()
		defer ptrNode.objLock4Connection.Unlock()

		ptrNode.aryConnectionList = append(ptrNode.aryConnectionList, ptrServerConnection)
		log.Infof("add connection of node:%v szIP:%v nPort:%v", ptrNode.szName, ptrNode.szIP, ptrNode.nPort)
	}
}

func (ptrNode *Info) alive() bool {
	return atomic.LoadUint32(&ptrNode.nStatus) != NODE_STATUS_DEAD
}

func (ptrNode *Info) markDead() (aryEntry []*node.EntryInfo) {
	atomic.StoreUint32(&ptrNode.nStatus, NODE_STATUS_DEAD)

	log.Infof("node:%v dead", ptrNode.szName)

	return ptrNode.getAllEntry()
}

func (ptrNode *Info) connectionAvailable() bool {
	ptrNode.objLock4Connection.RLock()
	defer ptrNode.objLock4Connection.RUnlock()
	return len(ptrNode.aryConnectionList) >= NeighbourMaxConnection/2
}

func (ptrNode *Info) reconnect() {
	if !ptrNode.alive() {
		return
	}

	if ptrNode.connectionAvailable() {
		return
	}

	if bSuccess, ptrServerConnection := network.NewClient(ptrNode.szIP, ptrNode.nPort, ptrNode.objHandler); bSuccess {
		log.Infof("connected to node:%v szIP:%v nPort:%v", ptrNode.szName, ptrNode.szIP, ptrNode.nPort)

		objReq := node.SayHelloReq{}
		objReq.NodeInfo = MyInfo(ptrNode.ctx)

		nNow := time.Now().Unix()
		objReq.Sign = makeSign(objReq.NodeInfo.Name, nNow)
		objReq.TimeStamp = nNow

		objRsp := node.SayHelloRsp{}

		ptrCallContext := rpc.NewCall(rpc.NODE_SAY_HELLO, &objReq, &objRsp, ptrServerConnection).Context(context.Background()).
			RequestID(MyName(ptrNode.ctx) + "_" + ptrNode.szName + "_" + strconv.FormatUint(uint64(atomic.AddUint32(&ptrNode.nSequence, 1)), 10))

		if ptrCallContext.Timeout(2000).Start() {
			if objRsp.Result != node.ResultCode_SUCCESS {
				log.Warningf("say hello fail:%+v", objRsp)
				ptrServerConnection.Close()
				return
			}

			if objRsp.NodeInfo == nil {
				log.Warningf("say hello fail: node info nil ->%+v ", objRsp)
				ptrServerConnection.Close()
				return
			}

			szSign := makeSign(objRsp.NodeInfo.Name, objRsp.TimeStamp)
			if szSign != objRsp.Sign {
				log.Warningf("sign validate fail: cal:%v input:%v %+v", szSign, objRsp.Sign, objRsp.NodeInfo)
				ptrServerConnection.Close()
				return
			}

			if objRsp.NodeInfo.Name != ptrNode.szName {
				log.Warningf("node name not match expected:%+v receive:%v %v", ptrNode.szName, objRsp.NodeInfo.Name, objReq.NodeInfo.Name)
				ptrServerConnection.Close()
				return
			}

			ptrServerConnection.Ctx = context.WithValue(ptrServerConnection.Ctx, CONTEXT_KEY_NODE, ptrNode)
			ptrNode.addConnection(ptrServerConnection)

			if objRsp.NodeInfo.InstanceID != ptrNode.getInstanceID() {
				log.Infof("data instance change:%v-%+v", ptrNode.getInstanceID(), objRsp.NodeInfo)
				ptrNode.setInstanceID(objRsp.NodeInfo.InstanceID)
			}
		}
	} else {
		log.Warningf("connect to node:%+v fail szIP:%v nPort:%v", ptrNode.szName, ptrNode.szIP, ptrNode.nPort)
		ptrServerConnection.Close()
	}
}

func (ptrNode *Info) getConnection() *network.ServerConnection {
	taskworker.TaskWorkerManagerInstance().GoTask(context.TODO(), TASK_WORKER_TAG_NODE_CONNECT, func(ctx context.Context) int {
		ptrNode.reconnect()
		return CODE_SUCCESS
	})

	ptrNode.objLock4Connection.RLock()
	defer ptrNode.objLock4Connection.RUnlock()

	if len(ptrNode.aryConnectionList) == 0 {
		return nil
	} else {
		return ptrNode.aryConnectionList[rand.Intn(len(ptrNode.aryConnectionList))]
	}
}

func (ptrNode *Info) cast(szUri string, objReq proto.Message, ctx context.Context) *rpc.CastContext {
	ptrServerConnection := ptrNode.getConnection()

	ptrCastContext := rpc.NewCast(szUri, objReq, ptrServerConnection).Context(ctx)
	szRequestID := network.GetRequestID(ctx)
	if szRequestID == "" {
		ptrCastContext.RequestID(MyName(ptrNode.ctx) + "_" + ptrNode.szName + "_" +
			strconv.FormatUint(uint64(atomic.AddUint32(&ptrNode.nSequence, 1)), 10))
	}

	return ptrCastContext
}

func (ptrNode *Info) call(szUri string, objReq proto.Message, objRsp proto.Message, ctx context.Context) *rpc.CallContext {
	ptrServerConnection := ptrNode.getConnection()

	ptrCallContext := rpc.NewCall(szUri, objReq, objRsp, ptrServerConnection).Context(ctx)

	szRequestID := network.GetRequestID(ctx)
	if szRequestID == "" {
		ptrCallContext.RequestID(MyName(ptrNode.ctx) + "_" + ptrNode.szName + "_" +
			strconv.FormatUint(uint64(atomic.AddUint32(&ptrNode.nSequence, 1)), 10))
	}

	return rpc.NewCall(szUri, objReq, objRsp, ptrServerConnection).Context(ctx).RequestID(szRequestID)
}

func (ptrNode *Info) syncEntry(aryEntry []*node.EntryInfo, nVersion uint64) {
	atomic.StoreUint64(&ptrNode.nDataVersion, nVersion)

	mEntryInfo := make(map[string]*node.EntryInfo)
	for _, ptrEntry := range aryEntry {
		szUri := getUri(ptrEntry)
		if ptrEntry.IsNew {
			mEntryInfo[szUri] = ptrEntry
		} else {
			delete(mEntryInfo, szUri)
		}
	}

	atomic.StoreUint32(&ptrNode.nStatus, NODE_STATUS_ALIVE)

	ptrNode.objLock4Entry.Lock()
	defer ptrNode.objLock4Entry.Unlock()
	ptrNode.mEntryInfo = mEntryInfo
}

func (ptrNode *Info) entryExist2(szUri string, enmEntryType node.EntryType, bTolerateDeadNode bool) bool {
	if !bTolerateDeadNode && !ptrNode.alive() {
		return false
	}

	ptrNode.objLock4Entry.RLock()
	defer ptrNode.objLock4Entry.RUnlock()

	_, bExist := ptrNode.mEntryInfo[getUri2(szUri, enmEntryType)]
	return bExist
}

func (ptrNode *Info) entryExist(ptrEntry *node.EntryInfo) bool {
	ptrNode.objLock4Entry.RLock()
	defer ptrNode.objLock4Entry.RUnlock()

	_, bExist := ptrNode.mEntryInfo[getUri(ptrEntry)]
	return bExist
}

func (ptrNode *Info) modifyEntry(aryEntry []*node.EntryInfo, nVersion uint64) (bool, []*node.EntryInfo) {
	if nVersion != atomic.LoadUint64(&ptrNode.nDataVersion)+1 {
		atomic.StoreUint32(&ptrNode.nStatus, NODE_STATUS_SYNCING)

		return true, nil
	}

	atomic.StoreUint64(&ptrNode.nDataVersion, nVersion)
	var aryChangeEntry []*node.EntryInfo

	ptrNode.objLock4Entry.Lock()
	defer ptrNode.objLock4Entry.Unlock()

	for _, ptrEntry := range aryEntry {
		szUri := getUri(ptrEntry)
		if ptrEntry.IsNew && ptrNode.alive() {
			ptrNode.mEntryInfo[szUri] = ptrEntry
			aryChangeEntry = append(aryChangeEntry, ptrEntry)
		} else if !ptrEntry.IsNew {
			delete(ptrNode.mEntryInfo, szUri)
			aryChangeEntry = append(aryChangeEntry, ptrEntry)
		}
	}

	return false, aryChangeEntry
}

func (ptrNode *Info) getAllEntry() []*node.EntryInfo {
	var aryEntry []*node.EntryInfo

	ptrNode.objLock4Entry.RLock()
	defer ptrNode.objLock4Entry.RUnlock()

	for _, ptrEntry := range ptrNode.mEntryInfo {
		aryEntry = append(aryEntry, ptrEntry)
	}

	return aryEntry
}

func (ptrNode *Info) needSyncAll(nDataVersion uint64) bool {

	if atomic.LoadUint64(&ptrNode.nDataVersion) != nDataVersion {
		atomic.StoreUint32(&ptrNode.nStatus, NODE_STATUS_SYNCING)
		return true
	}

	return false
}

func (ptrNode *Info) getState() uint32 {
	return atomic.LoadUint32(&ptrNode.nStatus)
}

func (ptrNode *Info) refreshKeepLiveTime() {
	atomic.StoreInt64(&ptrNode.nLastKeepAliveTime, time.Now().Unix())
}

func (ptrNode *Info) timeout(nTimeout int64) bool {
	return time.Now().Unix() > atomic.LoadInt64(&ptrNode.nLastKeepAliveTime)+nTimeout
}

func getUri(ptrEntry *node.EntryInfo) string {
	return ptrEntry.URI + "_" + ptrEntry.EntryType.String()
}

func getUri2(szUri string, enmEntryType node.EntryType) string {
	return szUri + "_" + enmEntryType.String()
}

func (ptrNode *Info) setInstanceID(nInstanceID uint64) {
	atomic.StoreUint64(&ptrNode.nDataVersion, 0)
	atomic.StoreUint64(&ptrNode.nInstanceID, nInstanceID)
}

func (ptrNode *Info) getInstanceID() uint64 {
	return atomic.LoadUint64(&ptrNode.nInstanceID)
}
