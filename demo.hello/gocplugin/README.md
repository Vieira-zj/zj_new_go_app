# Goc Plugin

## Overview

- 基于 Goc + diff_cover (二次开发)
- Goc Plugin 包括 Goc Report, Goc Watch Dog 和 Goc Portal 三部分

### Goc Report

1. 获取指定服务的覆盖率 raw 数据
  - 一个服务的一个 commit 下，可能会包括 N 份覆盖率数据（1份最新数据+历史数据）
  - 如果服务已下线（异常退出），则从 goc watch dog 中获取覆盖率数据

2. 生成服务覆盖率报告
  - 生成 func/html 覆盖率报告
  - 生成增量代码覆盖率报告

3. 定时拉取 goc server list 中服务的覆盖率数据
  - 定时 + 随机时间 防止同一时间有多个覆盖率任务在执行
  - 如果覆盖率数据没有更新，则基于退让机制拉取覆盖率数据
    - 退让机制：20min -> 60min -> 2hours -> 4hours

问题：

1. 实时拉取和生成覆盖率结果性能问题。
  - 同步执行，前端需要处理超时的情况
  - 覆盖率结果缓存 N 分钟，防止频繁执行

### Goc Watch Dog

1. 定时从 goc server list 中删除已下线的服务（pod）
  - 检查 goc attached server 端口（代替从 pod monitor 服务接口查询 pod 状态）

2. 服务异常退出前，拉取覆盖率 raw 数据
  - 通过设置 pod pre-stop webhook

### Goc Portal

服务覆盖率结果展示：

- 服务正常，获取最新的覆盖率结果
- 服务下线，获取服务异常退出前的覆盖率结果
- 服务下线，没有获取到当前 commit 的覆盖率结果

### Dev Plan

1. 获取和展示服务最新的覆盖率结果，包括 func/html 报告，cover total 趋势图
2. 服务有多个副本时，覆盖率结果合并
3. 展示历史覆盖率结果（基于 commit）
4. 展示基于 commit 和 branch 比较的增量覆盖率结果
5. 支持单测覆盖率结果上报，及展示

## Goc Plugin Design

### 服务部署

// TODO:

### Goc Report 工作目录

工作目录（root dir）结构如下：

```text
- goc_report_root/
  - sqlite.db
  - module_x/
    - repo/
    - cov_data/
      - module_x_1.cov
      - module_x_1_report.txt
      - module_x_1_report.html
      - module_x_2.cov
      - module_x_2_report.txt
      - module_x_2_report.html
  - module_y/
    - repo/
    - cov_data/
      - module_y_1.cov
      - module_y_1_report.txt
      - module_y_1_report.html
```

### Goc Report Table

- `goc_o_staging_service_cover`: 服务覆盖率数据。

字段包括：`id,created_at,updated_at,deleted_at,env,region,app_name,addresses,git_branch,git_commit,is_latest,cov_file_path,cover_total`

使用 key `env + region + app_name + git_commit` 标识一个服务的覆盖率数据，有多条数据，但 `is_latest=y` 的数据只会有1条。

## Local Test

1. Build and start goc server

```sh
# build
cd goc/; go build .
# run
goc server
```

2. Build and start echoserver

```sh
# build
cd echo/; goc build . -o goc_echoserver
# run
ENV=staging APPTYPE=apa REGION=th GIT_BRANCH=origin/master GIT_COMMIT=b63d82705a ./goc_echoserver -p 8081
```

3. Check register service

```sh
goc list
# {
#  "staging_th_apa_goc_echoserver_master_518e0a570c":["http://127.0.0.1:49970","http://127.0.0.1:51007"],
#  "staging_th_apa_goc_echoserver_v2_master_518e0a570c":["http://127.0.0.1:51025"]
# }
```

4. Test echoserver api

```sh
curl -i http://localhost:8081/
curl -i http://localhost:8081/ping
curl -XPOST "http://localhost:8081/mirror?name=foo" -H "X-Test:Mirror" -d 'hello' | jq .
```

## Goc API

- list service

```sh
curl http://localhost:7777/v1/cover/list | jq .
```

- register service in goc center

```sh
curl -XPOST "http://localhost:7777//v1/cover/register?name=staging_th_apa_goc_echoserver_v1&address=http://127.0.0.1:49971"
```

- remove service from goc center

```sh
curl -XPOST http://localhost:7777/v1/cover/remove -H "Content-Type:application/json" \
  -d '{"service":["staging_th_apa_goc_echoserver_v1"]}'
```

- attach server api

```sh
curl http://127.0.0.1:51025/v1/cover/coverage
```

## Goc Report API

Server health check:

```sh
curl -i http://127.0.0.1:8089/
curl http://127.0.0.1:8089/ping | jq .
```

------

Get service cover data:

- `/cover/list`: get list of services cover info.

```sh
curl http://127.0.0.1:8089/cover/list | jq .
```

- `/cover/total/latest`: get latest service cover total.

```sh
curl -XPOST http://127.0.0.1:8089/cover/total/latest -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a"}' | jq .
```

- `/cover/total/history`: get history service cover totals.

```sh
curl -XPOST http://127.0.0.1:8089/cover/total/history -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a"}' | jq .
```

------

Sync service cover data and create report:

- `/cover/report/sync`: sync service cover results, generate report, and returns cover total.

```sh
curl -XPOST http://127.0.0.1:8089/cover/report/sync -H "Content-Type:application/json" \
  -d '{"srv_name": "staging_th_apa_goc_echoserver_master_b63d82705a"}' | jq .

# force sync
curl -XPOST "http://127.0.0.1:8089/cover/report/sync?force=true" -H "Content-Type:application/json" \
  -d '{"srv_name": "staging_th_apa_goc_echoserver_master_b63d82705a"}' | jq .
```

- `/cover/raw`: get service cover raw data.

```sh
curl -XPOST http://127.0.0.1:8089/cover/raw -H "Content-Type:application/json" \
  -d '{"srv_addr":"http://127.0.0.1:51007"}' -o 'raw_profile.cov'
```

- `/cover/report/latest`: get latest cover func/html report.

```sh
# func report
curl -XPOST http://127.0.0.1:8089/cover/report/latest -H "Content-Type:application/json" \
  -d '{"rpt_type":"func", "srv_name":"staging_th_apa_goc_echoserver_master_845820727e"}' -o 'cover_report.func'

# html report
curl -XPOST http://127.0.0.1:8089/cover/report/latest -H "Content-Type:application/json" \
  -d '{"rpt_type":"html", "srv_name":"staging_th_apa_goc_echoserver_master_845820727e"}' -o 'cover_report.html'
```

- `/cover/report/history`: get latest cover func/html report.

## Goc Watch Dog API

- `/watcher/cover/list`: list saved service cov file.

```sh
curl -XPOST http://127.0.0.1:8089/watcher/cover/list -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a", "limit": 3}' | jq .
```

- `/watcher/cover/get`: get service cov file, default latest one.

```sh
cov_file="staging_th_apa_goc_echoserver_master_b63d82705a_20220421_183420_0.cov"
curl -XPOST http://127.0.0.1:8089/watcher/cover/get -H "Content-Type:application/json" \
  -d "{\"srv_name\":\"staging_th_apa_goc_echoserver_master_b63d82705a\", \"cov_file_name\":\"${cov_file}\"}" -o ${cov_file} | jq .
```

- `/watcher/cover/save`: fetch service cover data and save.

```sh
curl -XPOST http://127.0.0.1:8089/watcher/cover/save -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a", "addresses": ["http://127.0.0.1:54621"]}' | jq .
```
