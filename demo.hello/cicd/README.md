# README

> Get jira issues and linked MR data, and store in cache for search.

## Data Struct

Cache存储使用 sharding map + 读写锁，通过设置合理的并发数和分片数，减少锁竞争。

一个map保存issue结构体，另一个map保存MR结构体。

> TODO: Hash函数的随机性，减少热点数据。

## Data Collect

使用 Goroutine + Fix Channel + waitGroup 完成issue和MR数据收集。

|              | Goroutine + Blocked Queue (Channel)                          | Goroutine + Semaphore + waitGroup                            |
| ------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 描述         | 启动固定数量的goroutine, 消费阻塞队列中的任务。              | 每个任务启动一个goroutine来执行，通过 semaphore 来控制goroutine数量。 |
| 网络请求     | 在固定的goroutine中发起网络请求，只需建立一个长连接。        | 每个goroutine中都会建立一个网络连接。                        |
| 任务状态判断 | 不能准确判断任务是否执行完成。<br />需要保存全局cancel函数，通过定时任务来停止已完成的任务，否则goroutine会泄露。 | 通过 waitgroup 可准确判断任务是否执行完成。                  |
| 性能         | 由固定数量的goroutine来分别处理ticket和MR任务。              | Ticket和MR任务共享goroutine来处理。                          |
| 适用场景     | 作为worker持续的处理任务。                                   | 执行短期任务，完成后停止。                                   |

配置项：

- `parallel`：默认设置为10. 可通过调高并发数来实时拉取数据。
- `refreshInterval`：刷新周期。根据每个store中缓存的数据量来定时更新数据。
- `expired`：数据过期时间。（TODO: 淘汰策略）

## Issues Tree Print Data

从map中构造结构化数据给前端展示。

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

