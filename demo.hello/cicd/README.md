# readme

- http 长连接
- sharding map

## API test

```sh
# test
curl http://localhost:8081/
curl http://localhost:8081/ping

# put, get
curl http://localhost:8081/store?releaseCycle=2021.04.v3+-+AirPay
curl http://localhost:8081/store?releaseCycle=2021.04.v3+-+AirPay&forceUpdate=true
curl http://localhost:8081/get?storeKey=2021.04.v3+-+AirPay

curl http://localhost:8081/store?fixVersion=apa_v1.0.20.20210426
curl http://localhost:8081/get/issue?storeKey=apa_v1.0.20.20210426

curl http://localhost:8081/store?query=key+in+%28AIRPAY-46283%2CSPPAY-196%29
curl http://localhost:8081/get?storeKey=key+in+%28AIRPAY-46283%2CSPPAY-196%29

# usage
curl http://localhost:8081/store/usage?storeKey=apa_v1.0.20.20210426
```

