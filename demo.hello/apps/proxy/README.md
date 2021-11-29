# Proxy Server

代理 http 和 https 请求。

> Refer: <https://github.com/lyyyuna/xiaolongbaoproxy>
> 

## Test

Http proxy test:

```sh
# start echo server
./echoserver
# test raw api
curl http://127.0.0.1:8081/

# test api by proxy
curl -x http://localhost:8082 http://127.0.0.1:8081
```

Https proxy test:

```sh
curl https://www.baidu.com

# test api by proxy
curl -x http://localhost:8082 https://www.baidu.com
```

