# NodeManager.handleGetEntryReq

### 什么是Entry
>指uri，entryType(http/rpc)，isNew信息

>处理请求公共逻辑
>1. 校验签名
>2. 判断请求节点的nStatus是否为NODE_STATUS_SYNCING/MANAGER_STATUS_STARTED
>3. 判断是否需要同步（版本号），是否有改变的entry（Node.modifyEntry）
>>改变：对通知中的全部entry，如果isNew，添加进来（算是有改变），否则清理掉（也是改变）
>
>对于改变的entry，如果isNew，添加到url2node中；否则：如果是新的节点，使用新节点覆盖之前的节点？（迷惑代码：manager.go:580~589）