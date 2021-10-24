# K8S Pod Monitor

> Monitor pod status by k8s client-go `Informer`. If any invalid pod exist, it will send notification to mm channel by interval.
>

## Pod List APIs

```sh
curl http://localhost:8081/ping

curl http://localhost:8081/resource/pods | jq .
curl http://localhost:8081/list/pods | jq .
```

