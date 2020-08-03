package network

import (
	"context"
	"git.yayafish.com/nbagent/log"
	"net"
	"strconv"
	"sync"
)

const PACKET_LIMIT uint32 = 1024 * 1024 * 8

type Server struct {
	szIP                 string
	nPort                uint16
	objListener          net.Listener
	objHandler           ServerHandler
	objConnectionHandler ConnectionHandler
}

func NewServer(szIP string, nPort uint16, bStartDynamic bool, objHandler ServerHandler, objConnectionHandler ConnectionHandler) *Server {
	for {
		szAddress := szIP + ":" + strconv.FormatUint(uint64(nPort), 10)

		if objListener, anyErr := net.Listen("tcp", szAddress); anyErr != nil && !bStartDynamic {
			log.Panicf("tcp server start fail!:%v", anyErr)
			return nil
		} else if anyErr != nil && bStartDynamic {
			nPort += 1
			log.Warningf("tcp server start fail!:%v", anyErr)
			continue
		} else {
			log.Infof("tcp server start at %v", szAddress)
			return &Server{
				szIP,
				nPort,
				objListener,
				objHandler,
				objConnectionHandler,
			}
		}
	}
}

func (this *Server) Stop() {
	_ = this.objListener.Close()
}

func (this *Server) Start() {
	defer this.objListener.Close()

	go this.objHandler.OnStart(this)

	for {
		// io.EOF
		if objConnection, anyErr := this.objListener.Accept(); anyErr != nil {
			log.Errorf("tcp accept error:%v", anyErr)
			break
		} else {
			mRequestContext := make(map[string]*RpcContext)
			objServerConnection := ServerConnection{
				objConnection,
				this,
				nil,
				new(sync.Mutex),
				this.objConnectionHandler,
				0,
				context.TODO(),
				&mRequestContext,
				&sync.Mutex{},
			}

			go this.objHandler.OnNewConnection(&objServerConnection)
			go objServerConnection.readLoop()
		}
	}

	this.objHandler.OnStop(this)
}

func (this *Server) GetIP() string {
	return this.szIP
}

func (this *Server) GetPort() uint16 {
	return this.nPort
}
