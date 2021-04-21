# readme

- http 长连接
- sharding map

## Rest APIs

```sh
# test
curl http://localhost:8081/
curl http://localhost:8081/ping

# put, get
curl http://localhost:8081/store -d '{"releaseCycle": "2021.04.v3 - AirPay"}'
curl http://localhost:8081/store -d '{"releaseCycle": "2021.04.v3 - AirPay", "forceUpdate": true}'
curl http://localhost:8081/get -d '{"storeKey": "2021.04.v3 - AirPay"'

curl http://localhost:8081/store -d '{"fixVersion": "apa_v1.0.20.20210426"}'
curl http://localhost:8081/get -d '{"storeKey": "apa_v1.0.20.20210426"'
curl http://localhost:8081/get/issue -d '{"storeKey": "apa_v1.0.20.20210426", "issueKey": "SPPAY-675"}'

curl http://localhost:8081/store -d '{"query": "key in (AIRPAY-46283,SPPAY-196)"}'
curl http://localhost:8081/get -d '{"storeKey": "key in (AIRPAY-46283,SPPAY-196)"}'

# usage
curl http://localhost:8081/store/usage -d '{"storeKey": "key in (AIRPAY-46283,SPPAY-196)"}'
```

