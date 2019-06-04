[English](./Readme.md) | **中文**
<p align="center">
<img align="left" width = "30%" src="images/logo.png">
<br>
</p>
<br>
<br>
<br>

### 功能介绍
- proxy代理
- 日志记录
- 一级节点的ip白名单(支持继承)
- 限速

### 架构图
<center>
<img src="images/arch.png" width = "60%" />
</center>

### 压测

```
BenchmarkProxyGet-48               10000            144907 ns/op
BenchmarkZKGet-48                  10000            102587 ns/op
BenchmarkProxyConnGet-48            1000           2211095 ns/op
BenchmarkZKConnGet-48               1000           1825001 ns/op
BenchmarkProxyCreateSet-48          2000            799873 ns/op
BenchmarkZKCreateSet-48             2000            658289 ns/op
```

### 部署
- 修改zookeeper的maxClientCnxns配置

```
maxClientCnxns=1000000
```
- 编译proxy

```
cd zk-proxy
sh build.sh
```
- 监控proxy

```
echo isok|nc 127.1 2182
echo info|nc 127.1 2182
echo sess|nc 127.1 2182
curl http://127.1:8000/debug/vars
```

### 感谢
- [https://github.com/samuel/go-zookeeper](https://github.com/samuel/go-zookeeper)
- [https://github.com/etcd-io/zetcd](https://github.com/etcd-io/zetcd)