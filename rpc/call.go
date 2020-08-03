package rpc

import (
	"context"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

type CallContext struct {
	szUri               string
	szRequestID         string
	objReq              proto.Message
	objRsp              proto.Message
	nTimeoutMs          int64
	ptrServerConnection *network.ServerConnection
	ctx                 context.Context
}

const (
	DEFAULT_TIMEOUT = 2000
)

func NewCall(szUri string, objReq proto.Message, objRsp proto.Message, ptrServerConnection *network.ServerConnection) *CallContext {
	objUUID := uuid.New()
	return &CallContext{
		szUri, objUUID.String(),
		objReq, objRsp,
		DEFAULT_TIMEOUT, ptrServerConnection, context.Background(),
	}
}

func (this *CallContext) Timeout(nTimeoutMs int64) *CallContext {
	this.nTimeoutMs = nTimeoutMs
	return this
}

func (this *CallContext) RequestID(szRequestID string) *CallContext {
	this.szRequestID = szRequestID
	return this
}

func (this *CallContext) Context(ctx context.Context) *CallContext {
	this.ctx = ctx
	return this
}

func (this *CallContext) Start() bool {
	if this.ptrServerConnection == nil {
		log.Warningf("ptrServerConnection not available szUri:%v szRequestID:%v", this.szUri, this.szRequestID)
		return false
	}

	log.Debugf("uri:%v, req_id:%v, starting", this.szUri, this.szRequestID)

	byteData, _ := protocol.NewProtocol(this.szUri, protocol.MSG_TYPE_REQUEST,
		this.szRequestID, this.objReq)

	ptrRpcContext := this.ptrServerConnection.AddRpcContext(this.szRequestID, this.nTimeoutMs, this.ctx)

	if !this.ptrServerConnection.Write(byteData) {
		return false
	}

	select {
	case byteResult := <-ptrRpcContext.Msg:
		if  ptrRpcContext.ErrorMsg() != "" {
			log.Warningf("call error %v szUri:%v szRequestID:%v", ptrRpcContext.ErrorMsg(), this.szUri, this.szRequestID)
			return false
		} else {
			ptrRpcContext.CancelFunc()
			if anyErr := proto.Unmarshal(byteResult, this.objRsp); anyErr != nil {
				log.Warningf("unmarshal data fail:%v szUri:%v szRequestID:%v", anyErr, this.szUri, this.szRequestID)
				//TODO:report
				return false
			} else {
				return true
			}
		}
	case <-ptrRpcContext.Ctx.Done():
		if ptrRpcContext.ErrorMsg() == "" {
			//只有超时才需要手动删除上下文，其他情况会自动删除
			this.ptrServerConnection.RemoveRpcContext(this.szRequestID)
			log.Warningf("call timeout szUri:%v szRequestID:%v", this.szUri, this.szRequestID)
		} else {
			log.Warningf("call error %v szUri:%v szRequestID:%v", ptrRpcContext.ErrorMsg(), this.szUri, this.szRequestID)
		}
		return false
	}
}
