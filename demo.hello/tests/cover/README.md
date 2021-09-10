# Golang Coverage

> `go tool cover` 源码 git: <https://github.com/golang/tools.git>

## 流程

1. 通过 ast 解析 `mock.go`，并注入 counter 和 var `GoCover`。参考：`mock_cover.go`
  - 执行注入以代码块（block）为单元

2. 执行 `go test` 单元测试，生成覆盖率结果数据 `mock_ut_cover.cov`
  - 覆盖率结果每一行数据与 `mock_cover.go` 中的变量 `GoCover` 对应

3. 基于 ast 解析 `mock.go` 结构化数据 + 覆盖率结果数据，生成覆盖率 func/html 报告

