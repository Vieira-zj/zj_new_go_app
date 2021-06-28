# TCP server

1. Start tcp server

```sh
go run main.go
```

2. Test TCP server

By telnet:

```sh
telnet 127.0.0.1 8080
> hello world
```

By go test:

```sh
go test -timeout 30s -run ^TestEchoHandler$ demo.hello/apps/tcpserver/pkg -v -count=1
```

