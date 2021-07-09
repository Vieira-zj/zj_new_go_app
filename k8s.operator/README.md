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

### Define CR and CRD

1. Define CR by update `api/v1alpha1/nginx_types.go`.

```golang
type NginxSpec struct {}
type NginxStatus struct {}
```

2. 执行下面命令生成 CR 资源相关代码 `api/v1alpha1/zz_generated.deepcopy.go`

```text
$ make generate
```

注：每次修改 CR 的定义后都需执行该命令。

3. 执行下面的命令生成 CRD 定义文件 `nginx-operator/config/crd/bases/proxy.example.com_nginxes.yaml`

```text
$ make manifests
```

4. 实现控制器处理逻辑 `controllers/nginx_controller.go`

```golang
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// ...
}
```

### Build and Deploy CRD

1. Build CRD image from `Dockerfile`.

```text
# connect to minikube docker socket
$ eval $(minikube -p minikube docker-env)

# build crd
$ make docker-build
Step 9/14 : RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go
Successfully built 32ec2c7963fc
Successfully tagged controller:latest

$ docker images
controller:latest
```

Custom image tag name in `Makefile`:

```text
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)
```

2. Deploy CRD in k8s.

```text
$ make deploy
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

3. Check CRD deploy.

```text
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
Failed to pull image "controller:latest": rpc error: code = Unknown desc = Error response from daemon: pull access denied for controller, repository does not exist or may require 'docker login': denied: requested access to the resource is denied
```

Update deploy config `imagePullPolicy: IfNotPresent`:

```text
$ kubectl edit deployment/nginx-operator-controller-manager -n nginx-operator-system
```

4. Check logs of nginx-operator-controller-manager pod.

```text
$ kubectl logs nginx-operator-controller-manager-b8f479c94-s7x9z -c manager -n nginx-operator-system
...
Starting Controller     {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx"}
Starting workers        {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "worker count": 1}
```

### Deploy a CRD instance

1. Create deploy yaml `config/samples/proxy_v1alpha1_nginx.yaml`

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

2. Apply a CRD instance.

```text
$ kubectl apply -f config/samples/proxy_v1alpha1_nginx.yaml
nginx.proxy.example.com/nginx-app created

$ kubectl get nginx/nginx-app
NAME        AGE
nginx-app   24m
```

3. Check k8s components of CRD instance.

```text
$ kubectl get all -n default
```

If auth error:

```text
$ kubectl logs nginx-operator-controller-manager-b8f479c94-nqsv2 -c manager -n nginx-operator-system
failed to list *v1.Deployment: deployments.apps is forbidden: User "system:serviceaccount:nginx-operator-system:nginx-operator-controller-manager" cannot list resource "deployments" in API group "apps" at the cluster scope
```

Check clusterrole:

```text
$ kubectl describe clusterrolebinding/nginx-operator-manager-rolebinding -n nginx-operator-system
Role:
  Kind:  ClusterRole
  Name:  nginx-operator-manager-role
Subjects:
  Kind            Name                               Namespace
  ----            ----                               ---------
  ServiceAccount  nginx-operator-controller-manager  nginx-operator-system

$ kubectl describe clusterrole/nginx-operator-manager-role
$ kubectl edit clusterrole/nginx-operator-manager-role
```

Add below auth to clusterrole:

```text
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

Check nginx-operator-controller-manager pod logs.

```text
$ kubectl logs nginx-operator-controller-manager-b8f479c94-2ttm6 -c manager -n nginx-operator-system
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Create CRD Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
```

4. Check created k8s components of CRD instance.

```text
$ kubectl get all -n default

$ kubectl describe service/nginx-app
Port:                     <unset>  80/TCP
TargetPort:               80/TCP
NodePort:                 <unset>  30002/TCP
Endpoints:                172.17.0.4:80
```

5. Access nginx `http://192.168.99.103:30002/`

Note: here use minikube vm ip instead of `localhost`.

### Update CRD instance

1. Update `proxy_v1alpha1_nginx.yaml`: `size: 2`, `nodePort: 30009`

```text
$ kubectl apply -f config/samples/proxy_v1alpha1_nginx.yaml
nginx.proxy.example.com/nginx-app configured

$ kubectl get pod -n default
pod/nginx-app-6f8b78cb8-8jj2m   1/1     Running   0          5s
pod/nginx-app-6f8b78cb8-tg9n4   1/1     Running   0          5s

$ kubectl describe service/nginx-app
Port:                     <unset>  80/TCP
TargetPort:               80/TCP
NodePort:                 <unset>  30009/TCP
Endpoints:                172.17.0.4:80,172.17.0.6:80
```

2. Check logs.

```text
$ kubectl logs nginx-operator-controller-manager-b8f479c94-s7x9z -c manager -n nginx-operator-system
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Update CRD Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
Reconciling Nginx {"reconciler group": "proxy.example.com", "reconciler kind": "Nginx", "name": "nginx-app", "namespace": "default"}
```

### Clear CRD

1. Clear k8s components of CRD instance.

```text
$ kubectl delete -f config/samples/proxy_v1alpha1_nginx.yaml
```

2. Remove CRD related k8s components.

```text
$ make undeploy
```

