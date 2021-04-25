# README

> Get jira issues and linked MR data, and store in cache for search.

## Structs and Cache

> Sharding Map.

## Issues Tree Collect Data

### Channel (Blocked Queue)

TODO:

### Fix Chancel + waitGroup

TODO:

## Issues Tree Print Data

TODO:

## Rest APIs

- Test

```sh
curl http://localhost:8081/
curl http://localhost:8081/ping
```

- Put

```sh
curl http://localhost:8081/store/save -d '{"storeKey": "2021.04.v4 - AirPay", "storeKeyType": "ReleaseCycle"}'
curl http://localhost:8081/store/save -d '{"storeKey": "2021.04.v4 - AirPay", "storeKeyType": "ReleaseCycle", "forceUpdate": true}'

# usage
curl http://localhost:8081/store/usage -d '{"storeKey": "2021.04.v4 - AirPay"}'
```

- Get

```sh
# get by relcycle
curl http://localhost:8081/get/store -d '{"storeKey": "2021.04.v4 - AirPay", "storeKeyType": "ReleaseCycle"}'
curl http://localhost:8081/get/repos -d '{"storeKey": "2021.04.v4 - AirPay"}'

# get by version
curl http://localhost:8081/get/store -d '{"storeKey": "apa_v1.0.20.20210426", "storeKeyType": "FixVersion"}'
curl http://localhost:8081/get/issue -d '{"storeKey": "apa_v1.0.20.20210426", "issueKey": "AIRPAY-56683"}'

# get by jql
curl http://localhost:8081/get/store -d '{"storeKey": "key in (AIRPAY-46283,SPPAY-196)", "storeKeyType": "jql"}'
```

