# Kubernetes Operator

## Nginx Operator Demo

### Env build

1. Install operator-sdk env on mac.

```text
$ brew install operator-sdk
```

2. Cli `operator-sdk`.

```text
$ operator-sdk version
operator-sdk version: "v1.9.0", commit: "205e0a0c2df0715d133fbe2741db382c9c75a341", kubernetes version: "v1.20.2", go version: "go1.16.5", GOOS: "darwin", GOARCH: "amd64"

$ operator-sdk -h
init     Initialize a new project
create   Scaffold a Kubernetes API or webhook
```

3. Create an operator project.

```text
$ operator-sdk init -h
$ operator-sdk init --plugins go/v3 --domain example.com --repo github.com/example/nginx-operator --owner "ZhengJin"
```

4. Create a CRD api.

```text
$ operator-sdk create api -h
$ operator-sdk create api --group proxy --version v1alpha1 --kind Nginx --resource --controller
```

### 定义 CRD

1. Implement API.

- `api/v1alpha1/nginx_types.go`

```golang
type NginxSpec struct {}
type NginxStatus struct {}
```

2. 执行下面命令生成 CRD 资源相关代码 `api/v1alpha1/zz_generated.deepcopy.go`。

```text
$ make generate
```

注：每次修改 CRD 的定义后都需执行该命令。

3. 执行下面的命令生成 CRD manifests 定义文件 `nginx-operator/config/crd/bases/proxy.example.com_nginxes.yaml`

```text
$ make manifests
```

### 定义 Controller

1. Implement Controller.

- `controllers/nginx_controller.go`

```golang
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// ...
}
```

2. 添加 rbac 定义。

当前 controller 会创建 deployment 和 service, 因此需要添加对应的 rbac 定义（参考下面创建 CRD 实例时，日志中有权限错误）：

```golang
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;delete;get;list;update;patch;watch
//+kubebuilder:rbac:groups="",resources=deployments;services,verbs="*"
```

执行 `make manifests` 生成配置文件 `config/rbac/role.yaml`。

### Build and Deploy Operator

1. Build Operator image. (from `Dockerfile`)

```text
# connect to minikube docker socket
$ eval $(minikube -p minikube docker-env)

$ make docker-build IMG=example.docker.io/nginx-operator:v0.1
Step 9/14 : RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go
Successfully built 32ec2c7963fc
Successfully tagged example.docker.io/nginx-operator:v0.1

$ docker images
example.docker.io/nginx-operator:v0.1
```

Custom image tag name in `Makefile`:

```text
IMG ?= proxy.example.com/nginx-operator:v0.1
```

2. Deploy Operator in k8s.

```text
$ make deploy IMG=example.docker.io/nginx-operator:v0.1

namespace/nginx-operator-system created
customresourcedefinition.apiextensions.k8s.io/nginxes.proxy.example.com created
serviceaccount/nginx-operator-controller-manager created
role.rbac.authorization.k8s.io/nginx-operator-leader-election-role created
clusterrole.rbac.authorization.k8s.io/nginx-operator-manager-role created
clusterrole.rbac.authorization.k8s.io/nginx-operator-metrics-reader created
clusterrole.rbac.authorization.k8s.io/nginx-operator-proxy-role created
rolebinding.rbac.authorization.k8s.io/nginx-operator-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/nginx-operator-manager-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/nginx-operator-proxy-rolebinding created
configmap/nginx-operator-manager-config created
service/nginx-operator-controller-manager-metrics-service created
deployment.apps/nginx-operator-controller-manager created
```

3. Check deploy Operator components in k8s.

```text
$ kubectl get ns

$ kubectl get svc,deployment,pod -n nginx-operator-system
NAME                                                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/nginx-operator-controller-manager-metrics-service   ClusterIP   10.106.213.52   <none>        8443/TCP   18m
NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx-operator-controller-manager   1/1     1            1           18m
NAME                                                    READY   STATUS    RESTARTS   AGE
pod/nginx-operator-controller-manager-b8f479c94-2ttm6   2/2     Running   0          17m
```

If nginx-operator-controller-manager pod error `ImagePullBackOff`:

```text
Failed to pull image "example.docker.io/nginx-operator:v0.1": rpc error: code = Unknown desc = Error response from daemon: pull access denied for controller, repository does not exist or may require 'docker login': denied: requested access to the resource is denied
```

Update deploy config `imagePullPolicy: IfNotPresent`:

```text
$ kubectl edit deployment/nginx-operator-controller-manager -n nginx-operator-system
```

4. Check logs of nginx-operator-controller-manager pod.

```text
$ kubectl logs -f nginx-operator-controller-manager-7c7cbfc886-4x6sb -c manager -n nginx-operator-system
...
Starting Controller     {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx"}
Starting workers        {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "worker count": 1}
```

### Deploy a CRD instance

1. Create CRD Nginx manifest.

- `config/samples/proxy_v1alpha1_nginx.yaml`

