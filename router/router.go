package router

import (
	"context"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/reporter"
	"io"
	"git.yayafish.com/nbagent/log"
	"bytes"
	"sync"
	"git.yayafish.com/nbagent/protocol"
	"encoding/binary"
	"time"
)

type Router struct {
	mRouter   map[string]func(byteData []byte, ctx context.Context) string
	objRWLock sync.RWMutex
}

type packetReader struct {
	buffer *bytes.Buffer
}

func (this *packetReader) Bytes() []byte {
	if this.buffer != nil {
		return this.buffer.Bytes()
	}
	return []byte{}
}

func (this *packetReader) Read(b []byte) (int, error) {
	if this.buffer != nil {
		return this.buffer.Read(b)
	}
	return 0, io.EOF
}

func (this *packetReader) GetRest(nFrom int) []byte {
	return this.buffer.Next(nFrom)
}

func NewRouter() *Router {
	return &Router{
		make(map[string]func(byteData []byte, ctx context.Context) string),
		sync.RWMutex{},
	}
}

func (this *Router) Process(byteData []byte, ctx context.Context) bool {
	ptrContext := network.GetRequestContext(ctx)
	if ptrContext == nil {
		log.Errorf("ptrContext is nil!")
		return false
	}

	var nMsgType uint8
	objReader := packetReader{bytes.NewBuffer(byteData)}
	if szUri, bSuccess := protocol.ReadString(&objReader); !bSuccess {
		reporter.Result("router_read_uri", "read_uri_fail")
		log.Warningf("read uri fail!")
		return false
	} else if anyErr := binary.Read(&objReader, binary.LittleEndian, &nMsgType); anyErr != nil {
		reporter.Result("router_read_uri", "read_binary_fail:" + anyErr.Error())
		log.Warningf("read nMsgType fail!")
		return false
	} else if szRequestID, bSuccess := protocol.ReadString(&objReader); !bSuccess {
		reporter.Result("router_read_uri", "read_request_id_fail")
		log.Warningf("read szRequestID fail!")
		return false
	} else {
		network.SetRequestID(ctx, szRequestID)
		fnRouter2Handler := func () {
			this.objRWLock.RLock()
			if fnHandler, bExist := this.mRouter[szUri]; !bExist {
				this.objRWLock.RUnlock()
				reporter.Result("router_read_uri", "uri_nod_found")
				log.Warningf("router not found of uri:%v nMsgType:%v", szUri, nMsgType)
			} else {
				this.objRWLock.RUnlock()
				objNow := time.Now()
				//同一链接收发包尽量保证线性，不放协程池，在应用层做并发
				szResult := fnHandler(objReader.Bytes(), ctx)
				reporter.Result("router_read_uri", "success")
				reporter.Result("uri_" + szUri, "success", time.Since(objNow).Milliseconds())
				log.Debugf("uri:%v result:%v", szUri, szResult)
			}
		}

		switch nMsgType {
		case protocol.MSG_TYPE_CAST,protocol.MSG_TYPE_REQUEST:
			fnRouter2Handler()
			break
		case protocol.MSG_TYPE_RESPONE:
			ptrServerConnection := network.GetServerConnection(ctx)
			ptrRpcContext, bSuccess := ptrServerConnection.GetRpcContext(szRequestID)

			if bSuccess {
				ptrRpcContext.Msg <- objReader.Bytes()
			} else {
				fnRouter2Handler()
			}
			break
		default:
			log.Warningf("invalidate msg type:%v", nMsgType)
			return false
		}
	}

	return true
}


func (this *Router) RegisterRouter(szUri string, fnHandler func(byteData []byte, ctx context.Context) string) bool {
	this.objRWLock.Lock()
	defer this.objRWLock.Unlock()

	if _, bExist := this.mRouter[szUri]; bExist {
		log.Errorf("repeat register uri:%v", szUri)
		return false
	} else {
		this.mRouter[szUri] = fnHandler
		return true
	}
}
