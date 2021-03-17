# demo.hello

> golang grpc, grpc gateway and grpc web ui demos.

------

## grpc

> Refer: <https://grpc.io/docs/languages/go/quickstart/>

安装grpc包及工具：

```sh
go get google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
```

基于proto文件，生`pb.go`代码：

```sh
protoc --help

cd ..
protoc -I . --go_out=plugins=grpc:. demo.grpc/grpc/proto/message/*.proto
protoc -I . --go_out=plugins=grpc:. demo.grpc/grpc/proto/helloworld.proto
# -IPATH Specify the directory in which to search for imports.
# If not given, the current working directory is used.
```

启动grpc服务端和客户端：

```sh
go run greeter_server/main.go
go run greeter_client/main.go [string]
```

------

## grpc-gateway

安装 grpc-gateway 工具：

```sh
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
```

生成 grpc 和 gateway 服务端代码：

```sh
cd [proto_dir]
protoc -I . --go_out=plugins=grpc:. google/api/*.proto
protoc -I . --go_out=plugins=grpc:. ./service.proto
protoc --grpc-gateway_out=logtostderr=true:. ./service.proto
```

### 测试

启动 grpc 和 grpc gateway 服务。

```sh
# grpc server
go run server/grpc_server/main.go
# grpc http gateway
go run server/http_server/main.go
```

访问 http api 接口：

```sh
curl -XPOST -k http://localhost:8081/v1/example/echo -d '{"value": "world"}'
curl -XPOST -k http://localhost:8081/v1/example/hello \
  -H "X-User-Id: 100d9f38" -H "Grpc-Metadata-Uid: b3a108f81a30" -d '{"name": "Jim"}'
```

------

## swagger

安装：

```sh
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
```

生成 swagger api 规范文件：

```sh
protoc -I . --swagger_out=logtostderr=true:. ./service.proto
```

### swagger ui

start swapper ui server:

```sh
docker pull metz/swaggerui
docker run --name swaggerui -d --rm -p 8080:8080 metz/swaggerui
```

访问swagger ui页面：<http://localhost:8080/>

发布 swagger api json 文件到swagger服务端：

```sh
curl -v -XPOST "http://localhost:8080/publish" -H "Content-Type: application/json" -d @service.swagger.json
```

------

## grpc web ui

> Refer: <https://studygolang.com/articles/24551>
>
> Github: <https://github.com/fullstorydev/grpcui>

### grpcui 服务构建

下载并构建 grpcui 文件：

```sh
git clone https://github.com/fullstorydev/grpcui.git
cd [grpcui_project]/cmd/grpcui
go build
```

检查 grpcui 构建是否成功：

```sh
./grpcui -help
```

### grpcui 支持

1. 反射

grpc 服务添加反射支持：

```golang
reflection.Register(s)
```

构建 grpc 服务文件：

```sh
cd [grpc_project]/grpc_server
go build
```

否则启动 grpcui 服务报错：`Failed to compute set of methods to expose: server does not support the reflection API`

2. protoset

grpcui使用protoset文件代替反射，例子：

```sh
./grpcui -v -protoset USER_SERVICE.protoset -protoset ACCOUNT.protoset \n
  -protoset ACCOUNT_API.protoset -protoset ACCOUNT_ADMIN.protoset -bind 0.0.0.0 -port 30008
```

### grpcui 测试

启动 grpc 和 grpcui 服务：

```sh
# start grpc server
./grpc_server

# start grpcui
./grpcui -plaintext localhost:9090
gRPC Web UI available at http://127.0.0.1:57142/
```

访问 grpc ui 页面，测试 grpc 接口：<http://127.0.0.1:57142/>

访问 grpc ui api 接口来访问grpc服务：

