# 自定义 Kubernetes Ingress Controller

> Refer: <https://github.com/cnych/simple-ingress/>

## 项目结构

- `server.go`: 代理 ingress 请求到后面 service 服务
- `route.go`: 路由表 (host => path => service_name:service_port)
- `watcher.go`: 监听 k8s 组件变化，更新路由表

## 部署

TODO:

## k8s ingress

ingress 例子1

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

ingress 例子2

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

## k8s cert

TODO:

