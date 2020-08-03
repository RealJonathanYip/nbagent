package rpc

import (
	"context"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"github.com/golang/protobuf/proto"
)

func Reply(objMsg proto.Message, ctx context.Context) {
	if byteByte, bSuccess := protocol.NewProtocol("",
		protocol.MSG_TYPE_RESPONE, network.GetRequestID(ctx), objMsg); bSuccess {
		network.GetServerConnection(ctx).Write(byteByte)
	}
}
