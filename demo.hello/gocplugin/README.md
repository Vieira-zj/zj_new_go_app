# Goc Plugin

## Overview

- Goc Plugin 包括 Goc Report, Goc Watch Dog 和 Goc Portal 三部分
- 基于工具 Goc + diff_cover (二次开发)

### Goc Report

1. 从 goc server 获取指定服务的覆盖率数据
  - 一个服务的一个 commit 下，可能会包括 N 份覆盖率数据（1份最新数据+历史数据）
  - 如果服务已下线（异常退出），则从 goc watch dog 获取覆盖率数据
  - 支持服务多个副本的覆盖率数据合并

2. 生成覆盖率报告
  - func/html 覆盖率报告
  - 增量代码覆盖率报告

问题：

1. 拉取和生成覆盖率报告性能问题
  - 异步执行，指定时间内完成则返回结果，否则返回超时
  - 覆盖率结果缓存 N 分钟

### Goc Watch Dog

1. 定时从 goc server 中删除已下线的服务（pod）
  - 检查 goc attached server 端口（代替从 pod monitor 服务接口查询 pod 状态）

2. 定时拉取 goc server 中服务覆盖率数据

3. 服务异常退出前，拉取服务覆盖率数据
  - 提供回调接口，在 pod 中设置 pre-stop webhook

### Goc Portal

服务覆盖率结果展示：

- Cover total
- Func cover report
- Html cover report
- 增量覆盖率报告
- 历史覆盖率数据趋势图

操作：

- Refresh
- Force refresh
- Clear cover

### Todos

1. 展示基于 commit 和 branch 比较的增量覆盖率结果
2. 展示基于 commit 的历史覆盖率结果
3. 支持单测覆盖率结果上报，及展示

## Goc Plugin Design

### 服务架构

![img](static/framework.drawio.svg)

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

- `goc_o_staging_service_cover`

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

3. Check register services

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

- List services

```sh
curl http://localhost:7777/v1/cover/list | jq .
```

- Register service in goc center

```sh
curl -XPOST "http://localhost:7777//v1/cover/register?name=staging_th_apa_goc_echoserver_v1&address=http://127.0.0.1:49971"
```

- Remove service from goc center

```sh
curl -XPOST http://localhost:7777/v1/cover/remove -H "Content-Type:application/json" \
  -d '{"service":["staging_th_apa_goc_echoserver_v1"]}'
```

- Attached server api

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
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a"}' | jq .

# force sync
curl -XPOST "http://127.0.0.1:8089/cover/report/sync" -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a", "is_force":true}' | jq .
```

- `/cover/raw`: get service cover raw data.

```sh
curl -XPOST http://127.0.0.1:8089/cover/raw -H "Content-Type:application/json" \
  -d '{"srv_addr":"http://127.0.0.1:51007"}' -o 'raw_profile.cov'
```

- `/cover/report/latest`: get latest cover func/html report.

```sh
# download func report
curl -XPOST http://127.0.0.1:8089/cover/report/latest -H "Content-Type:application/json" \
  -d '{"rpt_type":"func", "srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a"}' -o 'cover_report.func'

# download html report
curl -XPOST http://127.0.0.1:8089/cover/report/latest -H "Content-Type:application/json" \
  -d '{"rpt_type":"html", "srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a"}' -o 'cover_report.html'
```

- `/cover/report/history`: get latest cover func/html report.

## Goc Watch Dog API

- `/watcher/cover/list`: list saved service cov file.

```sh
curl -XPOST http://127.0.0.1:8089/watcher/cover/list -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a", "limit":3}' | jq .
```

- `/watcher/cover/get`: get service cov file, default latest one.

```sh
# download latest cov file
curl -XPOST http://127.0.0.1:8089/watcher/cover/get -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a"}' -o profile.cov | jq .

# download cov file
cov_file="staging_th_apa_goc_echoserver_master_b63d82705a_20220422_150331.cov"
curl -XPOST http://127.0.0.1:8089/watcher/cover/get -H "Content-Type:application/json" \
  -d "{\"srv_name\":\"staging_th_apa_goc_echoserver_master_b63d82705a\", \"cov_file_name\":\"${cov_file}\"}" -o ${cov_file} | jq .
```

- `/watcher/cover/save`: fetch service cover data and save.

```sh
curl -XPOST http://127.0.0.1:8089/watcher/cover/save -H "Content-Type:application/json" \
  -d '{"srv_name":"staging_th_apa_goc_echoserver_master_b63d82705a", "addresses":["http://127.0.0.1:62542"]}' | jq .
```

