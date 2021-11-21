# K8S Webshell Demo

> Refer: <https://github.com/maoqide/kubeutil>
>

## Overview

原理：

1. 通过实现 `io.Reader` 和 `io.Writer` 接口与 pod executor 交互。
2. 前端与服务端通过 `websocket` 交互。

数据流：

```text
browser xterm.js <=> websocket (client) <=> websocket (server) <=> terminal session reader/writer <=> executor stdin/stdout <=> pod
```

## Test Steps

1. start a debug pod

```sh
kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
```

2. run webshell server `go run main.go`

3. test server APIs

```sh
# namespaces
curl -v "http://localhost:8090/query/ns" | jq .
# pods
curl -v "http://localhost:8090/query/pods?ns=k8s-test" | jq .
# containers
curl -v "http://localhost:8090/query/containers?ns=k8s-test&pod=containers-pod" | jq .
```

4. `cd webshell/static`, and open `terminal.html` in chrome

5. select `namespace`, `pod` and `container`

6. `connect` to open a webshell terminal session to pod

