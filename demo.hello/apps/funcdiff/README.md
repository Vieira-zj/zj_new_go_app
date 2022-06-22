# Funcs Diff

1. Diff fucns between src and dst `.go` files. (same files by diff commits)

2. Combine `.cov` files of diff version code base on funcs diff results.
  - 合并不同版本代码的覆盖率数据数据。如 bug fix 后的覆盖率数据与上一个版本（全量回归测试）的覆盖率数据合并。

## Coverage Profile 合并方案

### 方案1: 基于 line 维度合并

1. git 获取 diff line => same_lines, diff_lines

2. 匹配 src_line 和 dst_line, 构建 diff_entry => same_entries, diff_entries
  - line struct: `{"file":"path/main.go", "position":1, "content":"..."}`
  - entry struct: `{"src_line":{...}, "dst_line":{...}, "diff":"same/add/delete/change"}`
  - 考虑文件 rename 的情况
  - 由于 add/delete 的原因，部分代码只有行号变化

3. 解析 `profile.cov` 文件（block 代码块维度），将 profile 转换为代码行维度的覆盖率数据
  - cover struct: `{"position":{...}, "count":0}`

4. 将 line 与 cover 数据关联 => same_cover_entries, diff_cover_entries
  - 通过 file+position 关联数据
  - cover_line struct: `{"file":"path/main.go", "position":1, "content":"...", "cover":{...}}`
  - entry struct: `{"src_cover_line":{...}, "dst_cover_line":{...}, "diff":"same/add/delete/change"}`

5. 合并 same_cover_entries 中的覆盖率数据 => merged_profile_for_same
  - 合并 src_cover_line 和 dst_cover_line 覆盖率数据，取较大的 cover 值

6. 合并 diff_cover_entries 中的覆盖率数据 => merged_profile_for_diff
  - 合并规则
    - line delete: 不需要处理
    - line add/change: 取 dst_cover_line 中的覆盖率数据
    - line onchange: 合并 src_cover_line 和 dst_cover_line 覆盖率数据，取较大的 cover 值

7. merged_profile_for_same + merged_profile_for_diff 得到合并的覆盖率数据 => merged_profile

优点：

1. line 维度比较通用，支持同不语言，如 go, js
  - js 覆盖率数据是 statement + function + branch 维度
  - go 覆盖率数据是 block 维度

缺点：

1. 解析生成 line 维度的数据较复杂
2. go 覆盖率工具链是基于 block 维度，使用 line 维度后不能复用
3. 合并数据精准度问题
  - 当 func 变化后，该 func 中包含代码行的覆盖率都应该设置为 0
  - 上下游的 func 代码覆盖率是否应该设置为 0 ?

### 方案2: func 维度合并

1. git 获取 diff file => nochange_files, diff_files

2. ast 解析 diff 文件，得到 diff func 数据 => nochange_entries, diff_entries
  - 基于 filepath + funcname 匹配 func
  - func struct: `{"file":"path/main.go", "func_name":"main", "position":"start_line,start_col,end_line,end_col"}`
  - entry struct: `{"src_func":{...}, dst_func:{...}, "diff":"same/add/delete/change"}`
  - 过滤空行、注释行

3. 解析 `profile.cov` 文件，将 block 与 func 关联 => nochange_entries, diff_entries
  - cover_func struct: `{"file":"path/main.go", "func_name":"main", "position":"start_line,start_col,end_line,end_col"}, "cover_blocks":[...]`
  - entry struct: `{"src_cover_func":"{...}", "dst_cover_func":"{...}", "diff":"same/add/delete/change"}`

4. nochange_entries 合并覆盖率数据 => nochange_merged_profile
  - 合并 block 覆盖率数据，取较大的 cover 值

5. diff_entries 合并覆盖率数据 => diff_merged_profile
  - 合并规则
    - case func delete: 不需要处理
    - case func add/change: 取 dst_cover_func 中 block 覆盖率数据
    - case func nochange: 合并 src_cover_func 和 dst_cover_func 的 block 覆盖率数据，取较大的 cover 值

6. nochange_merged_profile + diff_merged_profile 得到合并的覆盖率数据 => merged_profile

优点：

1. 基于 ast 解析，不受 line 行号变化影响
2. 复用 go 覆盖率工具链，如使用 go tool 生成 func 和 html 覆盖率报告

问题：

1. 该方案只适用于 golang 覆盖率数据合并

## 方案2 - Demo

Case: run func diff and profile merge for a diff go file.

