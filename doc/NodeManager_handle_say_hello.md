# NodeManager.handleSayHello
>处理请求公共逻辑
>1. 校验签名
>2. 判断自己的nStatus是否为MANAGER_STATUS_STARTED

>响应公共逻辑
>1. 签名

>下面的消息包括请求，响应，广播
>
>每次OnRead前会创建一个parent是TODO的ValueContext，并初始一个key是REQUEST_CONTEXT的RequestContext（network/common.go:(ServerConnection) readLoop）
>
>RequestContext是一个普通的struct，里面放有地址，连接，requestID（requestID初始为空，将消息解包后，才把请求中的requestID设置进去）
>
>后续的消息处理的函数都会带着这个ValueContext

从获取请求的context

再从请求context中获取服务连接

#### handleSayHello

对于消息的连接，考虑新旧节点以及新旧连接：

##### 新旧节点
1. 新节点：创建节点，添加到节点信息中表中

2. 旧节点：更新实例id

3. 新连接：添加到连接列表，？直接根据数量是否超过close

3. 再创建一个新连接的context，这个context的parent为请求的context，并放入key为CONTEXT_KEY_NODE，值为节点

##### 同步数据（TODO）
尝试同步全量数据

##### 响应请求

##### 再进行一次keepalive
最多取出GOOSIP_SIZE个邻接点，广播一次keepalive
