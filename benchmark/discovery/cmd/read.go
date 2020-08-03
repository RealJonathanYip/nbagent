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

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Discovery read",

	Run: readFunc,
}

var (
	nReadUriSize int
	nReadTotal   int

	szReadSecretKey  string
	aryReadAgentInfo []AgentPoint

	szReadClientName = "RegisterClient"
	szReadUriName    = "RegisterURI"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	RootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVar(&szReadSecretKey, "secret-key", "node", "secret key size of agent")
	readCmd.Flags().IntVar(&nReadUriSize, "uri-size", 256, "URI size of register request")
	readCmd.Flags().IntVar(&nReadTotal, "total-read", 1, "Total number of register requests")
}

func readFunc(cmd *cobra.Command, args []string) {

	fmt.Printf("agents: %v \n", szAgents)
	fmt.Printf("clients: %v \n", nTotalClients)
	fmt.Printf("call-timeout: %v \n", nCallTimeout)
	fmt.Printf("only-agent: %v \n", bOnlyAgent)
	fmt.Printf("secret-key: %v \n", szReadSecretKey)
	fmt.Printf("uri-size: %v \n", nReadUriSize)
	fmt.Printf("total-read: %v \n", nReadTotal)

	aryEndPoint := strings.Split(szAgents, ",")
	aryReadAgentInfo = []AgentPoint{}
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
		aryReadAgentInfo = append(aryReadAgentInfo, objAgentPoint)
	}

	aryReqEntry := make(chan agent.EntryInfo, nTotalClients)
	var aryClient []*ReadRequestClient
	var nCount uint = 0
	for nCount = 0; nCount < nTotalClients; nCount++ {

		var objAgentPoint AgentPoint
		if bOnlyAgent {
			objAgentPoint = aryReadAgentInfo[0]
		} else {
			objAgentPoint = aryReadAgentInfo[rand.Intn(len(aryReadAgentInfo))]
		}

		szInstanceID := GenerateInstanceID(szReadClientName, int(nCount))
		ptrClient := NewReadClient(objAgentPoint.Ip, objAgentPoint.Port, szInstanceID)
		ptrClient.Start()
		aryClient = append(aryClient, ptrClient)
	}

	objReport := RequestReport{}
	objReport.Reset()

	for i := range aryClient {
		objWaitGroup.Add(1)

		go func(ptrClient *ReadRequestClient) {
			defer objWaitGroup.Done()

			for objEntry := range aryReqEntry {

				objStartTime := time.Now()
				bRet := ptrClient.DoRead(objEntry)
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
		for i := 0; i < nReadTotal; i++ {
			var objEntry agent.EntryInfo
			objEntry.URI = GenerateRegisterURI(szReadUriName, i, nReadUriSize)
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

type ReadConnectionHandler struct{}

func (objHandler ReadConnectionHandler) OnRead(ptrConnection *network.ServerConnection, byteData []byte,
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
func (objHandler ReadConnectionHandler) OnCloseConnection(ptrConnection *network.ServerConnection) {
}

type ReadRequestClient struct {
	ptrConn      *network.ServerConnection
	szIP         string
	nPort        uint16
	szInstanceID string
}

func NewReadClient(szIP string, nPort uint16, szInstanceID string) *ReadRequestClient {

	var ptrReadRequestClient *ReadRequestClient
	ptrReadRequestClient = &ReadRequestClient{
		szIP:         szIP,
		nPort:        nPort,
		szInstanceID: szInstanceID,
	}
	return ptrReadRequestClient
}

func (ptrRequest *ReadRequestClient) Start() bool {

	var bRet bool = false
	var ptrConn *network.ServerConnection
	bRet, ptrConn = network.NewClient(ptrRequest.szIP, ptrRequest.nPort, ReadConnectionHandler{})
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

func (ptrRequest *ReadRequestClient) DoRead(objEntry agent.EntryInfo) bool {

	var bRet bool = false
	var ptrCall *rpc.CallContext
	var objReq agent.AgentEntryCheckReq = agent.AgentEntryCheckReq{}
	objReq.URI = objEntry.URI
	objReq.EntryType = objEntry.EntryType
	var objRsp agent.AgentEntryCheckRsp = agent.AgentEntryCheckRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_ENTRY_CHECK, &objReq, &objRsp, ptrRequest.ptrConn)
	ptrCall.Timeout(int64(nCallTimeout) * 1000)
	bRet = ptrCall.Start()
	if !bRet {
		log.Warningf("rpc call start fail")
	}
	return bRet
}

func (ptrRequest *ReadRequestClient) initRegister() {

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

func (ptrRequest *ReadRequestClient) keepAlive() {

	var bRet bool = false
	var objReq agent.AgentKeepAliveNotify = agent.AgentKeepAliveNotify{}
	var ptrCall *rpc.CallContext
	var objRsp agent.AgentKeepAliveRsp = agent.AgentKeepAliveRsp{}

	ptrCall = rpc.NewCall(rpc.AGENT_KEEP_ALIVE, &objReq, &objRsp, ptrRequest.ptrConn)
	ptrCall.Timeout(int64(nCallTimeout) * 1000)
	bRet = ptrCall.Start()
	if bRet {
		log.Infof("AGENT_KEEP_ALIVE success, Rsp: %+v", objRsp)
	} else {
		log.Warningf("AGENT_KEEP_ALIVE error,Req: %v", objReq)
	}
}

func (ptrRequest *ReadRequestClient) makeSign(szID string, nNow int64) string {
	objMac := hmac.New(md5.New, []byte(szReadSecretKey))
	objMac.Write([]byte(szID + strconv.FormatInt(nNow, 10)))

	return hex.EncodeToString(objMac.Sum(nil))
}
