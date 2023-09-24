# TCP server

1. Start Tcp server

```sh
go run main.go
```

2. Test Tcp server

- Telnet

```text
telnet 127.0.0.1 8080
> foo
> bar
```

- Tcp client by go test

```sh
go test -timeout 30s -run ^TestEchoHandler$ demo.hello/apps/tcpserver/pkg -v -count=1
```

