# NodeManager

### 结构
```go
type NodeManager struct {
	mTempConnection       map[*network.ServerConnection]int64
	objTempConnectionLock sync.Mutex
	ptrNodeRouter         *router.Router
	mNodeInfo             map[string]*Info
	objLock4NodeInfoMap   sync.RWMutex
	mRpcFunc2Node         map[string][]*Info
	objRpcFunc2NodeMap    sync.RWMutex
	ptrSelf               *Info
	nSequence             uint64
	aryNeighbour          []Neighbour
	ctx                   context.Context
	nStatus               int32 // 初始状态：MANAGER_STATUS_STARTED
	ptrServer             *network.Server
}
```

### 各变量初始
初始化函数
>NewManager(szName, szIP string, nPort uint16, aryNeighbour []Neighbour) *NodeManager

成员初始值
```go
mTempConnection: make(map[*network.ServerConnection]int64)
objTempConnectionLock: sync.Mutex{}
ptrRouter := router.NewRouter()
ptrNodeRouter: ptrRouter
mNodeInfo: make(map[string]*Info)
objLock4NodeInfoMap: sync.RWMutex{}
mRpcFunc2Node: make(map[string][]*Info)
objRpcFunc2NodeMap: sync.RWMutex{}
ptrSelf: ptrSelf = newNode("node_" + uuid.New(), szIP, nInstance, nPort, ptrManager, ptrManager.ctx)
nSequence: 0
aryNeighbour: aryNeighbour(参数)
ctx: context.WithValue(context.TODO(), NODE_INFO_KEY,
     			&node.NodeInfo{Name: szName, IP: szIP, Port: uint32(nPort), InstanceID: nInstance})
nStatus: MANAGER_STATUS_STARTED
ptrServer: nil
```

下面这个指针暂时不知道后面会不会用
```go
mNodeInfo[ptrSelf.szName] = ptrManager.ptrSelf
```

### 注册路由
```go
ptrRouter.RegisterRouter(rpc.NODE_SAY_HELLO, ...)
ptrRouter.RegisterRouter(rpc.NODE_SAY_GOOBEY, ...)
ptrRouter.RegisterRouter(rpc.NODE_GET_ENTRY, ...)
ptrRouter.RegisterRouter(rpc.NODE_ENTRY_NOTIFY, ...)
ptrRouter.RegisterRouter(rpc.NODE_KEEP_ALIVE, ...)
ptrRouter.RegisterRouter(rpc.NODE_RPC_REQ, ...)
ptrRouter.RegisterRouter(rpc.NODE_RPC_RESP, ...)
```

### 服务启动
```go
ptrServer := network.NewServer(szIP, nPort, false, ptrManager, ptrManager)
ptrServer = ptrServer
go ptrServer.Start()
```

### 定时器
>其中，cleanTimeout不会发送协议，其它两个会发送协议

#### 启动3个定时器
- [cleanTimeout](#cleanTimeout)
- [checkNeighbour](#checkNeighbour)
- [keepalive](#keepalive)

###### cleanTimeout
清理超时连接
1. (this *NodeManager) deleteNode(ptrNode *Info)
>先根据uri清理节点
>>(this *NodeManager) deleteEntry(ptrNode *Info, aryEntry []*node.EntryInfo)

>再根据name清理节点
>>delete(this.mNodeInfo, ptrNode.szName)


###### checkNeighbour
对初始化时传入的neighbour（可能后续也有新增）
新增不存在的节点信息，同时启动一个协程进行连接
```go
ptrNode := newNode(objNeighbour.szName, objNeighbour.szIP, 0, objNeighbour.nPort, this, this.ctx)
this.mNodeInfo[objNeighbour.szName] = ptrNode
...
ptrNode.reconnect()
```

###### keepalive
```go
随机选5个（GOOSIP_SIZE）节点，对非自己的节点广播keepalive
```
