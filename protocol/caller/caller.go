package caller

import (
	"encoding/json"
	"git.yayafish.com/nbagent/protocol/node"
)

const (
	Caller_Agent int32 = 1
	Caller_Node  int32 = 2
)

type Caller struct {
	Type int32 `json:"nType"`
	AgentCaller
	NodeCaller
}

type AgentCaller struct {
	AgentID      string `json:"szAgentID"`
	ConnectionID string `json:"szConnectionID"`
	RequestID    string `json:"szRequestID"`
}

type NodeCaller struct {
	Node string `json:"szNode"`
}

func PushStack(ptrReq *node.RpcCallReq, szStackValue string) {
	ptrReq.Caller = append(ptrReq.Caller, szStackValue)
}

func PopStack(ptrResp *node.RpcCallResp) (string, bool) {
	if len(ptrResp.Caller) == 0 {
		return "", false
	}

	szTail := ptrResp.Caller[len(ptrResp.Caller)-1]
	ptrResp.Caller = ptrResp.Caller[:len(ptrResp.Caller)-1]

	return szTail, true
}

func GetTopStack(ptrResp *node.RpcCallResp) (string, bool) {
	if len(ptrResp.Caller) == 0 {
		return "", false
	}

	return ptrResp.Caller[len(ptrResp.Caller)-1], true
}

func GenNodeCaller(szNode string) string {

	var objCaller Caller
	objCaller.Type = Caller_Node
	objCaller.Node = szNode

	byteCaller, _ := json.Marshal(&objCaller)
	return string(byteCaller)
}

//todo Id-->id //done
func GenAgentCaller(szAgentID, szRequestID, szConnectionID string) string {

	var objCaller Caller
	objCaller.Type = Caller_Agent
	objCaller.AgentID = szAgentID
	objCaller.RequestID = szRequestID
	objCaller.ConnectionID = szConnectionID

	byteCaller, _ := json.Marshal(&objCaller)
	return string(byteCaller)
}
