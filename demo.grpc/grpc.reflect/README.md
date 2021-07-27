# Grpc Reflect Demo

> A grpc reflect demo refer to grpcui and grpcurl.

功能：

1. 获取 proto 文件信息。
2. 获取 grpc 服务信息（用来构造 grpc 服务请求数据），包括 service, rpc method, message 元数据。
3. 通过 grpc 服务名 + 构造的json数据（不依赖pb文件），请求 grpc 服务。

## 测试

1. 启动 grpc 服务

```sh
./bin/grpc_hello
```

2. 调用 grpc 服务

```sh
go build .

# 打印 grpc 服务 meta 信息
./grpc.reflect -addr=127.0.0.1:50051 -debug
# 调用 grpc 服务
./grpc.reflect -addr=127.0.0.1:50051 -method=proto.Greeter.SayHello -body='{"metadata":[],"data":[{"name":"grpc"}]}'
```

> 参考 `run.sh`。

## Grpc Mock 服务

### Mock 服务

部署一个通用的 grpc mock 服务。

问题：

- client需要指定 服务名、请求及响应结构体 来调用grpc服务，如果访问mock服务，需要修改`pb.go`文件代码；
- mock服务对 请求及响应结构体 的泛化。

### 模板文件

使用 template 模板文件 + .pb 文件 + json数据 动态生成 `mock_server.go` 文件，部署后代替原grpc服务。

问题：

- 只能mock一个服务，不能通用；
- 每次使用前需要部署。

