# Golang Coverage

> `go tool cover` 源码：<https://github.com/golang/tools.git>
>

## 流程

1. 通过 ast 解析 `mock.go`，并注入 counter 和 `var GoCover`。参考：`mock_cover.go`
  - 执行注入以代码块（block）为单元
  - `func isOk` 包含3个 blocks

2. 执行 `go test` 单元测试，生成覆盖率结果数据 `mock_ut_cover.cov`
  - 覆盖率结果每一行数据与 `mock_cover.go` 中的变量 `GoCover` 对应

3. 以 `.go` 文件为粒度，关联 `.go func` 和 `.cov block`，基于 profile 生成 func 覆盖率报告
  - 3.1 解析 `.cov` 覆盖率数据，生成 `files:filename:profile:blocks`
  - 3.2 通过 ast 解析 `.go` 文件，生成 `filename:func`
  - 3.3 func 关联到多个 blocks, 然后基于 block profile `NumStmt` 和 `Count` 计算出 `CoveredStmt` 和 `TotalStmt`，最后得到 func 覆盖率数据
  - 注：保证本地代码与生成 `.cov` 文件的代码一致，func 会匹配到多个 blocks

### 覆盖率报告

- func cover output: func粒度，语句覆盖

- html cover output:
  - go文件粒度，语句覆盖
  - 代码染色结果，分支覆盖

