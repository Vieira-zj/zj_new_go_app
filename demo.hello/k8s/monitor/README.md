# K8S Pod Monitor

> Monitor pod status by k8s client-go `Informer`. If any invalid pod exist, it will send notification to mm channel by interval.
>

## Deploy in K8s

```sh
kubectl create -f deploy/monitor_rbac.yaml
./run.sh # deploy_pod_monitor

# check resources
kubectl get sa -n k8s-test
kubectl get all -n k8s-test

# watch logs
kubectl logs -f pod-monitor-68df8b57b5-wg2dw -n k8s-test

# start a error pod for test
kubectl create -f deploy/error_exit_deploy.yaml
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

