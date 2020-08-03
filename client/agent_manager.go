package client

import (
	"fmt"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"sync"
	"time"
)

type AgentConnectPoolInfo struct {
	AgentInfo
	ConnectionPool
	//nLastKeepAliveTime int64
}

func (ptrAgentConnectPool *AgentConnectPoolInfo) Init(objAgent AgentInfo, nInitConnectNum, nMaxConnectNum int) error {
	ptrAgentConnectPool.AgentInfo = objAgent
	//ptrAgentConnectPool.nLastKeepAliveTime = time.Now().Unix()
	return ptrAgentConnectPool.ConnectionPool.Init(nInitConnectNum, nMaxConnectNum)
}

func (ptrAgentConnectPool *AgentConnectPoolInfo) CreateAgentConnect(objHandler network.ConnectionHandler, aryFn ...func(szIP string, nPort uint16, ptrServerConnection *network.ServerConnection) error) (*ConnectionInfo, error) {

	if ptrConnect, anyErr := ptrAgentConnectPool.NewConnect(ptrAgentConnectPool.IP, ptrAgentConnectPool.Port, objHandler, aryFn...); anyErr != nil {
		anyErr := fmt.Errorf("ConnectPool.NewConnect err:%v, ip:%v, port:%v", anyErr, ptrAgentConnectPool.IP, ptrAgentConnectPool.Port)
		log.Warningf(anyErr.Error())
		return nil, anyErr
	} else {
		// 这时候还没注册，先不放入进行管理

		return ptrConnect, nil
	}
}

//func (ptrAgentConnectPool *AgentConnectPoolInfo) refreshKeepAliveTime() {
//	ptrAgentConnectPool.objRWLock.Lock()
//	defer ptrAgentConnectPool.objRWLock.Unlock()
//
//	ptrAgentConnectPool.nLastKeepAliveTime = time.Now().Unix()
//}

type AgentGoodByeInfo struct {
	LastUpdateTimestamp int64
}

type AgentManager struct {
	szActiveAgent string
	mAgentConnectPool map[string]*AgentConnectPoolInfo

	mAgentGoodByeInfo     map[string]*AgentGoodByeInfo
	nAgentGoodByeTimeoutS int

	objRWLock sync.RWMutex
}

func (ptrAgentManager *AgentManager) GetActiveAgentConnectPool() (ptrPool *AgentConnectPoolInfo, anyErr error) {
	ptrAgentManager.objRWLock.RLock()
	defer ptrAgentManager.objRWLock.RUnlock()

	if ptrAgentPoolInfo, bExisted := ptrAgentManager.mAgentConnectPool[ptrAgentManager.szActiveAgent]; bExisted {
		ptrPool = ptrAgentPoolInfo
	} else {
		if len(ptrAgentManager.szActiveAgent) == 0 {
			anyErr = fmt.Errorf("agent pool not exist agent:%v, need add agent", ptrAgentManager.szActiveAgent)
		} else {
			anyErr = fmt.Errorf("agent pool not exist agent:%v", ptrAgentManager.szActiveAgent)
		}
	}

	return
}

func (ptrAgentManager *AgentManager) AddAgent(ptrPool *AgentConnectPoolInfo) (bExisted bool) {
	szHostKey := makeHostKeyByIpPort(ptrPool.IP, ptrPool.Port)

	ptrAgentManager.objRWLock.Lock()
	defer ptrAgentManager.objRWLock.Unlock()

	// 第一次设为第一个
	if len(ptrAgentManager.szActiveAgent) == 0 {
		ptrAgentManager.szActiveAgent = szHostKey
	}

	if _, bExisted = ptrAgentManager.mAgentConnectPool[szHostKey]; ! bExisted {
		ptrAgentManager.mAgentConnectPool[szHostKey] = ptrPool
	}

	return
}

