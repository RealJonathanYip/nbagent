package nbserver

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.yayafish.com/nbagent/cli_agent"
	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/node"
)

type NBServer struct {
	ptrNodeManager   *node.NodeManager
	ptrClientManager *cli_agent.ClientManager
}

func NewNBServer(objServerConf config.Config) *NBServer {

	var aryNeighbour []node.Neighbour
	for _, objNeighbour := range objServerConf.NodeConf.Neighbours {
		aryNeighbour = append(aryNeighbour,
			node.GetNeighbour(objNeighbour.Name, objNeighbour.IP, uint16(objNeighbour.Port), uint16(objNeighbour.AgentPort)))
	}
	ptrNodeManager := node.NewManager(objServerConf.NodeConf.Name,
		config.ServerConf.Host,
		uint16(config.ServerConf.NodeConf.Port),
		uint16(config.ServerConf.AgentConf.Port),
		aryNeighbour)

	ptrClientManager := cli_agent.NewClientManager(objServerConf.AgentConf.Name,
		config.ServerConf.Host,
		uint16(config.ServerConf.AgentConf.Port))

	ptrNodeManager.RegisterAgentHandler(ptrClientManager)
	ptrClientManager.RegisterNodeHandler(ptrNodeManager)

	var objNBServer NBServer = NBServer{
		ptrNodeManager:   ptrNodeManager,
		ptrClientManager: ptrClientManager,
	}
	return &objNBServer
}

func (ptrNBServer *NBServer) Init() {}

func (ptrNBServer *NBServer) Run(stopCh chan bool) {

	var anySignal = make(chan os.Signal)
	signal.Notify(anySignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-stopCh:
			{
				ptrNBServer.ptrNodeManager.Stop()
				ptrNBServer.ptrClientManager.Stop()
				goto WaitStop
			}
		case <-anySignal:
			{
				ptrNBServer.ptrNodeManager.Stop()
				ptrNBServer.ptrClientManager.Stop()
				goto WaitStop
			}
		}
	}

WaitStop:
	time.Sleep(time.Second * time.Duration(config.ServerConf.StopTimeout))
}
