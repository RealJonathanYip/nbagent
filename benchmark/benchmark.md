# 客户端测试方案
## RPC 性能
### 测试流程
> client通过protobuf协议和server通信。
> 请求发送给server, server解码、更新两个字段、编码再回复给client，client通过字段判断是否正常
### 测试参数以及统计数据
- 并发数（客户端数量）
- 请求总数
- 平均分配请求到每个客户端
- 统计每个并发请求成功的数量，请求的总数，
- 记录每个客户端的每个收发时间，全部客户端总的收发时间
- 计算最大值，最小值，平均值

## 注册反注册性能
### 测试流程
> 单次注册反注册性能测试中，使用两个客户端，一个动态uri注册，一个使用动态的uri发起rpc
> 注册成功后，通过rpc成功判断是否注册成功
> 反注册成功后，通过rpc失败判断是否反注册是否成功
### 测试参数以及统计数据
- 并发数（double客户端数量）
- 测试总数
- 平均分配测试数量到每个并发中
- 统计数据基本同上



## 结果
- Machine 1
  - Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz, 32G RAM
  - Ubuntu 16.04 x86_64
  - 100Mb/s

- Machine 2
  - Intel(R) Core(TM) i7-7700 CPU @ 3.60GHz, 16G RAM
  - Window 10 64bit
  - 100Mb/s

## 单Agent
##### 
- Machine 1: nbagent & Server
- Machine 2: Client

| Number of request | Number of concurrency | Number of received requests OK | Total time | QPS | Average latency | Median latency | Max latency | Min latency | p99 |
|------------------:|----------------------:|-------------------------------:|-----------:|----:|----------------:|---------------:|------------:|------------:|----:|
| 1,000,000 | 10 | 999,981 | 316,198 ms | 3,162 | 3 ms | 2 ms | 29 ms | 0 ms | 15 ms |
| 100,000 | 10 | 100,000 | 31,776 ms | 3,147 | 3 ms | 2 ms | 22 ms | 0 ms | 14 ms |
| 500,000 | 10 | 500,000 | 156,313 ms | 3,198 | 3 ms | 2 ms | 21 ms | 0 ms | 14 ms |
| 500,000 | 50 | 500,000 | 26,968 ms | 18,540 | 2 ms | 1 ms | 132 ms | 0 ms | 17 ms |
| 500,000 | 100 | 500,000 | 28,409 ms | 17,600 | 5 ms | 3 ms | 137 ms | 0 ms | 41 ms |
| 500,000 | 200 | 497,399 | 36,129 ms | 13,839 | 14 ms | 9 ms | 3055 ms | 0 ms | 77 ms |



| Number of request | Number of concurrency | Number of received requests OK | Total time | QPS | Average latency | Median latency | Max latency | Min latency | p99 | CPU | Mem |
|------------------:|----------------------:|-------------------------------:|-----------:|----:|----------------:|---------------:|------------:|------------:|----:|----:|----:|
| 6,000,000 | 30 | 5,993,782 | 395,378 ms | 15,175 | 1 ms | 1 ms | 94 ms | 0 ms | 7 ms | 640% | 0.4% |
| 6,000,000 | 40 | 5,991,234 | 340,717 ms | 17,609 | 2 ms | 1 ms | 3001 ms | 0 ms | 10 ms | 650% | 0.4% |
| 6,000,000 | 50 | 5,995,758 | 310,472 ms | 19,325 | 2 ms | 1 ms | 102 ms | 0 ms | 12 ms | 530% | 0.4% |



#####
- Machine 1: nbagent
- Machine 2: Client & Server

| Number of request | Number of concurrency | Number of received requests | Number of received requests OK | Total time | QPS | Average latency | Median latency | Max latency | Min latency | p99 |
|------------------:|----------------------:|----------------------------:|-------------------------------:|-----------:|----:|----------------:|---------------:|------------:|------------:|----:|
| 500,000 | 50 | 500,000 | 500,000 | 53,565 ms | 9,334 | 5 ms | 4 ms | 39 ms | 1 ms | 26 ms |
| 500,000 | 100 | 500,000 | 500,000 | 52,076 ms | 9,601 | 10 ms | 8 ms | 125 ms | 1 ms | 56 ms |
| 500,000 | 200 | 500,000 | 496,732 | 55,334 ms | 9,036 | 21 ms | 17 ms | 3026 ms | 0 ms | 145 ms |



## 双Agent
- Machine 1: nbagent & Server
- Machine 2: nbagent & Client

