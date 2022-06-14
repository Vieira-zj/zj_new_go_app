# Funcs Diff

1. Diff fucns between src and dst `.go` files. (same files with diff commit of code)
2. Combine `.cov` files of diff version code base on funcs diff results.

## Test: Funcs Diff

### Go Tools

Dependency go tools:

- go list
- gofmt

### Build Go Project for Test

1. Create project dir `go_funcdiff_space`

2. Add `go.mod`

```text
module demo.funcdiff

go 1.16
```

3. Copy `test/src1/main.go` and `test/src2/main.go` to project

### Run Funcs Diff

1. Format go files
  - filter out empty and comments line
  - go format
2. Diff funcs between src and dst go files, and output:

```text
src: [demo.funcdiff/src1/main_format.go:fnHello] [16:1,18:2]
dst: [demo.funcdiff/src2/main_format.go:fnHello] [8:1,10:2]
diff: same

src: [demo.funcdiff/src1/main_format.go:fnConditional] [25:1,31:2]
dst: [demo.funcdiff/src2/main_format.go:fnConditional] [19:1,25:2]
diff: same

src: [demo.funcdiff/src1/main_format.go:fnDel] [22:1,24:2]
diff: remove

src: [demo.funcdiff/src1/main_format.go:fnChange] [19:1,21:2]
dst: [demo.funcdiff/src2/main_format.go:fnChange] [16:1,18:2]
diff: change

src: [demo.funcdiff/src1/main_format.go:(p)fnHello] [13:1,15:2]
dst: [demo.funcdiff/src2/main_format.go:(p)fnHello] [32:1,34:2]
diff: change

dst: [demo.funcdiff/src2/main_format.go:fnAdd] [11:1,15:2]
diff: add
```

## Test: Merges Cov Files

> TODO:

