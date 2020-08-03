package node

import (
	"context"

	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/protocol/node"
)

type AgentHandler interface {
	HandleRequest(ptrReq *node.RpcCallReq, ctx context.Context)
	HandleResponse(ptrResp *node.RpcCallResp, ctx context.Context)
}

type DummyAgentHandler struct{}

func (ptrDummyNodeHandler *DummyAgentHandler) HandleRequest(ptrReq *node.RpcCallReq, ctx context.Context) {
	log.Warningf("DummyAgentHandler HandleRequest")
}

func (ptrDummyNodeHandler *DummyAgentHandler) HandleResponse(ptrResp *node.RpcCallResp, ctx context.Context) {
	log.Warningf("DummyAgentHandler HandleResponse")
}
