# Grpc Application

## 介绍

1. grpc client 通过直接解析 proto 文件来完成 grpc 请求。
2. grpc server 提供 mocker 服务。

### Grpc Client

1. 动态加载并解析 proto 文件
2. protoc 构造 req proto message 后发送请求

> TODO: 通过 reflect 反射接口来加载 proto 定义，代替从 proto 文件加载。参考 grpcurl 工具。

原理：调用 grpc.ClientConn 的 Invoke 方法实现，参考 `application/client.go` 实现。

### Grpc Mocker Server

1. 根据 full method 匹配已定义的 json body
2. protoc 构造 resp proto message 并返回

原理：注册一个 unknown stream server handler 处理 mock 请求。参考 `application/server.go` 实现。

------

## Protoc

流程：

1. 加载和解析 proto 文件，生成 method descriptor
2. coder 通过 method desc 关联 req/resp message desc, 动态生成 proto message
3. 序列化 json body 到 proto message

### Proto 准备

1. proto 文件定义

```text
proto
├── account
│   ├── constant.proto  # go_package="pb/account;account"
│   └── deposit.proto   # go_package="pb/account;account"; import "constant.proto"
└── greeter
    ├── helloworld.proto  # go_package="pb/greeter;greeter"; import "msg/hello.proto"
    └── msg
        └── hello.proto   # go_package="pb/greeter;greeter"
```

参数定义：

- `go_package="path:pkg"` 生成 pb 文件的路径及 go package 定义。
- `import="path"` 从指定的路径中引入依赖的 pb 文件。

注：关联的 proto 文件的 `go_package` 一般是相同的，表示在同一个目录下生成的 pb 文件。

2. 执行 protoc 生成 pb 文件

```sh
cd protoc/
# account
local pb_dir="proto/account"
protoc --proto_path=${pb_dir} --go_out=. --go-grpc_out=. ${pb_dir}/*.proto

# greeter
local pb_dir="proto/greeter"
protoc --proto_path=${pb_dir} --go_out=. --go-grpc_out=. ${pb_dir}/helloworld.proto ${pb_dir}/msg/*.proto
```

3. protoc cli 解释

- `account/deposit.proto`
  - 导入 constant proto 文件 `--proto_path="proto/account"` + `import "constant.proto"`
  - 在 `pb/account` 目录下生成的 pb 文件 `--go_out=. --go-grpc_out=.` + `go_package="pb/account;account"`

- `greeter/helloworld.proto`
  - 导入 hello proto 文件 `--proto_path="proto/greeter"` + `import "msg/hello.proto"`
  - 在 `pb/greeter` 目录下生成的 pb 文件 `--go_out=. --go-grpc_out=.` + `go_package="pb/greeter;greeter"`

注：`--proto_path` 一般设置为 main proto 文件所在的目录。

### 解析 proto 文件

参考 `protoc/load_proto.go`, `protoc/coder.go` 实现。

