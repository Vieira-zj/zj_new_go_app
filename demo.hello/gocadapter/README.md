# Goc Adapter

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

## Goc Adapter Work 目录结构

设置变量 `GOC_ADAPTER_HOME`，目录结构如下：

```text
- goc_adapter/
  - module_x/
    - repo/
    - sqlite.db
    - cov_data/
      - module_x_1.cov
      - module_x_1_report.txt
      - module_x_1_report.html
      - module_x_2.cov
      - module_x_2_report.txt
      - module_x_2_report.html
  - module_y/
    - repo/
    - sqlite.db
    - cov_data/
      - module_y_1.cov
      - module_y_1_report.txt
      - module_y_1_report.html
```

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
curl http://localhost:7777/v1/cover/list
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

## Goc Adapter API

Test goc adapter server:

```sh
curl -i http://127.0.0.1:8089/
curl http://127.0.0.1:8089/ping | jq .
```

- `/cover/raw`: get cover raw data

```sh
curl -XPOST http://127.0.0.1:8089/cover/raw -H "Content-Type:application/json" -d '{"srv_addr":"http://127.0.0.1:51007"}' -o 'raw_profile.cov'
```

- `/cover/report/sync`: sync cover results, and generate report

- `cover/report/func`: get cover func report

- `cover/report/html`: get cover html report

- `/cover/total`: get cover total

- `/cover/historytotals`: get history cover total

TODO:

