package cmd

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/protocol"
	"git.yayafish.com/nbagent/protocol/agent"
	"git.yayafish.com/nbagent/rpc"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Discovery register",

	Run: registerFunc,
}

var (
	nUriSize       int
	nUriSpace      int
	nRegisterTotal int

	szAgentSecretKey string
	aryAgentInfo     []AgentPoint

	szRegisterClientName = "RegisterClient"
	szRegisterUriName    = "RegisterURI"
)

type AgentPoint struct {
	Ip   string
	Port uint16
}

func init() {
	rand.Seed(time.Now().UnixNano())

	RootCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringVar(&szAgentSecretKey, "secret-key", "node", "secret key size of agent")
	registerCmd.Flags().IntVar(&nUriSize, "uri-size", 256, "URI size of register request")
	registerCmd.Flags().IntVar(&nUriSpace, "uri-space", 1, "URI space of register request")
	registerCmd.Flags().IntVar(&nRegisterTotal, "total-register", 1, "Total number of register requests")
}

func registerFunc(cmd *cobra.Command, args []string) {

	fmt.Printf("agents: %v \n", szAgents)
	fmt.Printf("clients: %v \n", nTotalClients)
	fmt.Printf("call-timeout: %v \n", nCallTimeout)
	fmt.Printf("only-agent: %v \n", bOnlyAgent)
	fmt.Printf("secret-key: %v \n", szAgentSecretKey)
	fmt.Printf("uri-size: %v \n", nUriSize)
	fmt.Printf("uri-space: %v \n", nUriSpace)
	fmt.Printf("total-register: %v \n", nRegisterTotal)

	aryEndPoint := strings.Split(szAgents, ",")
	aryAgentInfo = []AgentPoint{}
	for _, szPoint := range aryEndPoint {
		aryIpPort := strings.Split(szPoint, ":")
		log.Infof("agent: %v", aryIpPort)
		if len(aryIpPort) != 2 {
			panic("agent point error")
		}
		objAgentPoint := AgentPoint{}
		objAgentPoint.Ip = aryIpPort[0]
		nPort, _ := strconv.ParseUint(aryIpPort[1], 10, 64)
		objAgentPoint.Port = uint16(nPort)
		aryAgentInfo = append(aryAgentInfo, objAgentPoint)
	}

	aryReqEntry := make(chan agent.EntryInfo, nTotalClients)
	var aryClient []*RegisterRequestClient
	var nCount uint = 0
	for nCount = 0; nCount < nTotalClients; nCount++ {

		var objAgentPoint AgentPoint
		if bOnlyAgent {
			objAgentPoint = aryAgentInfo[0]
		} else {
			objAgentPoint = aryAgentInfo[rand.Intn(len(aryAgentInfo))]
		}

		szInstanceID := GenerateInstanceID(szRegisterClientName, int(nCount))
		ptrClient := NewRegisterClient(objAgentPoint.Ip, objAgentPoint.Port, szInstanceID)
		ptrClient.Start()
		aryClient = append(aryClient, ptrClient)
	}

	objReport := RequestReport{}
	objReport.Reset()

	for i := range aryClient {
		objWaitGroup.Add(1)

		go func(ptrClient *RegisterRequestClient) {
			defer objWaitGroup.Done()

			for objEntry := range aryReqEntry {

				objStartTime := time.Now()
				bRet := ptrClient.DoRegister(objEntry)
				nDuration := time.Since(objStartTime)
				var anyErr error
				if !bRet {
					anyErr = fmt.Errorf("DoRegister fail")
				}
				objReport.Result(nDuration, anyErr)
				log.Infof("DoRegister use time: %v", nDuration)
			}
		}(aryClient[i])
	}

	go func() {
		for i := 0; i < nRegisterTotal; i++ {
			var objEntry agent.EntryInfo
			objEntry.URI = GenerateRegisterURIWithSpace(szRegisterUriName, i, nUriSize, nUriSpace)
			objEntry.EntryType = agent.EntryType_RPC

			aryReqEntry <- objEntry
		}
		close(aryReqEntry)
	}()

	objWaitGroup.Wait()

	objStats := objReport.Stats()
	fmt.Printf("total    requests    : %v \n", objStats.TotalReq)
	fmt.Printf("total    time        : %v \n", objStats.TotalTime)
	fmt.Printf("error    requests    : %v \n", objStats.ErrorNum)
	fmt.Printf("success  requests    : %v \n", objStats.SuccessNum)
	fmt.Printf("average  QPS         : %.2f \n", objStats.ReqPerSecond)
	fmt.Printf("min      latency     : %.2f ms \n", objStats.Fastest)
	fmt.Printf("max      latency     : %.2f ms \n", objStats.Slowest)
	fmt.Printf("average  latency     : %.2f ms \n", objStats.AvgTotal)
}

type RegisterConnectionHandler struct{}

