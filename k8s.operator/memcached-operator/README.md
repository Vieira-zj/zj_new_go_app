# Memcached Operator

## Operator Demo

> Refer: <https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/>

1. Create a new project

```sh
mkdir -p $HOME/projects/memcached-operator
cd $HOME/projects/memcached-operator
operator-sdk init --domain example.com --repo github.com/example/memcached-operator
operator-sdk create api --group cache --version v1alpha1 --kind Memcached --resource --controller
```

Update namespaces in `ctrl.Options{}` of `main.go` to list-watch.

2. Implement API and Controller

- crd: `api/v1alpha1/memcached_types.go`
- controller: `controllers/memcached_controller.go`

Generate code and CRD manifests.

```sh
make generate
make manifests
```

3. Build and deploy Operator

```sh
make docker-build IMG=example.docker.io/memcached-operator:v0.1
make deploy IMG=example.docker.io/memcached-operator:v0.1
```

Check memcached-operator components:

```text
kubectl get svc,pod -n memcached-operator-system
NAME                                                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/memcached-operator-controller-manager-metrics-service   ClusterIP   10.100.146.46   <none>        8443/TCP   54s
service/memcached-operator-webhook-service                      ClusterIP   10.102.52.26    <none>        443/TCP    54s
NAME                                                         READY   STATUS    RESTARTS   AGE
pod/memcached-operator-controller-manager-784d74869d-blp4b   2/2     Running   0          54s
```

4. Create a Memcached CR

- `config/samples/cache_v1alpha1_memcached.yaml`

```yaml
apiVersion: cache.example.com/v1alpha1
kind: Memcached
metadata:
  name: memcached-sample
spec:
  size: 3
```

Apply memcached-sample manifest:

```text
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml

kubectl get memcached/memcached-sample -o yaml
spec:
  size: 3
status:
  nodes:
  - memcached-sample-6c765df685-hxk6h
  - memcached-sample-6c765df685-vmklf
  - memcached-sample-6c765df685-j97cj

kubectl get deploy,pod
```

Check logs of operator-controller-manager pod:

```sh
kubectl get pod -n memcached-operator-system
kubectl logs memcached-operator-controller-manager-784d74869d-bllcm -n memcached-operator-system -c manager
```

5. Update size by `kubectl patch`

```text
kubectl patch memcached/memcached-sample -p '{"spec":{"size": 1}}' --type=merge

kubectl get memcached/memcached-sample -o yaml
spec:
  size: 1
status:
  nodes:
  - memcached-sample-6c765df685-j97cj
```

## Webhook Demo

> Refer: <https://vincenthou.medium.com/how-to-create-validating-webhook-with-operator-sdk-73f9c6332609>

1. Init project

```sh
cd $GOPATH/src/github.com/example/memcached-operator
operator-sdk create webhook --group cache --version v1alpha1 --kind Memcached --programmatic-validation
```

2. Implement webhooks

- `api/v1beta1/memcached_webhook.go`

```text
admissionReviewVersions={v1alpha1,v1beta1}

func (r *Memcached) Default() {}
func (r *Memcached) ValidateCreate() error {}
func (r *Memcached) ValidateUpdate(old runtime.Object) error {}
func (r *Memcached) ValidateDelete() error {}
```

3. Update the generated code and regenerate the CRD manifests

```sh
make generate
make manifests
```

4. Enable the webhook and certificate manager in manifests (去掉注释)

- `config/crd/kustomization.yaml`

```yaml
- patches/webhook_in_memcacheds.yaml
- patches/cainjection_in_memcacheds.yaml
```

- `config/default/kustomization.yaml`

```yaml
- ../webhook
- ../certmanager

- manager_webhook_patch.yaml
- webhookcainjection_patch.yaml

- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: serving-cert # this name should match the one in certificate.yaml
  fieldref:
    fieldpath: metadata.namespace
- name: CERTIFICATE_NAME
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: serving-cert # this name should match the one in certificate.yaml
- name: SERVICE_NAMESPACE # namespace of the service
  objref:
    kind: Service
    version: v1
    name: webhook-service
  fieldref:
    fieldpath: metadata.namespace
- name: SERVICE_NAME
  objref:
    kind: Service
    version: v1
    name: webhook-service
```

- `config/default/webhookcainjection_patch.yaml`

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
  annotations:
    cert-manager.io/inject-ca-from: $(CERTIFICATE_NAMESPACE)/$(CERTIFICATE_NAME)
```

5. Build and deploy the operator

```sh
make docker-build IMG=example.docker.io/memcached-operator:v0.0.1
make deploy IMG=example.docker.io/memcached-operator:v0.0.1
```

6. If below `cert-manager.io` error when deploy operator

```text
unable to recognize "STDIN": no matches for kind "Certificate" in version "cert-manager.io/v1"
unable to recognize "STDIN": no matches for kind "Issuer" in version "cert-manager.io/v1"
```

Install cert-manager in k8s:

> <https://cert-manager.io/docs/installation/kubernetes/>

```text
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.4.0/cert-manager.yaml

kubectl get pods --namespace cert-manager
cert-manager-5d7f97b46d-47mmj              1/1     Running   0          3h12m
cert-manager-cainjector-69d885bf55-vwj69   1/1     Running   0          3h12m
cert-manager-webhook-8d7495f4-87gt7        1/1     Running   0          3h12m
```

Then re-deploy operator:

```sh
make deploy IMG=example.docker.io/memcached-operator:v0.1
```

7. Create CR sample and deploy

- `config/samples/cache_v1beta1_memcached.yaml`

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: memcached-sample
---
apiVersion: cache.example.com/v1alpha1
kind: Memcached
metadata:
  name: memcached-sample
  namespace: memcached-sample
spec:
  size: 1
```

8. If `size = 2`, the request will be rejected by validating webhook

```text
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml
Error from server (Cluster size must be an odd number): error when creating "config/samples/cache_v1alpha1_memcached.yaml": admission webhook "vmemcached.kb.io" denied the request: Cluster size must be an odd number
```

Check memcached-operator-controller-manager logs:

```text
kubectl logs -f memcached-operator-controller-manager-784d74869d-blp4b -n memcached-operator-system -c manager
validate create {"name": "memcached-sample"}
wrote response  {"webhook": "/validate-cache-example-com-v1alpha1-memcached", "code": 403, "reason": "Cluster size must be an odd number", "UID": "0e6aba63-0174-428d-b15a-e9d59a9d4284", "allowed": false}
```

9. If `size = 0`, default create 3 instances by mutating webhook

```text
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml

kubectl get memcached/memcached-sample -o yaml -n memcached-sample
spec:
  size: 3
status:
  nodes:
  - memcached-sample-6c765df685-mb5nj
  - memcached-sample-6c765df685-q7sjf
  - memcached-sample-6c765df685-7w65w
```

Check memcached-operator-controller-manager logs (mutating then validating):

```text
default {"name": "memcached-sample"}
wrote response  {"webhook": "/mutate-cache-example-com-v1alpha1-memcached", "code": 200, "reason": "", "UID": "023a6f1e-934a-4c47-8d49-e57ef34772d3", "allowed": true}
validate create {"name": "memcached-sample"}
wrote response  {"webhook": "/validate-cache-example-com-v1alpha1-memcached", "code": 200, "reason": "", "UID": "633e76af-fd44-4733-9495-488271bc8491", "allowed": true}
```

## Cleanup

```sh
kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml
make undeploy
```

