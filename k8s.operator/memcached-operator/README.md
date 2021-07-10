# Memcached Operator

## Operator Demo

> Refer: <https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/>

1. create a new project

```sh
mkdir -p $HOME/projects/memcached-operator
cd $HOME/projects/memcached-operator
operator-sdk init --domain example.com --repo github.com/example/memcached-operator

operator-sdk create api --group cache --version v1alpha1 --kind Memcached --resource --controller
```

2. implement API and Controller

- `api/v1alpha1/memcached_types.go`
- `controllers/memcached_controller.go`

```sh
make generate
make manifests
```

3. deploy and run Operator

```sh
make docker-build IMG=example.docker.io/memcached-operator:v0.1
make deploy IMG=example.docker.io/memcached-operator:v0.1
```

4. create a Memcached CR

- `config/samples/cache_v1alpha1_memcached.yaml`

```yaml
apiVersion: cache.example.com/v1alpha1
kind: Memcached
metadata:
  name: memcached-sample
spec:
  size: 3
```

apply:

```sh
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml
kubectl get deployment,pod
kubectl get memcached/memcached-sample -o yaml
```

update:

```sh
kubectl patch memcached memcached-sample -p '{"spec":{"size": 5}}' --type=merge
```

## Webhook Demo

1. init project

```sh
cd $GOPATH/src/github.com/example/memcached-operator
operator-sdk create webhook --group cache --version v1alpha1 --kind Memcached --programmatic-validation
```

2. implement validating webhook

- `api/v1beta1/memcached_webhook.go`

```text
admissionReviewVersions={v1alpha1,v1beta1}

func (r *Memcached) ValidateCreate() error {}
func (r *Memcached) ValidateUpdate(old runtime.Object) error {}
func (r *Memcached) ValidateDelete() error {}
```

3. update the generated code and regenerate the CRD manifests

```sh
make generate
make manifests
```

4. enable the webhook and certificate manager in manifests (去掉注释)

- `config/crd/kustomization.yaml`

```yaml
#- patches/webhook_in_memcacheds.yaml
#- patches/cainjection_in_memcacheds.yaml
```

- `config/default/kustomization.yaml`

```yaml
#- ../webhook
#- ../certmanager
#- manager_webhook_patch.yaml
#- webhookcainjection_patch.yaml

#- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR
#  objref:
#    kind: Certificate
#    group: cert-manager.io
#    version: v1
#    name: serving-cert # this name should match the one in certificate.yaml
#  fieldref:
#    fieldpath: metadata.namespace
#- name: CERTIFICATE_NAME
#  objref:
#    kind: Certificate
#    group: cert-manager.io
#    version: v1
#    name: serving-cert # this name should match the one in certificate.yaml
#- name: SERVICE_NAMESPACE # namespace of the service
#  objref:
#    kind: Service
#    version: v1
#    name: webhook-service
#  fieldref:
#    fieldpath: metadata.namespace
#- name: SERVICE_NAME
#  objref:
#    kind: Service
#    version: v1
#    name: webhook-service
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

5. build and deploy the operator

```sh
make docker-build docker-push IMG=example.docker.io/memcached-operator:v0.0.1
make deploy IMG=example.docker.io/memcached-operator:v0.0.1
```

6. create CR sample

- `config/samples/cache_v1beta1_memcached.yaml`

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: memcached-sample
---
apiVersion: cache.example.com/v1beta1
kind: Memcached
metadata:
  name: memcached-sample
  namespace: memcached-sample
spec:
  replicaSize: "3.2"
```

7. the request will be rejected, and check logs

```sh
kubectl apply -f config/samples/cache_v1beta1_memcached.yaml
kubectl logs -f deploy/memcached-operator-controller-manager -n memcached-operator-system -c manager
```

## Cleanup

```sh
kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml
make undeploy
```

