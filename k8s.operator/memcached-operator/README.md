# Memcached Operator Demo

> Refer: <https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/>

## Create a new project

```sh
mkdir -p $HOME/projects/memcached-operator
cd $HOME/projects/memcached-operator
operator-sdk init --domain example.com --repo github.com/example/memcached-operator

operator-sdk create api --group cache --version v1alpha1 --kind Memcached --resource --controller
```

## Implement API and Controller

- `api/v1alpha1/memcached_types.go`
- `controllers/memcached_controller.go`

```sh
make generate
make manifests
```

## Deploy and Run Operator

```sh
make docker-build IMG=cache.example.com/memcached-operator:v0.1
make docker-push
make deploy
```

## Create a Memcached CR

`config/samples/cache_v1alpha1_memcached.yaml`

```yaml
apiVersion: cache.example.com/v1alpha1
kind: Memcached
metadata:
  name: memcached-sample
spec:
  size: 3
```

Apply:

```sh
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml
kubectl get deployment,pod
kubectl get memcached/memcached-sample -o yaml
```

Update:

```sh
kubectl patch memcached memcached-sample -p '{"spec":{"size": 5}}' --type=merge
```

## Cleanup

```sh
kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml
make undeploy
```