// TODO: 删除的agent需要留意还剩的连接数，这个逻辑由外部保证
func (ptrAgentManager *AgentManager) DeleteAgent(szHostKey string) (bDeleted bool) {
	ptrAgentManager.objRWLock.Lock()
	defer ptrAgentManager.objRWLock.Unlock()

	if ptrAgentManager.szActiveAgent == szHostKey {
		log.Errorf("cannot delete active agent:%v, delete fail", ptrAgentManager.szActiveAgent)
		return false
	}

	if _, bExist := ptrAgentManager.mAgentConnectPool[szHostKey]; bExist {
		bDeleted = bExist
		delete(ptrAgentManager.mAgentConnectPool, szHostKey)
	}

	return
}

// 最后保留新旧agent的交集+新agent(需要new连接池)+当前的agent
func (ptrAgentManager *AgentManager) UpdateAgent(aryUpdateAgent []AgentInfo, nMaxIdleConnectNum, nMaxConnectNum int) {
	mUpdateAgent := make(map[string]*AgentInfo)
	for i := range aryUpdateAgent {
		mUpdateAgent[makeHostKeyByIpPort(aryUpdateAgent[i].IP, aryUpdateAgent[i].Port)] = &aryUpdateAgent[i]
	}

	mCurrentAllAgent := ptrAgentManager.getAllAgentPool()
	mAllAgentPool := make(map[string]*AgentConnectPoolInfo)

	// 取交集，同时拿到新的agent
	//mSameAgent := make(map[string]*AgentInfo)
	mNewAgent := make(map[string]*AgentInfo)
	for szUpdateAgent := range mUpdateAgent {
		if _, bExist := mCurrentAllAgent[szUpdateAgent]; bExist {
			// 交集
			//mSameAgent[szUpdateAgent] = mUpdateAgent[szUpdateAgent]
			// 取出交集的连接池
			mAllAgentPool[szUpdateAgent] = mCurrentAllAgent[szUpdateAgent]
		} else {
			// 新agent
			mNewAgent[szUpdateAgent] = mUpdateAgent[szUpdateAgent]
		}
	}

	// 初始化新agent的连接池
	for szAgent := range mNewAgent {
		ptrAgentConnectPool := &AgentConnectPoolInfo{}
		anyErr := ptrAgentConnectPool.Init(*mNewAgent[szAgent], nMaxIdleConnectNum, nMaxConnectNum)
		if anyErr != nil {
			log.Errorf("AgentConnectPool.Init err:%v, maxConnectNum:%v", anyErr, nMaxConnectNum)
			continue
		}

		mAllAgentPool[szAgent] = ptrAgentConnectPool
	}

	// 将当前的agent的复制出来以及切换map两个操作需要是原子的
	ptrAgentManager.objRWLock.Lock()
	defer ptrAgentManager.objRWLock.Unlock()
	// 如果之前的信息里面没有当前的agent，复制agent
	if _, bExist := mAllAgentPool[ptrAgentManager.szActiveAgent]; ! bExist {
		mAllAgentPool[ptrAgentManager.szActiveAgent] = ptrAgentManager.mAgentConnectPool[ptrAgentManager.szActiveAgent]
	}
	ptrAgentManager.mAgentConnectPool = mAllAgentPool
}

func (ptrAgentManager *AgentManager) GetAgent(szHostKey string) (ptrPool *AgentConnectPoolInfo, bExist bool) {
	ptrAgentManager.objRWLock.RLock()
	defer ptrAgentManager.objRWLock.RUnlock()
	ptrPool, bExist = ptrAgentManager.mAgentConnectPool[szHostKey]
	return
}

func (ptrAgentManager *AgentManager) SwitchActiveAgent() {

	var bHasOtherAgent = false
	var szNewAgentHost string

	ptrAgentManager.objRWLock.RLock()
	for szNewAgentHost = range ptrAgentManager.mAgentConnectPool {
		// 不使用发送了good bye的agent
		if _, bIsGoodByeAgent := ptrAgentManager.mAgentGoodByeInfo[szNewAgentHost]; bIsGoodByeAgent {
			continue
		}
		if szNewAgentHost != ptrAgentManager.szActiveAgent {
			bHasOtherAgent = true
			break
		}
	}
	ptrAgentManager.objRWLock.RUnlock()

	if bHasOtherAgent {
		ptrAgentManager.objRWLock.Lock()
		ptrAgentManager.szActiveAgent = szNewAgentHost
		// 目前不清map
		ptrAgentManager.objRWLock.Unlock()
	}

	return
}

