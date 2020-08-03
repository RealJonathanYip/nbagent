package rpc

import (
	"context"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

type CastContext struct {
	szUri               string
	szRequestID         string
	objMessage          proto.Message
	ptrServerConnection *network.ServerConnection
	ctx                 context.Context
}

func NewCast(szUri string, objReq proto.Message, ptrServerConnection *network.ServerConnection) *CastContext {
	objUUID := uuid.New()

	return &CastContext{
		szUri, objUUID.String(),
		objReq, ptrServerConnection, context.Background(),
	}
}

func (this *CastContext) RequestID(szRequestID string) *CastContext {
	this.szRequestID = szRequestID
	return this
}

func (this *CastContext) Context(ctx context.Context) *CastContext {
	this.ctx = ctx
	return this
}

func (this *CastContext) Start() bool {
	if this.ptrServerConnection == nil {
		log.Warningf("ptrServerConnection not available szUri:%v szRequestID:%v", this.szUri, this.szRequestID)
		return false
	}

	byteData, _ := protocol.NewProtocol(this.szUri, protocol.MSG_TYPE_CAST,
		this.szRequestID, this.objMessage)

	return  this.ptrServerConnection.Write(byteData)
}
