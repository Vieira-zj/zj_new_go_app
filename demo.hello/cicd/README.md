# README

> Get jira issues data and store in cache for search.

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

```sh
# test
curl http://localhost:8081/
curl http://localhost:8081/ping

# put
curl http://localhost:8081/store -d '{"storeKey": "2021.04.v4 - AirPay", "storeKeyType": "ReleaseCycle"}'
curl http://localhost:8081/store -d '{"storeKey": "2021.04.v4 - AirPay", "storeKeyType": "ReleaseCycle", "forceUpdate": true}'

# get by relcycle
curl http://localhost:8081/get -d '{"storeKey": "2021.04.v4 - AirPay", "storeKeyType": "ReleaseCycle"}'
# get by version
curl http://localhost:8081/get -d '{"storeKey": "apa_v1.0.20.20210426", "storeKeyType": "FixVersion"}'
curl http://localhost:8081/get/issue -d '{"storeKey": "apa_v1.0.20.20210426", "storeKeyType": "FixVersion", "issueKey": "AIRPAY-56683"}'
# get by jql
curl http://localhost:8081/get -d '{"storeKey": "key in (AIRPAY-46283,SPPAY-196)", "storeKeyType": "jql"}'

# usage
curl http://localhost:8081/store/usage -d '{"storeKey": "2021.04.v4 - AirPay"}'
```