### Test Env

Source files:

```text
└── test
    ├── src1
    │   ├── func.rpt
    │   ├── main.go
    │   ├── main_test.go
    │   └── profile.cov
    └── src2
        ├── func.rpt
        ├── func_merged.rpt
        ├── main.go
        ├── main_test.go
        ├── profile.cov
        └── profile_merged.cov
```

Function definitions in `src1/main.go` and `src2/main.go`:

```text
(p person) fnToString() => same
fnHello()               => same
fnConditional()         => same, test for if cond: run true in src, and false for dst
fnAdd()                 => func added in dst
fnChange()              => func changed in dst
fnDel()                 => func deleted in dst
```

### Diff Funcs

1. Run diff funcs, and results:

```text
src: [demo.hello/apps/funcdiff/test/src1/main.go:fnHello] [22:1,33:2]
dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnHello] [12:1,23:2]
diff: same

src: [demo.hello/apps/funcdiff/test/src1/main.go:fnConditional] [43:1,55:2]
dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnConditional] [35:1,47:2]
diff: same

src: [demo.hello/apps/funcdiff/test/src1/main.go:(p person) fnToString] [14:1,16:2]
dst: [demo.hello/apps/funcdiff/test/src2/main.go:(p person) fnToString] [54:1,56:2]
diff: same

src: [demo.hello/apps/funcdiff/test/src1/main.go:fnDel] [39:1,41:2]
diff: remove

src: [demo.hello/apps/funcdiff/test/src1/main.go:fnChange] [35:1,37:2]
dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnChange] [31:1,33:2]
diff: change

dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnAdd] [25:1,29:2]
diff: add
```

### Add Profile Blocks

1. Run go test with `-cover`, and generate coverage results `profile.cov`.
2. Parse `profile.cov` file.
3. Link profile blocks to func.

Output diff entries:

```text
src: [demo.hello/apps/funcdiff/test/src1/main.go:fnHello] [22:1,33:2]
  {StartLine:22 StartCol:70 EndLine:23 EndCol:19 NumStmt:1 Count:1 isLink:false}
  {StartLine:23 StartCol:19 EndLine:25 EndCol:3 NumStmt:1 Count:0 isLink:false}
  {StartLine:27 StartCol:2 EndLine:27 EndCol:34 NumStmt:1 Count:1 isLink:false}
  {StartLine:27 StartCol:34 EndLine:29 EndCol:3 NumStmt:1 Count:1 isLink:false}
  {StartLine:30 StartCol:2 EndLine:32 EndCol:9 NumStmt:2 Count:1 isLink:false}
dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnHello] [12:1,23:2]
  {StartLine:12 StartCol:48 EndLine:13 EndCol:19 NumStmt:1 Count:0 isLink:false}
  {StartLine:13 StartCol:19 EndLine:15 EndCol:3 NumStmt:1 Count:0 isLink:false}
  {StartLine:18 StartCol:2 EndLine:18 EndCol:34 NumStmt:1 Count:0 isLink:false}
  {StartLine:18 StartCol:34 EndLine:20 EndCol:3 NumStmt:1 Count:0 isLink:false}
  {StartLine:21 StartCol:2 EndLine:22 EndCol:9 NumStmt:2 Count:0 isLink:false}
diff: same

src: [demo.hello/apps/funcdiff/test/src1/main.go:fnConditional] [43:1,55:2]
  {StartLine:43 StartCol:57 EndLine:45 EndCol:10 NumStmt:1 Count:1 isLink:false}
  {StartLine:45 StartCol:10 EndLine:47 EndCol:3 NumStmt:1 Count:1 isLink:false}
  {StartLine:47 StartCol:8 EndLine:48 EndCol:10 NumStmt:1 Count:0 isLink:false}
  {StartLine:48 StartCol:10 EndLine:49 EndCol:36 NumStmt:1 Count:0 isLink:false}
  {StartLine:49 StartCol:36 EndLine:50 EndCol:15 NumStmt:1 Count:0 isLink:false}
  {StartLine:52 StartCol:4 EndLine:52 EndCol:30 NumStmt:1 Count:0 isLink:false}
dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnConditional] [35:1,47:2]
  {StartLine:35 StartCol:31 EndLine:37 EndCol:10 NumStmt:1 Count:1 isLink:false}
  {StartLine:37 StartCol:10 EndLine:39 EndCol:3 NumStmt:1 Count:0 isLink:false}
  {StartLine:39 StartCol:8 EndLine:40 EndCol:10 NumStmt:1 Count:1 isLink:false}
  {StartLine:40 StartCol:10 EndLine:41 EndCol:36 NumStmt:1 Count:1 isLink:false}
  {StartLine:41 StartCol:36 EndLine:42 EndCol:15 NumStmt:1 Count:0 isLink:false}
  {StartLine:44 StartCol:4 EndLine:44 EndCol:30 NumStmt:1 Count:1 isLink:false}
diff: same

src: [demo.hello/apps/funcdiff/test/src1/main.go:(p person) fnToString] [14:1,16:2]
  {StartLine:14 StartCol:37 EndLine:16 EndCol:2 NumStmt:1 Count:0 isLink:false}
dst: [demo.hello/apps/funcdiff/test/src2/main.go:(p person) fnToString] [54:1,56:2]
  {StartLine:54 StartCol:37 EndLine:56 EndCol:2 NumStmt:1 Count:1 isLink:false}
diff: same

src: [demo.hello/apps/funcdiff/test/src1/main.go:fnDel] [39:1,41:2]
  {StartLine:39 StartCol:14 EndLine:41 EndCol:2 NumStmt:1 Count:1 isLink:false}
diff: remove

src: [demo.hello/apps/funcdiff/test/src1/main.go:fnChange] [35:1,37:2]
  {StartLine:35 StartCol:17 EndLine:37 EndCol:2 NumStmt:1 Count:1 isLink:false}
dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnChange] [31:1,33:2]
  {StartLine:31 StartCol:17 EndLine:33 EndCol:2 NumStmt:1 Count:0 isLink:false}
diff: change

dst: [demo.hello/apps/funcdiff/test/src2/main.go:fnAdd] [25:1,29:2]
  {StartLine:25 StartCol:14 EndLine:29 EndCol:2 NumStmt:3 Count:1 isLink:false}
diff: add
```

