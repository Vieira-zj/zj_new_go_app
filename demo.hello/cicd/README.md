# readme

- http 长连接
- sharding map

## API test

```sh
curl http://localhost:8081/
curl http://localhost:8081/ping

# store
curl http://localhost:8081/store?releaseCycle=2021.04.v3%20-%20AirPay
curl http://localhost:8081/store?releaseCycle=2021.04.v3%20-%20AirPay&forceUpdate=true

curl http://localhost:8081/store/usage?storeKey=2021.04.v3%20-%20AirPay

# get
curl http://localhost:8081/get/issue?storeKey=2021.04.v3%20-%20AirPay&key=AIRPAY-44435
curl http://localhost:8081/get?storeKey=2021.04.v3%20-%20AirPay
```

