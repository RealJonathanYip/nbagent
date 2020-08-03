package network

import (
	"bytes"
	"context"
	"encoding/binary"
	"git.yayafish.com/nbagent/log"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"git.yayafish.com/nbagent/taskworker"
)

type RpcContext struct {
	Msg        chan []byte
	Ctx        context.Context
	CancelFunc context.CancelFunc
	szErrorMsg string
	objLock    sync.Mutex
}

type ServerConnection struct {
	objConnection    net.Conn
	ptrServer        *Server
	ptrClient        *Client
	ptrWriteLock     *sync.Mutex
	objHandler       ConnectionHandler
	nLastReceiveTime uint32
	Ctx              context.Context
	mRequestContext  *map[string]*RpcContext
	objRWLock        *sync.Mutex
}

type ServerHandler interface {
	OnStart(ptrServer *Server)
	OnStop(ptrServer *Server)
	OnNewConnection(ptrConnection *ServerConnection)
}

type ConnectionHandler interface {
	OnRead(ptrConnection *ServerConnection, byteData []byte, ctx context.Context)
	OnCloseConnection(ptrConnection *ServerConnection)
}

const (
	CODE_SUCCESS = 0
	NETWORK_TASK_FLAG = "network_task_flag"
)

// TODO: this
func (this *RpcContext) setErrorMsg(szErr string) {
	this.objLock.Lock()
	defer this.objLock.Unlock()

	this.szErrorMsg = szErr
}

func (this *RpcContext) ErrorMsg() string {
	this.objLock.Lock()
	defer this.objLock.Unlock()

	return this.szErrorMsg
}

func (this *ServerConnection) cleanAllRpcContext(szReason string) {

	this.objRWLock.Lock()
	defer this.objRWLock.Unlock()

	for szRequestID, ptrRpcContext := range *this.mRequestContext {
		ptrRpcContext.setErrorMsg(szReason)
		ptrRpcContext.CancelFunc()

		close(ptrRpcContext.Msg)
		delete(*this.mRequestContext, szRequestID)
	}
}
func (this *ServerConnection) GetRpcContext(szRequestID string) (*RpcContext, bool) {
	this.objRWLock.Lock()
	defer this.objRWLock.Unlock()

	ptrRpcContext, bSuccess := (*this.mRequestContext)[szRequestID]
	delete(*this.mRequestContext, szRequestID)

	return ptrRpcContext, bSuccess
}

func (this *ServerConnection) RemoveRpcContext(szRequestID string) {
	this.objRWLock.Lock()
	defer this.objRWLock.Unlock()
	if ptrRpcContext, bSuccess := (*this.mRequestContext)[szRequestID]; bSuccess {
		close(ptrRpcContext.Msg)
		delete(*this.mRequestContext, szRequestID)
	}
}

func (this *ServerConnection) AddRpcContext(szRequestID string, nTimeoutMs int64, ctx context.Context) *RpcContext {
	ctxTimeout, fnCancel := context.WithTimeout(ctx, time.Duration(nTimeoutMs)*time.Millisecond)
	objRpcContext := RpcContext{Msg: make(chan []byte), Ctx: ctxTimeout, CancelFunc: fnCancel, szErrorMsg: ""}

	this.objRWLock.Lock()
	defer this.objRWLock.Unlock()

	(*this.mRequestContext)[szRequestID] = &objRpcContext

	return &objRpcContext
}

func (this *ServerConnection) IsTimeout(nNow, nTimeout uint32) bool {
	return atomic.LoadUint32(&this.nLastReceiveTime)+nTimeout > nNow
}

func (this *ServerConnection) Write(data []byte) bool {
	this.ptrWriteLock.Lock()
	defer this.ptrWriteLock.Unlock()

	if nWritedByte, anyErr := this.objConnection.Write(data); anyErr != nil {
		log.Warningf("[%s -> %s]write error :%v", this.objConnection.LocalAddr().String(), this.objConnection.RemoteAddr().String(), anyErr)
		taskworker.TaskWorkerManagerInstance().GoTask(this.Ctx, NETWORK_TASK_FLAG, func(ctx context.Context) int {
			this.cleanAllRpcContext("write err:" + anyErr.Error())
			_ = this.objConnection.Close()

			return CODE_SUCCESS
		})


		return false
	} else {
		log.Debugf("[%s -> %s]write byte :%v", this.objConnection.LocalAddr().String(), this.objConnection.RemoteAddr().String(), nWritedByte)
		return true
	}
}

