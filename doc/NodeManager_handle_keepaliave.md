# NodeManager.handleKeepAlive

>处理请求公共逻辑
>1. ~~校验签名~~
>2. 判断请求节点的nStatus是否为NODE_STATUS_DEAD

### 同步数据
1. 处理流言（introduceNeighbour）

将请求中带上的邻接节点，取出新节点，新增节点信息，建立连接

2. 尝试全量同步（trySyncAll）

根据版本以及自己的状态，判断是否需要从其它进程同步全部（rpc:GetEntryReq）