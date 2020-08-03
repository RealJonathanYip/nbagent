package common

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

var (
	AgentNameSrc string = "agent_src"
	AgentNameDest string = "agent_dest"

	RPC_URI_DEMO string = "rpc.demo"

	AgentSecretKey string = "node"
	AgentIp string = "127.0.0.1"
	AgentPort uint16 = 8800
)

func MakeSign(szID string, nNow int64) string {
	objMac := hmac.New(md5.New, []byte(AgentSecretKey))
	objMac.Write([]byte(szID + strconv.FormatInt(nNow, 10)))

	return hex.EncodeToString(objMac.Sum(nil))
}