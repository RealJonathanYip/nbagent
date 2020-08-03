package cli_agent

import (
	"context"

	"git.yayafish.com/nbagent/log"

	"git.yayafish.com/nbagent/protocol/node"
)

type NodeHandler interface {
	DispatchReq(ptrReq *node.RpcCallReq, ctx context.Context) node.ResultCode
	DispatchRsp(ptrResp *node.RpcCallResp, ctx context.Context) node.ResultCode
	AddEntry(szUri string, enumEntryType node.EntryType)
	RemoveEntry(szUri string, enumEntryType node.EntryType)
	NodeList() []*node.NodeInfo
	EntryCheck(szEntryUri string, nEntryType node.EntryType) node.ResultCode
}

type DummyNodeHandler struct{}

func (ptrDummyNodeHandler *DummyNodeHandler) DispatchReq(ptrReq *node.RpcCallReq, ctx context.Context) node.ResultCode {
	log.Warningf("DummyNodeHandler DispatchReq")
	return node.ResultCode_ERROR
}

func (ptrDummyNodeHandler *DummyNodeHandler) DispatchRsp(ptrResp *node.RpcCallResp, ctx context.Context) node.ResultCode {
	log.Warningf("DummyNodeHandler DispatchRsp")
	return node.ResultCode_ERROR
}

func (ptrDummyNodeHandler *DummyNodeHandler) AddEntry(szUri string, enumEntryType node.EntryType) {
	log.Warningf("DummyNodeHandler AddEntry")
}

func (ptrDummyNodeHandler *DummyNodeHandler) RemoveEntry(szUri string, enumEntryType node.EntryType) {
	log.Warningf("DummyNodeHandler RemoveEntry")
}

func (ptrDummyNodeHandler *DummyNodeHandler) NodeList() []*node.NodeInfo {
	log.Warningf("DummyNodeHandler NodeList")
	return []*node.NodeInfo{}
}

func (ptrDummyNodeHandler *DummyNodeHandler) EntryCheck(szEntryUri string, nEntryType node.EntryType) node.ResultCode {
	log.Warningf("DummyNodeHandler EntryCheck")
	return node.ResultCode_ERROR
}
