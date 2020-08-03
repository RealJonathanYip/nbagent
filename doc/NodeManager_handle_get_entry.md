# NodeManager.handleGetEntryReq

### 什么是Entry
>指uri，entryType(http/rpc)，isNew信息

>处理请求公共逻辑
>1. 校验签名
>2. 判断自己的nStatus是否为MANAGER_STATUS_STARTED

取出全部entry，返回