func (objHandler RegisterConnectionHandler) OnRead(ptrConnection *network.ServerConnection, byteData []byte,
	ctx context.Context) {

	var nMsgType uint8
	objReader := bytes.NewBuffer(byteData)
	if _, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("read uri fail!")
		return
	} else if anyErr := binary.Read(objReader, binary.LittleEndian, &nMsgType); anyErr != nil {
		log.Warningf("read nMsgType fail!")
		return
	} else if szRequestID, bSuccess := protocol.ReadString(objReader); !bSuccess {
		log.Warningf("read szRequestID fail!")
		return
	} else {
		if nMsgType == protocol.MSG_TYPE_RESPONE {
			ptrServerConnection := network.GetServerConnection(ctx)
			ptrRpcContext, bSuccess := ptrServerConnection.GetRpcContext(szRequestID)
			if bSuccess {
				ptrRpcContext.Msg <- objReader.Bytes()
			} else {
				log.Warningf("can not find response context szRequestID:%v", szRequestID)
			}
		}
	}
}
func (objHandler RegisterConnectionHandler) OnCloseConnection(ptrConnection *network.ServerConnection) {
}

type RegisterRequestClient struct {
	ptrConn      *network.ServerConnection
	szIP         string
	nPort        uint16
	szInstanceID string
}

func NewRegisterClient(szIP string, nPort uint16, szInstanceID string) *RegisterRequestClient {

	var ptrRegisterRequest *RegisterRequestClient
	ptrRegisterRequest = &RegisterRequestClient{
		szIP:         szIP,
		nPort:        nPort,
		szInstanceID: szInstanceID,
	}
	return ptrRegisterRequest
}

func (ptrRequest *RegisterRequestClient) Start() bool {

	var bRet bool = false
	var ptrConn *network.ServerConnection
	bRet, ptrConn = network.NewClient(ptrRequest.szIP, ptrRequest.nPort, RegisterConnectionHandler{})
	if !bRet {
		log.Errorf("network.NewClient error")
		return bRet
	}

	ptrRequest.ptrConn = ptrConn
	ptrRequest.initRegister()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			ptrRequest.keepAlive()
		}
	}()
	return bRet
}

func (ptrRequest *RegisterRequestClient) DoRegister(objEntry agent.EntryInfo) bool {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCall *rpc.CallContext
	var objReq agent.AgentRegisterReq = agent.AgentRegisterReq{}
	objReq.AryEntry = append(objReq.AryEntry, &objEntry)
	objReq.InstanceID = ptrRequest.szInstanceID
	objReq.TimeStamp = nTimeNow
	objReq.Sign = ptrRequest.makeSign(objReq.InstanceID, objReq.TimeStamp)
	var objRsp agent.AgentRegisterRsp = agent.AgentRegisterRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_REGISTER, &objReq, &objRsp, ptrRequest.ptrConn)
	ptrCall.Timeout(int64(nCallTimeout) * 1000)
	bRet = ptrCall.Start()
	if !bRet {
		log.Warningf("rpc call start fail")
	}
	return bRet
}

func (ptrRequest *RegisterRequestClient) initRegister() {

	var bRet bool = false
	var nTimeNow int64 = time.Now().Unix()
	var ptrCall *rpc.CallContext
	var objReq agent.AgentRegisterReq = agent.AgentRegisterReq{}
	objReq.InstanceID = ptrRequest.szInstanceID
	objReq.TimeStamp = nTimeNow
	objReq.Sign = ptrRequest.makeSign(objReq.InstanceID, objReq.TimeStamp)
	var objRsp agent.AgentRegisterRsp = agent.AgentRegisterRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_REGISTER, &objReq, &objRsp, ptrRequest.ptrConn)
	ptrCall.Timeout(int64(nCallTimeout) * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_REGISTER Req: %v, Rsp: %v", objReq, objRsp)
	}
}

func (ptrRequest *RegisterRequestClient) keepAlive() {

	var bRet bool = false
	var objReq agent.AgentKeepAliveNotify = agent.AgentKeepAliveNotify{}
	var ptrCall *rpc.CallContext
	var objRsp agent.AgentKeepAliveRsp = agent.AgentKeepAliveRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_KEEP_ALIVE, &objReq, &objRsp, ptrRequest.ptrConn)
	ptrCall.Timeout(int64(nCallTimeout) * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_KEEP_ALIVE success, Rsp: %+v", objRsp)
		//fmt.Printf("AGENT_KEEP_ALIVE success, Rsp: %+v \n", objRsp)
	} else {
		log.Warningf("AGENT_KEEP_ALIVE error,Req: %v", objReq)
	}
}

func (ptrRequest *RegisterRequestClient) makeSign(szID string, nNow int64) string {
	objMac := hmac.New(md5.New, []byte(szAgentSecretKey))
	objMac.Write([]byte(szID + strconv.FormatInt(nNow, 10)))

	return hex.EncodeToString(objMac.Sum(nil))
}
