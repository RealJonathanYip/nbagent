// auto-generate by `gen` don't edit

package demo

import (
	"context"

	"git.yayafish.com/nbagent/client"
	//"github.com/micro/protobuf/proto"
)

//client call
    
func NewPBDemoClient(ptrClient *client.Client) PBDemoClient {
	return PBDemoClient{
		ptrClient: ptrClient,
	}
}

type PBDemoClient struct {
	ptrClient *client.Client
}
func (ptrClient *PBDemoClient) CallTestHaha(szUri string, nTimeoutMs int64,
    ptrReq *TestMsgReq, ptrRsp *TestMsgRsp) error {
	return ptrClient.ptrClient.SyncCall(szUri, ptrReq, ptrRsp, nTimeoutMs)
}

//service call
    
func NewPBDemoServiceHandler(ptrClient *client.Client) PBDemoServiceHandler {
	return PBDemoServiceHandler{
		ptrClient: ptrClient,
	}
}

type PBDemoServiceHandler struct {
	ptrClient *client.Client
}

func (ptrService *PBDemoServiceHandler) RegisterServiceHandler(szUri string,
 fnHandler func(byteData []byte, ctx context.Context) []byte) error {
	return ptrService.ptrClient.HandleRpc(szUri, fnHandler)
}