# 自定义 Kubernetes Ingress Controller

> Refer: <https://github.com/cnych/simple-ingress/>

## Kubernetes Ingress

Ingress 对象，只是一个集群的资源对象而已，并不会去真正处理我们的请求，这个时候我们还必须安装一个 Ingress Controller，该控制器负责读取 Ingress 对象的规则并进行真正的请求处理。简单来说就是 Ingress 对象只是一个声明，Ingress Controllers 就是真正的实现。

对于 Ingress Controller 有很多种选择，比如 traefik、或者 ingress-nginx 等等。

### Ingress 配置例子

1. ingress http 例子

```yaml
spec:
  rules:
  - host: k8s.ingressdemo.com
    http:
      paths:
      - path: /echo
        pathType: Prefix
        backend:
          service:
            name: echo-ingress
            port:
              number: 8080
      - path: /hello
        pathType: Prefix
        backend:
          service:
            name: hello-ingress
            port:
              number: 8080
```

2. ingress backend 例子

```yaml
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          serviceName: s1
          servicePort: 80
```

3. ingress https 例子

```yaml
spec:
  tls:
  - hosts: "*.qikqiak.com"
    secretName: qikqiak-tls
  rules:
  - host: who.qikqiak.com
    http:
      paths:
      - path: /
        backend:
          serviceName: whoami
          servicePort: 80
```

## 自定义 Ingress Controller

自定义的控制器主要需要实现以下几个功能：

1. 通过 Kubernetes API 查询和监听 Service、Ingress 以及 Secret 这些对象；
2. 加载 TLS 证书用于 HTTPS 请求；
3. 根据加载的 Kubernetes 数据构造一个用于 HTTP 服务的路由，当然该路由需要非常高效，因为所有传入的流量都将通过该路由；
4. 在 80 和 443 端口上监听传入的 HTTP 请求，根据路由查找对应的后端服务，然后代理请求和响应。443 端口将使用 TLS 证书进行安全连接。

### 项目结构

- `server.go`: 代理 ingress 请求到后端 service 服务
- `watcher.go`: 监听 ingress, service, secret 组件变化，更新路由表
- `route.go`: 路由表 `host + path => service_name:service_port`

### 部署和 http 测试

1. 部署一个 whoami 的应用：

```sh
# deploy app
kubectl apply -f whoami.yaml
# test api
kubectl run debug-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
wget 'http://whoami' -O -
```

2. 创建 Ingress Controller:

```sh
kubectl apply -f k8s-ingress-controller.yaml
kubectl get pods -l app=ingress-controller -n k8s-test
```

3. 为 whoami 服务创建一个 Ingress 对象：

```sh
kubectl apply -f whoami-ingress.yaml
```

4. 将域名 `who.test.com` 解析到部署的 Ingress Controller 的 Pod 节点上，就可以直接访问了。

```sh
# test by default ingress
curl -v "http://who.test.com/"
# test by custom ingress controller
curl -v "http://who.test.com:8080/"
```

5. 查看代理服务日志：

```sh
kubectl logs -f k8s-simple-ingress-controller-wk7h8 -n k8s-test
```

### https 测试

TODO:

### k8s.io api 版本问题

Ingress yaml 部署使用 api 版本为 `networking.k8s.io/v1`，在 client-go 中：

- 使用 k8s.io 版本为 `k8s.io/api/networking/v1`时，`Ingresses().Lister()` 查询结果为0；
- 使用 `k8s.io/api/extensions/v1beta1` 查询结果正常。

## 补充 - K8s

### TLS Certificate

#### Create tls cert

1. 模拟CA产生一个私钥 `ca.key`
2. 使用CA私钥生成CA证书 `ca.crt`
3. 为服务端生成一个私钥 `server.key`
4. 使用私钥生成一个证书请求（包含服务端信息及公钥） `server.csr`
5. 使用CA证书和私钥签署服务器证书 `server.crt`

#### Apply tls ingress

1. Create private key and cert file

2. Apply k8s resource secret

```sh
kubectl create secret tls testsecret-tls --cert=tls/tls.crt --key=tls/tls.key -n k8s-test
```

3. Apply ingress with tls

```yaml
tls:
- hosts:
  - zj.ssltest.com
secretName: testsecret-tls
```

4. Test tls ingress

```sh
curl -vvv https://zj.ssltest.com/ --cacert ${HOME}/.ssl/ca.crt
```

#### Apply https server (admission webhook)

1. Create `server.csr`, `server-key.pem`

2. Apply k8s resource csr (CertificateSigningRequest)

3. Get csr approved (csr to crt)

```sh
kubectl certificate approve ${csrName}
```

4. Create `server-cert.pem`

```sh
kubectl get csr my-svc.my-namespace -o jsonpath='{.status.certificate}' | base64 --decode > server-cert.pem
```

5. Create k8s resource secret

6. Apply deployment with secret mounted

### K8s Informer

#### list-watch

- list 就是调用资源的 list API 罗列资源（获取全量数据），基于HTTP短链接实现；
- watch 则是调用资源的 watch API 监听资源变更事件（增量数据），基于HTTP长链接实现。watch 原理就是 chunked transfer encoding.

#### Informer

client-go 中的 Informer 对 list-watch 操作做了封装。所谓 Informer，其实就是一个带有本地缓存和索引机制的、可以注册 EventHandler 的 client, 本地缓存被称为 `Store`，索引被称为 `Index`。