Unlinked profile blocks:

```text
{StartLine:8 StartCol:20 EndLine:10 EndCol:2 NumStmt:1 Count:0 isLink:false}
{StartLine:58 StartCol:13 EndLine:61 EndCol:2 NumStmt:1 Count:0 isLink:false}
```

### Merge Profiles

1. Merge profile blocks between src and dst.
2. Generate merged `profile_merged.cov` file with blocks sorted by path and line,col.
3. Create and check merged coverage func report.

Src func report:

```text
demo.hello/apps/funcdiff/test/src1/main.go:14:  fnToString  0.0%     => same
demo.hello/apps/funcdiff/test/src1/main.go:22:  fnHello   83.3%      => same
demo.hello/apps/funcdiff/test/src1/main.go:35:  fnChange  100.0%     => change
demo.hello/apps/funcdiff/test/src1/main.go:39:  fnDel   100.0%       => delete, not include in merged results
demo.hello/apps/funcdiff/test/src1/main.go:43:  fnConditional 33.3%  => cover true cond
demo.hello/apps/funcdiff/test/src1/main.go:57:  main    0.0%         => exclude main
total:            (statements)  56.2%
```

Dst func report:

```text
demo.hello/apps/funcdiff/test/src2/main.go:12:  fnHello   0.0%       => same
demo.hello/apps/funcdiff/test/src2/main.go:25:  fnAdd   100.0%       => add
demo.hello/apps/funcdiff/test/src2/main.go:31:  fnChange  0.0%       => change
demo.hello/apps/funcdiff/test/src2/main.go:35:  fnConditional 66.7%  => cover false cond
demo.hello/apps/funcdiff/test/src2/main.go:54:  fnToString  100.0%   => same
demo.hello/apps/funcdiff/test/src2/main.go:58:  main    0.0%         => exclude main
total:            (statements)  44.4%
```

Merged func report `func_merged.rpt`:

```text
demo.hello/apps/funcdiff/test/src2/main.go:12:  fnHello   83.3%      => same, from src
demo.hello/apps/funcdiff/test/src2/main.go:25:  fnAdd   100.0%       => add, from dst
demo.hello/apps/funcdiff/test/src2/main.go:31:  fnChange  0.0%       => change, from dst
demo.hello/apps/funcdiff/test/src2/main.go:35:  fnConditional 83.3%  => same, merge src and dst
demo.hello/apps/funcdiff/test/src2/main.go:54:  fnToString  100.0%   => same, from dst
demo.hello/apps/funcdiff/test/src2/main.go:58:  main    0.0%         => exclude main
total:            (statements)  77.8%
```