```sh
# 获取 _grpcui_csrf_token
curl --head http://127.0.0.1:57688/invoke/demo.hello.Service1.Echo

# 直接访问 echo 接口
curl -v -XPOST http://127.0.0.1:57688/invoke/demo.hello.Service1.Echo \
  -H "Content-Type: application/json" \
  -H "x-grpcui-csrf-token: Iy6lXGmWWuyhsgG7eV-OWtruUiSf7Ijo1MoCjhW0fYM" \
  --cookie "_grpcui_csrf_token=Iy6lXGmWWuyhsgG7eV-OWtruUiSf7Ijo1MoCjhW0fYM" \
  -d '{"metadata":[],"data":[{"value":"world"}]}' | jq .
```

------

## pprof for grpc

启动 grpc 和 gateway 服务：

```sh
cd [grpc_project]/gateway
go run server/grpc_server/main.go
2020/09/14 11:42:23 grpc listen at: :9090
2020/09/14 11:42:23 start pprof server at: :8024

go run server/http_server/main.go
HTTP server (proxy to localhost:9090) start at :8081
```

访问 pprof 页面：<http://localhost:8024/debug/pprof/>

### 测试

设置执行时间，启动api接口性能测试脚本（wrk）：

```sh
./test/test.sh
```

获取各维度的性能数据：

```sh
pprof http://localhost:8024/debug/pprof/profile # 默认采集需要30秒
pprof http://localhost:8024/debug/pprof/profile\?seconds\=10

pprof http://localhost:8024/debug/pprof/heap
pprof http://localhost:8024/debug/pprof/block
pprof http://localhost:8024/debug/pprof/mutex
```

pprof 交互：

```sh
Saved profile in /Users/jinzheng/pprof/pprof.samples.cpu.001.pb.gz
(pprof)> top 10
(pprof)> exit
```

通过web访问性能分析结果：（地址 <http://localhost:8025/ui>）

```sh
pprof -http=localhost:8025 pprof.samples.cpu.001.pb.gz
pprof -http=localhost:8025 pprof.alloc_objects.alloc_space.inuse_objects.inuse_space.001.pb.gz
```

------

## goc coverage

1. start goc coverage register center.

```sh
goc server
```

2. build server bin with fixed goc agent port.

```sh
cd [grpc_src_dir]
goc build . --agentport :8026
```

3. start grpcui and run api tests.

```sh
./grpcui -plaintext localhost:9090
```

4. create coverage profile file for server under test.

```sh
# get available goc agents
goc list
goc profile --address "http://127.0.0.1:8026" -o profile.cov
```

5. generate coverage html report.

```sh
cd [grpc_src_dir]
cp profile.cov .

# if multiple .cov files, do merge
goc merge a.cov b.cov -o profile.cov
go tool cover -html=profile.cov
```

------

## k8s deploy: grpc + grpcui + goc

1. build gppc, grpcui, goc images, and push to k8s (minikube) repo

```sh
cd demo.grpc; ./run.sh
```

2. deploy gppc, grpcui, goc services in k8s

```sh
cd demo.grpc/grpc/deploy
kubectl apply -f deploy.yaml

# check deploy components
kubectl get all
kubectl get ing
```

3. check grpcui web by "http://grcui.default.k8s/", and test grpc api

4. verify goc center available

```sh
goc list --center=http://goccenter.default.k8s
# output: {"greeter_server":["http://127.0.0.1:50052"]}
```

5. create go coverage report

```sh
cd demo.grpc/grpc/greeter_server
# create coverage data file
goc profile --center=http://goccenter.default.k8s --address=http://127.0.0.1:50052 -o coverprofile.cov
# check coverage data file
go tool cover -func coverprofile.cov
# go tool cover -func coverprofile.cov -o cover_summary.txt
# create coverage html
go tool cover -html=coverprofile.cov -o coverage.html

# clear coverage data on server side
goc clear --center=http://goccenter.default.k8s --address=http://127.0.0.1:50052
```

------

## 其他

- mysql api

测试：

```sh
curl -v http://localhost:9091/ping

curl -v -XPOST http://localhost:9091/db/query -d '{"query":"select * from users where id < 4"}'
```

