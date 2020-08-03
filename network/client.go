package network

import (
	"context"
	"git.yayafish.com/nbagent/log"
	"net"
	"strconv"
	"sync"
)

type Client struct {
	szIP  string
	nPort uint16
}

func NewClient(szIP string, nPort uint16, objHandler ConnectionHandler,
	aryFn ...func(szIP string, nPort uint16, ptrServerConnection *ServerConnection) error) (bool, *ServerConnection) {
	szAddress := szIP + ":" + strconv.FormatUint(uint64(nPort), 10)
	if ptrAddr, anyErr := net.ResolveTCPAddr("tcp", szAddress); anyErr != nil {
		log.Warningf("get address:%s fail:%v", szAddress, anyErr)
		return false, nil
	} else if objConnection, anyErr := net.DialTCP("tcp", nil, ptrAddr); anyErr != nil {
		log.Warningf("connect address:%s fail:%v", szAddress, anyErr)
		return false, nil
	} else {
		objServerConnection := ServerConnection{
			objConnection,
			nil,
			&Client{
				szIP,
				nPort,
			},
			new(sync.Mutex),
			objHandler, 0,
			context.TODO(),
			nil,
			nil,
		}
		if len(aryFn) == 0 {
			mRequestContext := make(map[string]*RpcContext)

			objServerConnection.mRequestContext = &mRequestContext
			objServerConnection.objRWLock = &sync.Mutex{}
		} else {
			for _, ptrFn := range aryFn {
				if ptrFn != nil {
					if anyErr := ptrFn(szIP, nPort, &objServerConnection); anyErr != nil {
						return false, nil
					}
				}
			}
		}
		go objServerConnection.readLoop()

		return true, &objServerConnection
	}
}
