# K8s Informer

Refer:

- <https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/deployment/deployment_controller.go>
- <https://github.com/resouer/k8s-controller-custom-resource>

## Test

```sh
# start informer
go run main.go -t rs

# create resource
kubectl create -f deploy/busybox_errexit_deploy.yaml
kubectl get all -n k8s-test
# clear
kubectl delete -f deploy/busybox_errexit_deploy.yaml
```

