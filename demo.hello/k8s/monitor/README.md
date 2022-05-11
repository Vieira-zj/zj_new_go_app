# K8S Pod Monitor

## Overview

1. Monitor pod status by k8s client-go `Informer`. If any invalid pod exist, it will send notification to mm channel by interval.
2. Get specified pod status by name or ip address.

## Deploy in K8s

```sh
kubectl create -f deploy/monitor_rbac.yaml
# deploy_pod_monitor
./run.sh

# check resources
kubectl get sa -n k8s-test
kubectl get all -n k8s-test

# watch logs
kubectl logs -f pod-monitor-68df8b57b5-wg2dw -n k8s-test
```

## Monitor APIs Test

- local

```sh
curl http://localhost:8081/ping

curl http://localhost:8081/resource/pods | jq .
curl http://localhost:8081/list/pods | jq .

# list by filter
curl -XPOST http://localhost:8081/list/pods/filter \
  -d '{"names":["etcd","ingress-nginx"], "ips":["172.17.0.7"]}' | jq .
```

- cluster

```sh
curl http://test.monitor.com/ping

curl http://test.monitor.com/resource/pods | jq .
curl http://test.monitor.com/list/pods | jq .
```

## Notification Test

Test send notification when pod terminate or crash.

```sh
# set notify
curl "http://localhost:8081/notify?open=true"
# check notify setting
curl http://localhost:8081/state | jq .

# start a pod for test
kubectl run debug-pod -it --rm --restart=Never --image=busybox:1.35.0 sh
# test notify by exit pod with error code
root# exit 99

# deploy a error pod for test
kubectl create -f deploy/error_exit_deploy.yaml
```

