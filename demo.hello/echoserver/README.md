# goc覆盖率工具介绍

目的：获得功能测试后的代码覆盖率，检查哪些模块的覆盖率较低，或者当前feature所提交的代码哪些没有被覆盖。

> 单元测试使用go原生的代码覆盖率工具即可。

原理：使用开源工具goc来获取代码覆盖率。Github: <https://github.com/qiniu/goc>

优点：

1. 对项目源代码无侵入性，仅需要修改部署的pipeline：
  - 使用`goc build`替代`go build`来编译服务，插入桩代码，并开放一个agent服务端口用于上报覆盖率数据。
  - 部署goccenter服务用来获取、清除覆盖率数据等交互的服务。
2. 同时支持多个服务的覆盖率收集。
3. 被测服务部署完成后执行功能测试或回归测试，可动态获取覆盖率数据，并通过html方式展示覆盖率结果。

goccenter服务已在测试环境中部署：

- k8s内部地址：<http://goccenter.ns.ing:37777>
- 外部ingress地址：<http://goccenter.ns.ing>

注：覆盖率数据保存在被测服务的内存中，如果测试过程中被测服务重启，则覆盖率数据会丢失。因此如果要执行长时间的测试，建议定时获取覆盖率数据并保存。

## Run code coverage

Run and get code coverage results.

```sh
# Step1: compile service by goc, and deploy with jenkins pipeline
goc build --center=http://goccenter.ns.ing:37777 --agentport :7778 -o /app/echoserver

# Step2: after service deployed and started, check agent is register in goccenter
goc list --center=http://goccenter.ns.ing
# output: {"echoserver":["http://127.0.0.1:7778"]}

# Step3: run tests, and get code coverage data from goccenter
goc profile --center=http://goccenter.ns.ing --address=http://127.0.0.1:7778 -o coverprofile.cov

# Step4: copy coverage file to project root dir, and generate coverage report
cp coverprofile.cov echoserver/
# show code coverage summary for function
go tool cover -func=coverprofile.cov -o cover_summary.txt
# get code coverage html
go tool cover -html=coverprofile.cov -o coverage.html

# More:
# if there are multiple coverage files, do merge
goc merge coverprofile* -o coverprofile.cov

# clear agent code coverage data
goc clear --center=http://goccenter.ns.ing --address=http://127.0.0.1:7778
# clear all register agents info
goc init --center=http://goccenter.ns.ing
```

## Run coverage for specified files

Only output coverage data of the files matching the patterns.

```sh
# include ping.go and users.go
goc profile --coverfile="handlers/ping.go","users*" \
  --center=http://goccenter.ns.ing --address=http://127.0.0.1:7778 -o coverprofile.cov

# include go files in "handlers" dir
goc profile --coverfile="handlers/.*" \
  --center=goccenter.ns.ing --address=http://127.0.0.1:7778 -o coverprofile.cov
```

## go coverage report

覆盖率结果支持分支覆盖，代码染色结果支持路径覆盖。

1. `-func` 函数级别
2. `-html` 行级别

## Create cobertura results

Generate cobertura coverage results which is compatible with jenkins.

```sh
# create cobertura xml
gocover-cobertura < coverprofile.cov > cobertura.xml
# create cobertura html report
./run.sh
```

Html测试报告支持分支覆盖。

## 例子：多个pod副本的覆盖率收集

如下，Echoserver服务对应有3个pod副本：

```sh
kubectl get svc
NAME                 TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
service/echoserver   ClusterIP   100.127.232.10   <none>        38081/TCP   2m27s

kubectl get pod
NAME                          READY   STATUS    RESTARTS   AGE    IP               NODE     NOMINATED NODE   READINESS GATES
echoserver-7888666db7-6gtck   1/1     Running   0          106s   100.98.163.50    node14   <none>           <none>
echoserver-7888666db7-jbr5r   1/1     Running   0          2m5s   100.98.47.167    node23   <none>           <none>
echoserver-7888666db7-szcvr   1/1     Running   0          2m5s   100.98.226.161   node22   <none>           <none>
```

执行覆盖率收集：

```sh
# Step1. 被测服务部署完成后，确认三个pod副本已注册到goccenter (注册信息默认为pod ip+port)
goc list --center=http://goccenter.ns.ing
{"echoserver":["http://100.98.47.167:7778","http://100.98.226.161:7778","http://100.98.163.50:7778"]}

# Step2. 执行api测试
curl "http://echoserver.ns.ing/"
curl "http://echoserver.ns.ing/ping"
curl "http://echoserver.ns.ing/users/"

# Step3.1 使用"--address", 获取服务的每个pod副本对应的覆盖率数据
goc profile --center=http://goccenter.ns.ing --address=http://100.98.163.50:7778 -o coverprofile1.cov
goc profile --center=http://goccenter.ns.ing --address=http://100.98.47.167:7778 -o coverprofile2.cov
goc profile --center=http://goccenter.ns.ing --address=http://100.98.226.161:7778 -o coverprofile3.cov
# 合并上一步的覆盖率数据
goc merge coverprofile* -o coverprofile.cov

# Step3.2 或者使用"--service", 获取服务的覆盖率数据
goc profile --center=http://goccenter.ns.ing --service=echoserver -o coverprofile.cov

# Step4. 生成覆盖率测试报告
go tool cover -func=coverprofile.cov -o cover_summary.txt
go tool cover -html=coverprofile.cov -o coverage.html
```

## 被测试服务部署改造

1. `go build`的编译参数通过`goc build --buildflags`传入，如下：

```sh
goc build --buildflags="-mod=vendor -a -installsuffix cgo" --center=http://goccenter.ns.ing:37777 --agentport :7778 .
```

2. 如果项目代码库中依赖 vendor 目录，在使用`go tool cover`生成覆盖率报告前，需要设置环境变量`GOFLAGS="-mod=vendor"`。可通过`go env`查看当前`GOFLAGS`环境变量值。

## Echoserver APIs

```sh
curl "http://echoserver.ns.ing/"
curl "http://echoserver.ns.ing/ping"
curl "http://echoserver.ns.ing/users/"

curl "http://echoserver.ns.ing/test?base=3"
curl "http://echoserver.ns.ing/cover?cond1=true&cond2=false"
```

Test post api:

```sh
curl -XPOST "http://localhost:8081/users/new" -d '{"name": "tester01", "age": 39}'
```

## RateLimiter 测试

1. 安装  HTTP 负载测试工具 vegeta

```sh
brew install vegeta
```

2. 创建配置文件 `vegeta.conf`

- no limiter:

```text
GET http://localhost:8081/
```

- limiter:

```text
GET http://localhost:8081/ping
```

3. 以每个时间单元 10 个请求的速率执行 10 秒

```sh
vegeta attack -duration=10s -rate=10 -targets=vegeta.conf | vegeta report
```

测试结果：

```text
Requests      [total, rate, throughput]         100, 10.10, 1.41
Duration      [total, attack, wait]             9.9s, 9.899s, 733.31µs
Latencies     [min, mean, 50, 90, 95, 99, max]  302.664µs, 616.167µs, 566.177µs, 823.758µs, 928.856µs, 2.139ms, 3.14ms
Bytes In      [total, mean]                     1686, 16.86
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           14.00%
Status Codes  [code:count]                      200:14  429:86
Error Set:
429 Too Many Requests
```

