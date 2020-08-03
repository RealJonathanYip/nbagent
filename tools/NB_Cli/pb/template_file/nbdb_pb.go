package template_file

import (
	"git.yayafish.com/nbagent/client"
	"github.com/micro/protobuf/proto"
)

type DemoSrvService struct {
	ptrClient *client.Client
}

func NewDemoSrvService(ptrClient *client.Client) DemoSrvService {
	return DemoSrvService{
		ptrClient: ptrClient,
	}
}

func (ptrService *DemoSrvService) TestHaha(szUri string, nTimeoutMs int64, ptrReq, ptrRsp proto.Message) error {

	return ptrService.ptrClient.SyncCall(szUri, ptrReq, ptrRsp, nTimeoutMs)
}
