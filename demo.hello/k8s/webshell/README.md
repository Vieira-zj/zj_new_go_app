# K8S Webshell Demo

> Refer: https://github.com/maoqide/kubeutil
>

## Test Steps

1. start a debug pod

```sh
kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
```

2. `cd webshell/static`, and open `terminal.html`

3. connect to pod by webshell by selecting `namespace`, `pod` and `container`

