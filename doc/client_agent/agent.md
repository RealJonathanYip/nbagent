### 客户端代理模块概述

>1. 管理客户端连接
>2. 客户端注册，校验
>3. 管理调用路由，客户端请求转发,调用链路上下文（context）
>4. 心跳

### 应用客户端注册流程
1.客户端建立TCP连接，agent将连接将连接放到临时连接列表

2.客户端发起注册协议，agent 校验，若通过，则放到正式连接列表，
不通过则关闭连接

3.客户端定时发起心跳，agent 检查心跳


### 客户端请求处理流程
1.agent从连接处收到请求 RpcCallReq

2.判断本地服务是否能够处理请求：

（1）能够处理，将请求调用到对应服务

（2）不能处理，将请求投递给 node 模块

3.提供接口，接收来自node模块(或者其他模块)传过来的具体客户端路由请求

### 客户端请求处理链路
client_1 --> node --> agent --> clent_2

client_1 --> agent1 --> node1 --> node2 --> agent2 --> clent_2

### 协议
1. 注册协议
AgentRegisterReq

AgentRegisterRsp

2. 心跳协议
AgentKeepAliveNotify