func (this *ServerConnection) Close() bool {
	taskworker.TaskWorkerManagerInstance().GoTask(this.Ctx, NETWORK_TASK_FLAG, func(ctx context.Context) int {
		this.cleanAllRpcContext("user close connecion:")
		return CODE_SUCCESS
	})

	return this.objConnection.Close() == nil
}

func (this *ServerConnection) readLoop() {
	defer func() {
		if anyErr := recover(); anyErr != nil {
			buf := make([]byte, 1<<20)
			buf = buf[:runtime.Stack(buf, false)]
			log.Errorf("err:%s,static:%s\n", anyErr, buf)
		}

		this.Close()
	}()

	log.Debugf("read from : %s", this.objConnection.RemoteAddr())
	_ = this.objConnection.(*net.TCPConn).SetKeepAlive(true)
	_ = this.objConnection.(*net.TCPConn).SetKeepAlivePeriod(time.Second * 10)
	_ = this.objConnection.SetReadDeadline(time.Time{})

	for {
		var nSize uint32 = 0
		if anyErr := binary.Read(this.objConnection, binary.LittleEndian, &nSize); anyErr != nil {
			if anyErr != io.EOF {
				log.Warningf("connection exception:%s", anyErr)

				taskworker.TaskWorkerManagerInstance().GoTask(this.Ctx, NETWORK_TASK_FLAG, func(ctx context.Context) int {
					this.cleanAllRpcContext("read err:" + anyErr.Error())
					return CODE_SUCCESS
				})
			}
			break
		}

		atomic.StoreUint32(&this.nLastReceiveTime, uint32(time.Now().Unix()))
		if nSize > PACKET_LIMIT {
			log.Warningf("nSize:%v bigger then :%v", nSize, PACKET_LIMIT)

			taskworker.TaskWorkerManagerInstance().GoTask(this.Ctx, NETWORK_TASK_FLAG, func(ctx context.Context) int {
				this.cleanAllRpcContext("unmarshal err: size error")
				return CODE_SUCCESS
			})
			break
		}

		ptrBuffer := bytes.NewBuffer([]byte{})
		objLimitReader := io.LimitReader(this.objConnection, int64(nSize))
		if _, anyErr := ptrBuffer.ReadFrom(objLimitReader); anyErr != nil {
			if anyErr != io.EOF {
				log.Warningf("connection exception:%s", anyErr)

				taskworker.TaskWorkerManagerInstance().GoTask(this.Ctx, NETWORK_TASK_FLAG, func(ctx context.Context) int {
					this.cleanAllRpcContext("read err:" + anyErr.Error())
					return CODE_SUCCESS
				})
			}
			break
		}

		ctx := context.WithValue(context.TODO(), REQUEST_CONTEXT,
			newContext(this, ""))
		this.objHandler.OnRead(this, ptrBuffer.Bytes(), ctx)
	}

	this.Close()

	this.objHandler.OnCloseConnection(this)

	log.Infof("sc:[%s->%s] close", this.objConnection.LocalAddr().String(), this.objConnection.RemoteAddr().String())
}

func (this *ServerConnection) LocalAddr() string {
	return this.objConnection.LocalAddr().String()
}

func (this *ServerConnection) RemoteAddr() string {
	return this.objConnection.RemoteAddr().String()
}

func (this *ServerConnection) SetRequestContext(ptrRequestContext *map[string]*RpcContext) {
	this.mRequestContext = ptrRequestContext
}

func (this *ServerConnection) SetRWLock(ptrRWLock *sync.Mutex) {
	this.objRWLock = ptrRWLock
}
