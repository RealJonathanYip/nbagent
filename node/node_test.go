package node

import (
	"git.yayafish.com/nbagent/protocol/node"
	"testing"
	"time"
)

var (
	ptrNode1 = NewManager("node_1", "127.0.0.1", 8800, 0,nil)
	ptrNode2 = NewManager("node_2", "127.0.0.1", 8801, 0,[]Neighbour{Neighbour{"node_1", "127.0.0.1", 8800,18800}})
   ptrNode3 = NewManager("node_3", "127.0.0.1", 8802, 0,[]Neighbour{Neighbour{"node_1", "127.0.0.1", 8800,18800}})
   ptrNode4 = NewManager("node_4", "127.0.0.1", 8803, 0,[]Neighbour{Neighbour{"node_1", "127.0.0.1", 8800,18800}})
	ptrNode5 = NewManager("node_5", "127.0.0.1", 8804, 0,[]Neighbour{Neighbour{"node_1", "127.0.0.1", 8800,18800}})
	ptrNode6 = NewManager("node_6", "127.0.0.1", 8805, 0,[]Neighbour{Neighbour{"node_1", "127.0.0.1", 8800,18800}})
	ptrNode7 = NewManager("node_7", "127.0.0.1", 8806, 0,[]Neighbour{Neighbour{"node_1", "127.0.0.1", 8800,18800}})
)
func init() {

}

func TestAdd(ptrTest *testing.T) {
	ptrNode1.AddEntry("test_add_entry", node.EntryType_HTTP)
	objEntry := node.EntryInfo{URI: "test_add_entry", EntryType: node.EntryType_HTTP, IsNew: true}

	if !ptrNode1.ptrSelf.entryExist(&objEntry) {
		ptrTest.Errorf("add entry fail!")
		return
	}

	objEntry2 := node.EntryInfo{URI: "test_add_entry", EntryType: node.EntryType_RPC, IsNew: true}
	if ptrNode1.ptrSelf.entryExist(&objEntry2) {
		ptrTest.Errorf("entry exist error!")
		return
	}

	ptrNode1.AddEntry("test_add_entry", node.EntryType_RPC)

	if !ptrNode1.ptrSelf.entryExist(&objEntry) {
		ptrTest.Errorf("add entry fail!")
		return
	}

	ptrNode1.AddEntry("test_add_entry", node.EntryType_RPC)
}

func TestRemove(ptrTest *testing.T) {
	ptrNode1.AddEntry("test_add_entry", node.EntryType_HTTP)
	objEntry := node.EntryInfo{URI: "test_add_entry", EntryType: node.EntryType_HTTP, IsNew: true}

	if !ptrNode1.ptrSelf.entryExist(&objEntry) {
		ptrTest.Errorf("add entry fail!")
		return
	}

	ptrNode1.RemoveEntry("test_add_entry", node.EntryType_HTTP)

	if ptrNode1.ptrSelf.entryExist(&objEntry) {
		ptrTest.Fatalf("delete entry fail!")
		return
	}

}

func TestNeighbour(ptrTest *testing.T) {
	time.Sleep(time.Second * 3)

	if _, bExist := ptrNode1.mNodeInfo[ptrNode2.ptrSelf.szName]; !bExist {
		ptrTest.Fatalf("find neighbour fail!")
	}

}

func TestNeighbourFind(ptrTest *testing.T) {
	if _, bExist := ptrNode3.mNodeInfo[ptrNode4.ptrSelf.szName]; !bExist {
		ptrTest.Fatalf("find neighbour's neighbour fail!")
	}
}

func TestAddAndRemove(ptrTest *testing.T) {
	time.Sleep(2 * time.Second)
	ptrNode5.AddEntry("test_add_and_remove", node.EntryType_HTTP)

	time.Sleep(2 * time.Second)

	ptrNode6.objRpcFunc2NodeMap.RLock()
	if len(ptrNode6.mRpcFunc2Node["test_add_and_remove_" + node.EntryType_HTTP.String()]) == 0 {
		ptrTest.Fatalf("check add fail!:%+v %+v", ptrNode6.ptrSelf.mEntryInfo, ptrNode6.mNodeInfo[ptrNode6.ptrSelf.szName].mEntryInfo)
	}
	ptrNode6.objRpcFunc2NodeMap.RUnlock()

	ptrNode5.RemoveEntry("test_add_and_remove", node.EntryType_HTTP)

	time.Sleep(3 * time.Second)

	ptrNode6.objRpcFunc2NodeMap.RLock()
	if len(ptrNode6.mRpcFunc2Node["test_add_and_remove_" + node.EntryType_HTTP.String()]) > 0 {
		ptrTest.Fatalf("check remove fail!")
	}

	if len(ptrNode6.mRpcFunc2Node["test_add_entry_" + node.EntryType_RPC.String()]) == 0 {
		ptrTest.Fatalf("check add fail!")
	}
	ptrNode6.objRpcFunc2NodeMap.RUnlock()
}

func TestStop(ptrTest *testing.T) {
	time.Sleep(2 * time.Second)

	ptrNode1.Stop()

	time.Sleep(2 * time.Second)

	ptrNode7.objRpcFunc2NodeMap.RLock()

	if len(ptrNode7.mRpcFunc2Node["test_add_entry_" + node.EntryType_RPC.String()]) > 0 {
		ptrTest.Fatalf("check stop fail!")
	}

	ptrNode7.objRpcFunc2NodeMap.RUnlock()

	time.Sleep(2 * time.Second)
	ptrNode1 = NewManager("node_1", "127.0.0.1", 8800, 18800, []Neighbour{Neighbour{"node_2", "127.0.0.1", 8801, 18801}})

	time.Sleep(10 * time.Second)
	ptrNode7.objRpcFunc2NodeMap.RLock()

	if len(ptrNode7.mRpcFunc2Node["test_add_entry_" + node.EntryType_RPC.String()]) > 0 {
		ptrTest.Fatalf("check stop fail!")
	}

	ptrNode7.objRpcFunc2NodeMap.RUnlock()
}