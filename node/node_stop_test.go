package node

import (
	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/network"
	"git.yayafish.com/nbagent/taskworker"
	"testing"
	"time"
)

func init() {
	config.ServerConf.WorkerConfigs.Workers = append(
		config.ServerConf.WorkerConfigs.Workers, taskworker.WorkerConfig{"node_introduce", 100})
	config.ServerConf.WorkerConfigs.Workers = append(
		config.ServerConf.WorkerConfigs.Workers, taskworker.WorkerConfig{"node_manager", 100})
	config.ServerConf.WorkerConfigs.Workers = append(
		config.ServerConf.WorkerConfigs.Workers, taskworker.WorkerConfig{"tcp_msg", 100})
	config.ServerConf.WorkerConfigs.Workers = append(
		config.ServerConf.WorkerConfigs.Workers, taskworker.WorkerConfig{"node_connect", 100})
	config.ServerConf.WorkerConfigs.Workers = append(
		config.ServerConf.WorkerConfigs.Workers, taskworker.WorkerConfig{"network_task_flag", 100})

	taskworker.TaskWorkerManagerInstance().Init(config.ServerConf.WorkerConfigs.Workers)
}

func TestSafeStop(ptrTest *testing.T) {
	ptrNode1Temp := NewManager("node_1", "127.0.0.1", 8808, 18808,nil)
	ptrNode2Temp := NewManager("node_2", "127.0.0.1", 8809, 18809,
		[]Neighbour{Neighbour{"node_1", "127.0.0.1", 8808,18808}})

	time.Sleep(10 * time.Second)

	ptrNode1Temp.objLock4NodeInfoMap.RLock()
	if _, bExist := ptrNode1Temp.mNodeInfo[ptrNode2Temp.ptrSelf.szName]; !bExist {
		ptrTest.Fatalf("not found node!")
	}
	ptrNode1Temp.objLock4NodeInfoMap.RUnlock()

	ptrNode2Temp.Stop()

	time.Sleep(2 * time.Second)
	ptrNode1Temp.objLock4NodeInfoMap.RLock()
	if ptrInfo, bExist := ptrNode1Temp.mNodeInfo[ptrNode2Temp.ptrSelf.szName]; !bExist {
		ptrTest.Fatalf("not found node!")
	} else if ptrInfo.alive() {
		ptrTest.Fatalf("node still alive!")
	}
	ptrNode1Temp.objLock4NodeInfoMap.RUnlock()

	ptrNode2Temp.ptrServer.Stop()


	time.Sleep(5 * time.Second)

	ptrNode2Temp.objLock4NodeInfoMap.RLock()
	ptrNode1 := ptrNode2Temp.mNodeInfo[ptrNode1Temp.ptrSelf.szName]
	ptrNode2Temp.objLock4NodeInfoMap.RUnlock()

	var aryConnectionList []*network.ServerConnection
	ptrNode1.objLock4Connection.RLock()
	for _, ptrServerConnectionTmp := range ptrNode1.aryConnectionList {
		aryConnectionList = append(aryConnectionList, ptrServerConnectionTmp)
	}
	ptrNode1.objLock4Connection.RUnlock()

	for _, ptrConnection := range aryConnectionList {
		ptrConnection.Close()
	}

	time.Sleep(2 * time.Second)

	ptrNode1Temp.objLock4NodeInfoMap.RLock()
	if _, bExist := ptrNode1Temp.mNodeInfo[ptrNode2Temp.ptrSelf.szName]; bExist {
		ptrTest.Fatalf(" node! still here")
	}
	ptrNode1Temp.objLock4NodeInfoMap.RUnlock()
}
