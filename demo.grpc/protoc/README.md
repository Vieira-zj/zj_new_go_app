# protoc

> 通过直接解析 proto 文件来完成 grpc 请求。

## 环境定义

1. proto 文件定义

```text
proto
├── account
│   ├── constant.proto  # go_package="pb/account;account"
│   └── deposit.proto   # go_package="pb/account;account"; import "constant.proto"
└── greeter
    ├── helloworld.proto     # go_package="pb/greeter;greeter"; import "msg/hello.proto"
    └── msg
        └── hello.proto  # go_package="pb/greeter;greeter"
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

3. 解释

- `account/deposit.proto`
  - 导入 constant proto 文件 `--proto_path="proto/account"` + `import "constant.proto"`
  - 在 `pb/account` 目录下生成的 pb 文件 `--go_out=. --go-grpc_out=.` + `go_package="pb/account;account"`

- `greeter/helloworld.proto`
  - 导入 hello proto 文件 `--proto_path="proto/greeter"` + `import "msg/hello.proto"`
  - 在 `pb/greeter` 目录下生成的 pb 文件 `--go_out=. --go-grpc_out=.` + `go_package="pb/greeter;greeter"`

注：`proto_path` 一般设置为 main proto 文件所在的目录。

## 执行测试

TODO:
