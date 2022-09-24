# Grpc Reflection Demo

> A grpc reflection demo refer to grpcui and grpcurl.
>

不依赖pb文件，通过grpc反射接口实现：

- 从反射接口拉取 proto 文件信息；
- 解析 proto 文件获取 grpc 服务信息（用来构造 grpc 服务请求数据），包括 service, rpc method, message 元数据；
- 使用 grpc 服务名 + json body 数据（不依赖struct），完成 grpc 服务请求。

原理：

通过 grpc reflection 接口获取 grpc meta 信息 `descSource`，基于 `descSource + method + req body` 通过 `grpcurl` 完成 grpc 接口调用。

## Grpc reflection 测试

> 测试脚本参考 `./run.sh`。
>

1. 启动 grpc 服务

```sh
./svc_bin/grpc_hello
```

2. 调用 grpc 服务

打印 grpc 服务 meta 信息：

```sh
go run main.go -addr=127.0.0.1:50051 -debug
```

使用 json body 作为参数调用 grpc 服务：

```sh
go run main.go -addr=127.0.0.1:50051 -method=proto.Greeter.SayHello -body='{"metadata":[],"data":[{"name":"grpc"}]}'
# OnReceiveHeaders:
# content-type=[application/grpc]
# OnReceiveResponse: message:"Hello grpc"
# OnReceiveTrailers: OK 
```

## Grpc Mock 服务

### Mock 服务

部署一个通用的 grpc mock 服务。

问题：

- client 需要指定 服务名、请求及响应结构体 来调用 grpc 服务，如果访问 mock 服务，需要修改 `pb.go` 文件代码；
- mock 服务对 请求及响应结构体 的泛化。

### 模板文件

使用 template 模板文件 + .pb 文件 + json数据 动态生成 `mock_server.go` 文件，部署后代替原 grpc 服务。

问题：

- 只能 mock 一个服务，不能通用；
- 每次使用前需要部署。

