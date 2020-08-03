package client

import (
	"fmt"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"sync"
	"time"
)

type ConnectionPool struct {
	nInitConnectNum int
	nMaxConnectNum  int

	mConnection map[*network.ServerConnection]*ConnectionInfo

	objRWLock sync.RWMutex
}

func (ptrConnectPool *ConnectionPool) Init(nMaxIdleConnectNum, nMaxConnectNum int) error {

	ptrConnectPool.nInitConnectNum = nMaxIdleConnectNum
	ptrConnectPool.nMaxConnectNum  = nMaxConnectNum
	ptrConnectPool.mConnection 	   = make(map[*network.ServerConnection]*ConnectionInfo)

	return nil
}

func (ptrConnectPool *ConnectionPool) GetConnectNum() int {
	ptrConnectPool.objRWLock.RLock()
	defer ptrConnectPool.objRWLock.RUnlock()
	return len(ptrConnectPool.mConnection)
}

func (ptrConnectPool *ConnectionPool) GetMaxConnectNum() int {
	ptrConnectPool.objRWLock.RLock()
	defer ptrConnectPool.objRWLock.RUnlock()
	return ptrConnectPool.nMaxConnectNum
}

func (ptrConnectPool *ConnectionPool) GetInitConnectNum() int {
	ptrConnectPool.objRWLock.RLock()
	defer ptrConnectPool.objRWLock.RUnlock()
	return ptrConnectPool.nInitConnectNum
}

func (ptrConnectPool *ConnectionPool) GetConnect() (ptrConnection *ConnectionInfo) {

	ptrConnectPool.objRWLock.RLock()
	defer ptrConnectPool.objRWLock.RUnlock()

	for _, ptrConnection = range ptrConnectPool.mConnection {
		break
	}

	return
}

func (ptrConnectPool *ConnectionPool) AddConnect(ptrConnection *ConnectionInfo) {
	ptrConnectPool.objRWLock.Lock()
	defer ptrConnectPool.objRWLock.Unlock()

	ptrConnectPool.mConnection[ptrConnection.Connection] = ptrConnection
}

func (ptrConnectPool *ConnectionPool) NewConnect(szIP string, nPort uint16, objHandler network.ConnectionHandler, aryFn ...func(szIP string, nPort uint16, ptrServerConnection *network.ServerConnection) error) (*ConnectionInfo, error) {
	if bSuccess, ptrConnection := network.NewClient(szIP, nPort, objHandler, aryFn...); ! bSuccess {
		anyErr := fmt.Errorf("network.NewClient fail, ip:%v, port:%v", szIP, nPort)
		log.Warningf(anyErr.Error())
		return nil, anyErr
	} else {
		ptrConnection := &ConnectionInfo{
			LastKeepaliveTime: time.Now(),
			Connection:        ptrConnection,
			Host:              makeHostKeyByIpPort(szIP, nPort),
		}

		// 这时候还没注册，先不放入进行管理

		return ptrConnection, nil
	}
}

func (ptrConnectPool *ConnectionPool) ClearConnect(ptrConnection *network.ServerConnection) {
	log.Debugf("ClearConnect[%v]", ConnectToString(ptrConnection))

	ptrConnectPool.objRWLock.Lock()

	delete(ptrConnectPool.mConnection, ptrConnection)

	ptrConnectPool.objRWLock.Unlock()

	ptrConnection.Close()
}

func (ptrConnectPool *ConnectionPool) IsFullConnect() bool {
	ptrConnectPool.objRWLock.RLock()
	defer ptrConnectPool.objRWLock.RUnlock()

	return len(ptrConnectPool.mConnection) >= ptrConnectPool.nMaxConnectNum
}
