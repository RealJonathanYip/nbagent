package cli_agent

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"git.yayafish.com/nbagent/config"
	"math/rand"
	"strconv"
	"time"
)

const (
	NODE_INFO_KEY                = "NODE_INFO_KEY"
	CODE_SUCCESS                 = 0
	CODE_NODE_NOT_FOUND          = 101
	CODE_NODE_LOCAL_INVOKE       = 102
	CODE_REQUEST_MODE_INVALIDATE = 103

	TASK_WORKER_TAG_NODE_INTRODUCE = "node_introduce"
	TASK_WORKER_TAG_NODE_MANAGER   = "node_manager"
	TASK_WORKER_TAG_NODE_CONNECT   = "node_connect"
	TASK_WORKER_TAG_TCP_MESSAGE    = "tcp_msg"
	TASK_WORKER_TAG_AGENT_MANAGER  = "agent_manager"
)

var (
	AgentClientMaxConnection int
)

func init() {
	rand.Seed(time.Now().UnixNano())

	AgentClientMaxConnection = 10
	if config.ServerConf.AgentConf.ClientMaxConnection > 0 {
		AgentClientMaxConnection = config.ServerConf.AgentConf.ClientMaxConnection
	}
}

func makeSign(szID string, nNow int64) string {
	objMac := hmac.New(md5.New, []byte(config.ServerConf.SecretKey))
	objMac.Write([]byte(szID + strconv.FormatInt(nNow, 10)))

	return hex.EncodeToString(objMac.Sum(nil))
}
