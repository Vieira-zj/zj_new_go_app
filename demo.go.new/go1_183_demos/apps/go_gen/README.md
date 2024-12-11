# Go Generate by Stringer

1. Install go tool stringer

```sh
go install golang.org/x/tools/cmd/stringer@latest
```

2. Add `go:generate` announce

```go
//go:generate stringer -type=AutoPill -output=pill_string.go

//go:generate stringer -type=AutoPill -linecomment
```

3. Run `go generate`
