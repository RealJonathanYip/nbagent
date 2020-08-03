// auto-generate by `gen` don't edit

package agent

import (
	"context"

	"git.yayafish.com/nbagent/client"
	//"github.com/micro/protobuf/proto"
)

//client call
    
func NewPBAgentClient(ptrClient *client.Client) PBAgentClient {
	return PBAgentClient{
		ptrClient: ptrClient,
	}
}

type PBAgentClient struct {
	ptrClient *client.Client
}
func (ptrClient *PBAgentClient) CallRegister(szUri string, nTimeoutMs int64,
    ptrReq *AgentRegisterReq, ptrRsp *AgentRegisterRsp) error {
	return ptrClient.ptrClient.SyncCall(szUri, ptrReq, ptrRsp, nTimeoutMs)
}
func (ptrClient *PBAgentClient) CallKeepAlive(szUri string, nTimeoutMs int64,
    ptrReq *AgentKeepAliveRsp, ptrRsp *AgentKeepAliveRsp) error {
	return ptrClient.ptrClient.SyncCall(szUri, ptrReq, ptrRsp, nTimeoutMs)
}
func (ptrClient *PBAgentClient) CallEntryCheck(szUri string, nTimeoutMs int64,
    ptrReq *AgentEntryCheckReq, ptrRsp *AgentEntryCheckRsp) error {
	return ptrClient.ptrClient.SyncCall(szUri, ptrReq, ptrRsp, nTimeoutMs)
}

//service call
    
func NewPBAgentServiceHandler(ptrClient *client.Client) PBAgentServiceHandler {
	return PBAgentServiceHandler{
		ptrClient: ptrClient,
	}
}

type PBAgentServiceHandler struct {
	ptrClient *client.Client
}

func (ptrService *PBAgentServiceHandler) RegisterServiceHandler(szUri string,
 fnHandler func(byteData []byte, ctx context.Context) []byte) error {
	return ptrService.ptrClient.HandleRpc(szUri, fnHandler)
}