package node

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/protocol/node"
	"git.yayafish.com/nbagent/taskworker"
	"math/rand"
	"strconv"
	"time"
)

const (
	NODE_INFO_KEY = "NODE_INFO_KEY"
	CODE_SUCCESS  = 0

	TASK_WORKER_TAG_NODE_INTRODUCE = "node_introduce"
	TASK_WORKER_TAG_NODE_MANAGER   = "node_manager"
	TASK_WORKER_TAG_NODE_CONNECT   = "node_connect"
	TASK_WORKER_TAG_TCP_MESSAGE    = "tcp_msg"
)

var (
	NeighbourMaxConnection int
)

func init() {
	rand.Seed(time.Now().UnixNano())
	taskworker.TaskWorkerManagerInstance().Init(config.ServerConf.WorkerConfigs.Workers)

	NeighbourMaxConnection = 10
	if config.ServerConf.NodeConf.NeighbourMaxConnection > 0 {
		NeighbourMaxConnection = config.ServerConf.NodeConf.NeighbourMaxConnection
	}
}

func MyInfo(ctx context.Context) *node.NodeInfo {
	return ctx.Value(NODE_INFO_KEY).(*node.NodeInfo)
}

func MyName(ctx context.Context) string {
	return ctx.Value(NODE_INFO_KEY).(*node.NodeInfo).Name
}

func MyIP(ctx context.Context) string {
	return ctx.Value(NODE_INFO_KEY).(*node.NodeInfo).IP
}

func MyPort(ctx context.Context) uint32 {
	return ctx.Value(NODE_INFO_KEY).(*node.NodeInfo).Port
}

func makeSign(szID string, nNow int64) string {
	objMac := hmac.New(md5.New, []byte(config.ServerConf.SecretKey))
	objMac.Write([]byte(szID + strconv.FormatInt(nNow, 10)))

	return hex.EncodeToString(objMac.Sum(nil))
}