func (ptrAgentManager *AgentManager) SwitchToOtherAgent(szAgentHost string) {
	var bIsActiveAgent = false

	ptrAgentManager.objRWLock.RLock()
	if szAgentHost == ptrAgentManager.szActiveAgent {
		bIsActiveAgent = true
	}
	ptrAgentManager.objRWLock.RUnlock()

	if bIsActiveAgent {
		ptrAgentManager.SwitchActiveAgent()
	}
}

func (ptrAgentManager *AgentManager) AddGoodByeAgent(szAgentHost string) (bAgentExist bool) {
	bAgentExist = false
	var ptrGoodByeAgent *AgentGoodByeInfo
	var nCurTimestamp = time.Now().Unix()

	ptrAgentManager.objRWLock.Lock()
	if ptrGoodByeAgent, bAgentExist = ptrAgentManager.mAgentGoodByeInfo[szAgentHost]; bAgentExist && ptrGoodByeAgent != nil {
		ptrGoodByeAgent.LastUpdateTimestamp = nCurTimestamp
	} else {
		ptrAgentManager.mAgentGoodByeInfo[szAgentHost] = &AgentGoodByeInfo{
			LastUpdateTimestamp: nCurTimestamp,
		}
	}
	ptrAgentManager.objRWLock.Unlock()

	return
}

func (ptrAgentManager *AgentManager) ClearGoodbyeTimeout() {
	var nCurTimestamp = time.Now().Unix()
	var aryGoodByeTimeoutAgent []string

	ptrAgentManager.objRWLock.RLock()
	for szAgent, ptrAgent := range ptrAgentManager.mAgentGoodByeInfo {
		if ptrAgent.LastUpdateTimestamp - nCurTimestamp > int64(ptrAgentManager.nAgentGoodByeTimeoutS) {
			aryGoodByeTimeoutAgent = append(aryGoodByeTimeoutAgent, szAgent)
		}
	}
	ptrAgentManager.objRWLock.RUnlock()

	if len(aryGoodByeTimeoutAgent) > 0 {
		ptrAgentManager.objRWLock.Lock()
		for _, szAgent := range aryGoodByeTimeoutAgent {
			delete(ptrAgentManager.mAgentGoodByeInfo, szAgent)
		}
		ptrAgentManager.objRWLock.Unlock()
	}
}

func (ptrAgentManager *AgentManager) getAllAgentPool() (mAgentPool map[string]*AgentConnectPoolInfo) {
	mAgentPool = make(map[string]*AgentConnectPoolInfo)

	ptrAgentManager.objRWLock.RLock()
	defer ptrAgentManager.objRWLock.RUnlock()

	for szAgent := range ptrAgentManager.mAgentConnectPool {
		mAgentPool[szAgent] = ptrAgentManager.mAgentConnectPool[szAgent]
	}

	return
}

func (ptrAgentManager *AgentManager) getAllAgent() (mAgent map[string]*AgentInfo) {
	mAgent = make(map[string]*AgentInfo)

	ptrAgentManager.objRWLock.RLock()
	defer ptrAgentManager.objRWLock.RUnlock()

	for szAgent, objAgent := range ptrAgentManager.mAgentConnectPool {
		mAgent[szAgent] = &objAgent.AgentInfo
	}

	return
}

func (ptrAgentManager *AgentManager) getAgentSize() int {
	ptrAgentManager.objRWLock.RLock()
	defer ptrAgentManager.objRWLock.RUnlock()

	return len(ptrAgentManager.mAgentConnectPool)
}
