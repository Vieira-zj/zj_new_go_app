# Goc Watch Dog

## Overview

1. 获取指定服务的覆盖率 raw 数据
  - 一个服务的一个 commit 下，可能会包括 N 份覆盖率数据（1份最新数据+历史数据）

2. 生成服务覆盖率报告
  - 生成 func/html 覆盖率报告
  - 如果服务已下线，则使用历史覆盖率结果
  - 生成增量代码覆盖率报告

3. 定时清除已下线服务
  - 从 goc server list 中删除已下线的服务（pod）
  - 检查 goc attached server 端口（代替从 pod monitor 服务接口查询 pod 状态）

4. 定时拉取 goc server list 中服务的覆盖率数据
  - 定时 + 随机时间 防止同一时间有多个覆盖率任务在执行
  - 如果覆盖率数据没有更新，则基于退让机制拉取覆盖率数据
    - 退让机制：15min -> 60min -> 2hours -> 4hours

5. 服务异常退出前，拉取覆盖率 raw 数据
  - 通过设置 pod pre-stop webhook

注意：

1. 实时拉取覆盖率数据性能问题。

## Goc Watch Dog Work 目录

目录结构如下：

```text
- goc_watch_dog_root/
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

## DB

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
ENV=staging APPTYPE=apa REGION=th GIT_BRANCH=origin/master GIT_COMMIT=845820727e ./goc_echoserver
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
curl -XPOST http://localhost:7777/v1/cover/remove -H "Content-Type:application/json" -d '{"service":["staging_th_apa_goc_echoserver_v1"]}'
```

- attach server api

```sh
curl http://127.0.0.1:51025/v1/cover/coverage
```

## Goc Watch Dog API

Test goc watch dog server:

```sh
curl -i http://127.0.0.1:8089/
curl http://127.0.0.1:8089/ping | jq .
```

- `/cover/list`: get list of services cover info.

```sh
curl http://127.0.0.1:8089/cover/list | jq .
```

- `/cover/raw`: get services cover raw data.

```sh
curl -XPOST http://127.0.0.1:8089/cover/raw -H "Content-Type:application/json" \
  -d '{"srv_addr":"http://127.0.0.1:51007"}' -o 'raw_profile.cov'
```

- `/cover/report/sync`: sync cover results, and generate report.

```sh
curl -XPOST http://127.0.0.1:8089/cover/report/sync -H "Content-Type:application/json" \
  -d '{"srv_name": "staging_th_apa_goc_echoserver_master_845820727e", "addresses": ["http://127.0.0.1:51007"]}'
```

- `cover/latest/report`: get latest cover func/html report.

```sh
curl -XPOST http://127.0.0.1:8089/cover/latest/report -H "Content-Type:application/json" \
  -d '{"rpt_type":"func", "srv_name":"staging_th_apa_goc_echoserver_master_845820727e"}' -o 'cover_report.func'
```

- `cover/history/report`: get latest cover func/html report.

- `/cover/latest/total`: get service latest cover total.

- `/cover/history/total`: get service history cover totals.

TODO:

