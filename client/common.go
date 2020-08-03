package client

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"git.yayafish.com/nbagent/network"
	"strconv"
	"sync"
	"time"
)

const (
	CONTEXT_KEY_NODE       = "CONTEXT_KEY_AGENT"

	NETWORK_TASK_FLAG 			= "network_task_flag"
	TASK_WORKER_TAG_TCP_MESSAGE = "tcp_msg"

	CODE_SUCCESS = 0
)

type InstanceID string
func (objInstanceID InstanceID) String() string {
	return fmt.Sprintf("instanceID:%v", objInstanceID.String())
}

type AgentInfo struct {
	//Name string
	IP   string
	Port uint16
}

type AgentContext struct {
	ptrRequestContext *map[string]*network.RpcContext
	ptrRWLock         *sync.Mutex
}

type ConnectionInfo struct {
	LastKeepaliveTime time.Time
	Connection        *network.ServerConnection
	Host              string // 用于知道目前有哪些host有连接
}

func makeSign(szID string, nNow int64, szSecretKey string) string {
	objMac := hmac.New(md5.New, []byte(szSecretKey))
	objMac.Write([]byte(szID + strconv.FormatInt(nNow, 10)))

	return hex.EncodeToString(objMac.Sum(nil))
}

func ConnectToString(ptrConnection *network.ServerConnection) string {
	return ptrConnection.RemoteAddr() + "-" + ptrConnection.LocalAddr()
}

func makeHostKeyByIpPort(szIP string, nPort uint16) string {
	return szIP + ":" + strconv.FormatInt(int64(nPort), 10)
}
