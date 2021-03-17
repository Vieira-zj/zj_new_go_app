# Http Proxy

代理http请求：

- 记录请求信息
- 透传请求

## Test Proxy

1. Start a backend mock server.

2. Test mock server apis.

```sh
curl -v "http://127.0.0.1:17891/demo/1?userid=001&username=foo"
curl -v -X POST "http://127.0.0.1:17891/demo/3" -H "Content-Type:application/json" --data-binary @test.json
```

`test.json`:

```json
{
  "server_list": [
    {
      "server_name": "svr_a_002",
      "server_ip": "127.0.1.2"
    },
    {
      "server_name": "svr_a_013",
      "server_ip": "127.0.1.13"
    }
  ],
  "server_group_id": "svr_grp_001"
}
```

3. Start http proxy service for mock server.

```sh
go run main.go -t 127.0.0.1:17891
```

Compile proxy bin:

```sh
cd httpproxy; GOOS=linux GOARCH=amd64 go build .
```

4. Test proxy server by using addr `127.0.0.1:8088` instead of origin addr `127.0.0.1:17891`.

```sh
curl -v "http://127.0.0.1:8088/demo/1?userid=001&username=foo"
curl -v -X POST "http://127.0.0.1:8088/demo/3" -H "Content-Type:application/json" --data-binary @test.json
```