```yaml
apiVersion: proxy.example.com/v1alpha1
kind: Nginx
metadata:
  name: nginx-app
spec:
  size: 1
  image: nginx:1.7.9
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30002
```

2. Apply a CRD Nginx instance.

```text
$ kubectl apply -f config/samples/proxy_v1alpha1_nginx.yaml
nginx.proxy.example.com/nginx-app created
```

3. Check deploy CRD Nginx instance, and related k8s components.

```text
$ kubectl get nginx/nginx-app -o yaml
status:
  replicas: 1

$ kubectl get all -n default
```

**Note**: if auth error in nginx-operator-controller-manager pod:

```text
$ kubectl logs nginx-operator-controller-manager-b8f479c94-nqsv2 -c manager -n nginx-operator-system
failed to list *v1.Deployment: deployments.apps is forbidden: User "system:serviceaccount:nginx-operator-system:nginx-operator-controller-manager" cannot list resource "deployments" in API group "apps" at the cluster scope
```

Check deploy clusterrole:

```text
$ kubectl describe clusterrolebinding/nginx-operator-manager-rolebinding -n nginx-operator-system
Role:
  Kind:  ClusterRole
  Name:  nginx-operator-manager-role
Subjects:
  Kind            Name                               Namespace
  ----            ----                               ---------
  ServiceAccount  nginx-operator-controller-manager  nginx-operator-system
```

Then add below auth to clusterrole:

```text
$ kubectl describe clusterrole/nginx-operator-manager-role
$ kubectl edit clusterrole/nginx-operator-manager-role

- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - deployments
  - services
  verbs:
  - "*"
```

4. Again, check nginx-operator-controller-manager pod logs.

```text
$ kubectl logs nginx-operator-controller-manager-b8f479c94-2ttm6 -c manager -n nginx-operator-system
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Create CR Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Update CR Nginx Status {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
```

Verify created k8s components of CRD Nginx instance.

```text
$ kubectl get all -n default

$ kubectl describe service/nginx-app
Port:                     <unset>  80/TCP
TargetPort:               80/TCP
NodePort:                 <unset>  30002/TCP
Endpoints:                172.17.0.4:80
```

5. Access nginx by `http://192.168.99.103:30002/`

Note: here use minikube vm ip instead of `localhost`.

### Update CRD Nginx manifest

1. Update `proxy_v1alpha1_nginx.yaml`: 

- `size: 2`
- `nodePort: 30009`

2. Re-deploy CRD Nginx instance.

```text
$ kubectl apply -f config/samples/proxy_v1alpha1_nginx.yaml
nginx.proxy.example.com/nginx-app configured
```

3. Check newly created k8s components of CR instance.

```text
$ kubectl get pod -n default
pod/nginx-app-6f8b78cb8-8jj2m   1/1     Running   0          5s
pod/nginx-app-6f8b78cb8-tg9n4   1/1     Running   0          5s

$ kubectl describe service/nginx-app
Port:                     <unset>  80/TCP
TargetPort:               80/TCP
NodePort:                 <unset>  30009/TCP
Endpoints:                172.17.0.4:80,172.17.0.6:80
```

4. Check nginx-operator-controller-manager pod logs.

```text
$ kubectl logs nginx-operator-controller-manager-b8f479c94-fhtxv -c manager -n nginx-operator-system
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Update CR Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Update CR Nginx Status {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
```

### Clearup CRD

1. Delete CRD Nginx instance.

```text
$ kubectl delete -f config/samples/proxy_v1alpha1_nginx.yaml
nginx.proxy.example.com "nginx-app" deleted
```

2. Remove Operator and related k8s components.

```text
$ make undeploy

namespace "nginx-operator-system" deleted
customresourcedefinition.apiextensions.k8s.io "nginxes.proxy.example.com" deleted
serviceaccount "nginx-operator-controller-manager" deleted
role.rbac.authorization.k8s.io "nginx-operator-leader-election-role" deleted
clusterrole.rbac.authorization.k8s.io "nginx-operator-manager-role" deleted
clusterrole.rbac.authorization.k8s.io "nginx-operator-metrics-reader" deleted
clusterrole.rbac.authorization.k8s.io "nginx-operator-proxy-role" deleted
rolebinding.rbac.authorization.k8s.io "nginx-operator-leader-election-rolebinding" deleted
clusterrolebinding.rbac.authorization.k8s.io "nginx-operator-manager-rolebinding" deleted
clusterrolebinding.rbac.authorization.k8s.io "nginx-operator-proxy-rolebinding" deleted
configmap "nginx-operator-manager-config" deleted
service "nginx-operator-controller-manager-metrics-service" deleted
deployment.apps "nginx-operator-controller-manager" deleted
```

3. 删除 CRD docker build 过程中的临时镜像文件。

CRD docker build 使用二阶段构建，执行编译的镜像文件有 `1.84GB`。并且多次构建，会产生多个临时镜像文件。

```text
$ docker images | grep none | grep -v k8s | awk '{print $3}' | xargs docker rmi
```