| Number of request | Number of concurrency | Number of received requests OK | Total time | QPS | Average latency | Median latency | Max latency | Min latency | p99 | Client nbagent CPU | Client nbagent Mem | Server nbagent CPU | Server nbagent Mem |
|------------------:|----------------------:|-------------------------------:|-----------:|----:|----------------:|---------------:|------------:|------------:|----:|----------------:|----------------:|----------------:|----------------:|
| 500,000 | 10 | 499,479 | 100,459 ms | 4,977 | 2 ms | 1 ms | 201 ms | 0 ms | 6 ms | 28% | 40m | 150% | 0.3% |
| 500,000 | 50 | 499,724 | 48,973 ms | 10,209 | 4 ms | 3 ms | 1094 ms | 0 ms | 30 ms | 45% | 140m | 160% | 0.3% |
| 500,000 | 100 | 499,354 | 48,629 ms | 10,281 | 9 ms | 6 ms | 1220 ms | 0 ms | 66 ms | 60% | 240m | 240% | 0.3% |
| 500,000 | 200 | 498,946 | 66,780 ms | 7,487 | 26 ms | 19 ms | 1305 ms | 0 ms | 151 ms | 40% | 240m | 200% | 0.3% |




##### 初步测试
nbagent
- Intel(R) Xeon(R) Platinum 8163 CPU @ 2.50GHz, 32G RAM
- Ubuntu 16.04 x86_64
- 10000Mbps

Server/Client
- Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz, 32G RAM
- Ubuntu 16.04 x86_64
- 100Mb/s, upstream bandwidth: 2Mbps, downstream bandwidth: 20Mbps

- message size: 237 bytes

| Number of request | Number of concurrency | Number of received requests | Number of received requests OK | Total time | QPS | Average latency | Median latency | Max latency | Min latency | p99 |
|------------------:|----------------------:|----------------------------:|-------------------------------:|-----------:|----:|----------------:|---------------:|------------:|------------:|----:|
| 10,000 | 10 | 10,000 | 10,000 | 29,384 ms | 340 | 27 ms | 21 ms | 1530 ms | 19 ms | 666 ms |
| 10,000 | 50 | 10,000 | 9,998 | 35,783 ms | 279 | 169 ms | 234 ms | 3000 ms | 19 ms | 1,293 ms |
| 10,000 | 100 | 10,000 | 9,385 | 41,631 ms | 240 | 376 ms | 236 ms | 3000 ms | 0 ms | 3,000 ms |
| 10,000 | 200 | 10,000 | 6,446 | 37,297 ms | 268 | 574 ms | 234 ms | 8078 ms | 0 ms | 5,561 ms |

Sample config are:

###### nbagent
```ini
<server>
    <log_mode>asyncfile</log_mode>
    <log_level>info</log_level>
    <log_path>./nbagentlog/</log_path>
    <secret_key>node</secret_key>
    <workers>
        <worker name="node_introduce" size="100" queue="100"/>
        <worker name="node_manager" size="100" queue="100"/>
        <worker name="tcp_msg" size="100" queue="100"/>
        <worker name="node_connect" size="100" queue="100"/>
        <worker name="network_task_flag" size="100" queue="100"/>
    </workers>
    <host>0.0.0.0</host>
    <node>
        <name>node_1</name>
        <port>8900</port>
        <neighbours>
            <!--            <info name="" ip="" port="" agent_port=""></info>-->
        </neighbours>
        <neighbour_max_connection>10</neighbour_max_connection>
    </node>
    <agent>
        <name>node_1</name>
        <port>8800</port>
        <client_max_connection>300010</client_max_connection>
    </agent>

    <timeout_check_interval>10</timeout_check_interval> <!-- seconds -->
    <connect_timeout>60</connect_timeout> <!-- seconds -->
    <sign_timeout>5</sign_timeout> <!-- seconds -->
    <client_timeout>15</client_timeout> <!-- seconds -->
    <stop_timeout>60</stop_timeout> <!-- seconds -->
</server>
```


###### server
```ini
<server>
    <net_workers>
        <worker name="network_task_flag" size="100" />
    </net_workers>
    <host>{nbagent_IP}</host>
    <port>8810</port>
    <delay>0</delay>
    <init_conn_num>1000</init_conn_num>
    <max_conn_num>1100</max_conn_num>
    <good_bye_timeout_s>60</good_bye_timeout_s>
    <secret_key>node</secret_key>
</server>
```


####### client
```ini
<server>
    <net_workers>
        <worker name="network_task_flag" size="100" />
    </net_workers>
    <total>10000</total>
    <concurrency>200</concurrency>
    <max_total>100000</max_total>
    <max_concurrency>10000</max_concurrency>
    <delta_total>500</delta_total>
    <delta_concurrency>100</delta_concurrency>
    <init_conn_num>50</init_conn_num>
    <max_conn_num>100</max_conn_num>
    <good_bye_timeout_s>60</good_bye_timeout_s>
    <agents>
        <agent ip="{nbagent_IP}" port="8810" />
    </agents>
    <secret_key>node</secret_key>

    <log>
        <target>asyncfile</target>
        <encode>json</encode>
        <path>./log/nbagent-client-benchmark-test</path>
        <level>debug</level>
    </log>
</server>
```
