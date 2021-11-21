# K8S client-go app

## Build env

1. Create a namespace for test.

```sh
kubectl create namespace -h
kubectl create namespace k8s-test
```

2. Create service deployments.

```sh
kubectl create -f deploy/sidecar_deploy.yaml
kubectl create -f deploy/echoserver_deploy.yaml
```

## k8s 集群内使用 client

k8s 集群内使用 client 需要 `namespace.default-serviceaccount` 有 pod get,list 权限。

Role and ClusterRole:

- `Role, RoleBinding`: set authorization for a namespace.
- `ClusterRole, ClusterRoleBinding`: set authorization for the whole k8s cluster.

1. Create a role with pod list auth.

```sh
kubectl create -f deploy/pods_reader_role.yaml

# check role
kubectl get role -n k8s-test
kubectl describe role/pods-reader -n k8s-test
```

Role definition yaml:

```yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: k8s-test
  name: pods-reader
rules:
- apiGroups: [""] # "" indicates the core API group
  resources: ["services"]
  verbs: ["get", "watch", "list"]
```

2. Create a rolebinding which binds pre created role with user `k8s-test-ns.default`.

```sh
# cmd
kubectl create rolebinding -h
kubectl create rolebinding client-pods-reader --role=pods-reader --serviceaccount=k8s-test:default -n k8s-test
# yaml
kubectl create -f deploy/pods_reader_rolebind.yaml

# check rolebinding
kubectl get role,rolebinding -n k8s-test
kubectl describe rolebinding/client-pods-reader -n k8s-test
```

> Note: if no auth set, we will get error:
>
> `pods is forbidden: User "system:serviceaccount:k8s-test:default" cannot list resource "pods" in API group "" in the namespace "k8s-test"`

3. Copy k8s client bin into a pod and list pods.

```sh
kubectl run debug-pod -it --rm --restart=Never -n=k8s-test --image=busybox:1.30 sh

# copy
cd k8sclient; GOOS=linux GOARCH=amd64 go build .
kubectl cp k8sclient debug-pod:/tmp -n k8s-test
# run
root# ./k8sclient
```

## Inject sidecar

1. Check echoserver pod status.

```text
NAME                                   READY   STATUS    RESTARTS   AGE
pod/echoserver-test-6fb88b9b9b-8rxqx   1/1     Running   0          22m
```

2. Inject a sidecar into echoserver pod.

```sh
go run main.go -i
```

3. Check pod status after inject.

```text
NAME                                   READY   STATUS        RESTARTS   AGE
pod/echoserver-test-6bfcff48df-mr4bz   2/2     Running       0          7s
pod/echoserver-test-6fb88b9b9b-8rxqx   1/1     Terminating   0          23m
```

```sh
kubectl describe pod/echoserver-test-6bfcff48df-mr4bz -n k8s-test
```

4. Check pod and sidecar container ip are matched.

```sh
# pod ip
kubectl get pods -o wide -n k8s-test
NAME                               READY   STATUS    RESTARTS   AGE     IP           NODE       NOMINATED NODE   READINESS GATES
echoserver-test-6bfcff48df-mr4bz   2/2     Running   0          4m17s   172.17.0.5   minikube   <none>           <none>

# sidecar container ip
kubectl exec echoserver-test-6bfcff48df-mr4bz -it -n k8s-test -c inject-container-test sh
root# ifconfig
eth0 inet addr:172.17.0.5
```

> Note: we cannot directly inject a sidecar into a pod by modify pod yaml, then we will get error:
Forbidden: pod updates may not add or remove containers
> 
> We should inject a sidecar by update deployment.

## Inject init container for tcp forward

1. Check echoserver deployment.

```sh
kubectl get deploy -n k8s-test
```

2. Inject init container into echoserver pod.

```sh
go run main.go -f
```

Tcp forword from init container:

```sh
# 将访问 8080 端口（echoserver服务）的请求转发到 8088 端口（代理服务）
# addr:8080 -> prerouting -> 127.0.0.1:8088 proxy -> 127.0.0.1:8080 echoserver
iptables -t nat -A PREROUTING -p tcp --dport 8080 -j REDIRECT --to-port 8088
```

3. Check pod status.

```sh
kubectl get all -n k8s-test

NAME                                   READY   STATUS            RESTARTS   AGE
pod/echoserver-test-7d654648cf-rjdq8   0/2     PodInitializing   0          4s
pod/sidecar-test-675c9d6f86-smp65      2/2     Running           0          62m
```

4. Deploy a `httpproxy` in sidecar container which listen at `8088`.

```sh
kubectl cp httpproxy echoserver-test-6667f8d8cf-wzcqw:/tmp -c inject-container-test -n k8s-test
kubectl exec -it echoserver-test-6667f8d8cf-wzcqw -c inject-container-test -n k8s-test sh

# run http proxy (default listen 8088) for echoserver 127.0.0.1:8080
root# cd /tmp; ./httpproxy -t 127.0.0.1:8080 &
```

5. Test echoserver api by both origin and proxy addr.

- Local: login echoserver pod and test api local

```sh
kc exec -it echoserver-test-6667f8d8cf-wzcqw -c echo-server -n k8s-test sh

# 直接访问服务（不走forward）
root# curl -v http://127.0.0.1:8080/test
# 通过代理访问服务 proxy 8088 -> echoserver 8080
root# curl -v http://127.0.0.1:8088/test
```

> Note: it does not go iptable forward chain when access port 8080 from localhost.
>

- Remote: start a debug pod and test echoserver

```sh
# get pod ip
kubectl get pod -o wide -n k8s-test
NAME                               READY   STATUS    RESTARTS   AGE   IP           NODE       NOMINATED NODE   READINESS GATES
echoserver-test-6667f8d8cf-wzcqw   2/2     Running   0          43m   172.17.0.5   minikube   <none>           <none>

kubectl run debug-pod -it --rm --restart=Never -n=k8s-test --image=busybox:1.30 sh
# 直接通过代理访问服务 proxy 8088 -> echoserver 8080
root# wget "http://172.17.0.5:8088/" -S -O -
# 通过 iptables 规则走代理访问服务 origin addr 8080 -> forword proxy 8088 -> echoserver 8080
root# wget "http://172.17.0.5:8080/" -S -O -
```

------

## Remote exec debug

### 方案一

原理：

1. 获取目标pod所包含container的pid, 通过 pid 构建 nsexec 命令
  - k8s-client -> target pod -> node ip + container id -> docker-client -> pid -> command: nsexec -p /proc/{pid}/ns/pid -- {cmd}

2. 部署拥有特权的 deamonset pod, 通过该pod执行 nsexec 命令
  - k8s-client -> daemonset pod (with privilege) -> spdy executor -> run cmd within pid ns (by "nsexec")

步骤：

1. Build debug-daemon image

```sh
docker build -t zhengjin/debug-daemon:v1.0 -f images/Dockerfile .
```

2. Create debug DaemonSet with privileged

```sh
# deploy
kubectl apply -f deploy/debug_daemon_deploy.yaml
# check deploy
kubectl get daemonset/debug-daemon -n k8s-test
```

Note: since Kubernetes 1.6, DaemonSets do not schedule on master nodes by default. In order to schedule it on master, you have to add a toleration into the Pod spec section:

```sh
tolerations:
- key: node-role.kubernetes.io/master
  effect: NoSchedule
```

3. Start a pod for test

```sh
kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
```

4. Run remote exec debug test

```sh
go test -timeout 10s -run ^TestExecCmdBypass$ demo.hello/k8s/client/pkg -v -count=1
```

### 方案二

TODO:

