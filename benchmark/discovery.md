---
代理端测试方案
---
## URI写入性能
### 测试流程
> client通过agent.register注册URI到代理。
> client等到注册回复，判断注册是否正常。
### 测试中的变量
- 并发数（客户端数量）
- 每个客户端的URI个数
- 每个客户端的连接数
- URI的长度（字节）
- 代理的个数（1/3）
### 测试统计数据
- 注册URI总数
- 每个客户端注册请求的时延，全部客户端总的时延
- 注册请求的QPS
- 计算时延最大值，最小值，平均值
### 参考
- etcd key 写入性能
- https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/performance.md



## URI读取性能
### 测试流程
> 增加URI路由探测协议
> client 给代理发送路由探测协议。
> 代理检查路由，将结果返回给 client。
### 测试中的变量
- 并发数（客户端数量）
- 每个客户端的URI个数
- 每个客户端的连接数
- URI的长度（字节）
- 代理的个数（1/3）
### 测试统计数据
- 注册URI总数
- 每个客户端路由探测请求的时延，全部客户端总的时延
- 探测请求的QPS
- 计算时延最大值，最小值，平均值
### 参考
- etcd key 读取性能
- https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/performance.md


## 结果

- 3 docker of 8 vCPUs + 16GB Memory
- 1 machine(client) of 16 vCPUs + 16GB Memory
- Ubuntu 16.04

With this configuration, nbagent can approximately wtite:

| Number of URIs | URI size in uris | URI space in uris | Number of clients | Target agent | Average write QPS | Average latency per request | Average server RSS |
|---------------:|------------------:|--------------------:|----------------------:|------------------:|--------------------|------------------:|------------------:|
| 10,000 | 256 | 1   | 1   | one only | 3748 | 0.27ms | 22,816 KB |
| 100,000| 256 | 1000   | 1000 | one only | 13,675 | 0.07 ms |  50,904 KB |
| 100,000| 256 | 1000   | 1000 | all node | 40,751  | 0.02 ms |  27,680 KB |

Sample commands are:

```sh
# write to one
 
discovery --agents=HOST_1:Port --call-timeout=10 --clients=1 --only-agent=true \
register --secret-key=node --total-register=10000 --uri-size=256
discovery --agents=HOST_1:Port --call-timeout=10 --clients=1000 --only-agent=true \
register --secret-key=node --total-register=100000 --uri-size=256

# write to all nodes
    
discovery --agents=HOST_1:Port,HOST_2:Port,HOST_3:Port --call-timeout=10 --clients=1000 --only-agent=false \
register --secret-key=node --total-register=100000 --uri-size=256
```


nbagent can read:

| Number of URIs | URI size in uris | Number of clients | Average read QPS | Average latency per request |
|---------------:|------------------:|--------------------:|------------------:|--------------------|
| 10,000 | 256 | 1    | 4340   | 0.23 ms |
| 100,000| 256 | 1000 | 64,222 | 0.02 ms |

```sh
discovery --agents=HOST_1:Port,HOST_2:Port,HOST_3:Port --call-timeout=10 --clients=1 --only-agent=false \
read --secret-key=node --total-register=10000 --uri-size=256

discovery --agents=HOST_1:Port,HOST_2:Port,HOST_3:Port --call-timeout=10 --clients=1000 --only-agent=false \
read --secret-key=node --total-register=100000 --uri-size=256
```