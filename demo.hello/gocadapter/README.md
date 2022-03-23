# Goc Adapter

## Overview

1. 同步服务覆盖率 raw 数据
  - 获取指定服务的覆盖率 raw 数据
  - 一个服务的一个 commit 下，保存 5 份覆盖率历史数据，1 份最新的覆盖率结果

2. 定时清除已下线服务
  - 从 goc server list 中删除已下线的服务（pod）
  - 根据 pod name 或 ip, 通过调用 pod monitor 服务接口查询服务状态

3. 获取服务覆盖率结果
  - 同步服务覆盖率 raw 数据
  - 根据 goc server list 中注册的信息，获取 git branch 和 commit 信息
  - 本地 git repo checkout 代码
  - 生成覆盖率结果（全量/增量）
  - 提供接口，根据 server name 来获取覆盖率结果（func/html）
  - 如果服务下线，则使用历史覆盖率数据

4. 定时拉取 goc server list 中服务的覆盖率数据
  - 检查服务是否下线，如果已下线，则该服务从 goc server list 中删除
  - 同步服务覆盖率 raw 数据
  - 定时时间 + 随机时间 防止同一时间有多个覆盖率任务在执行
  - 如果覆盖率数据没有更新，则基于退让机制拉取覆盖率数据
    - 使用 md5sum 判断覆盖率数据是否变化
    - 退让机制：15min -> 60min -> 2hours -> 4hours

5. 生成增量覆盖率报告
  - 工具 diff_cover

问题：

1. 实时拉取覆盖率报告性能问题。

## Goc Adapter Image

1. Use image `bitnami/git:2.35.1`
2. Add `id_rsa` to `~/.ssh`
  - Init ssh connection (add git server fingerprint to know hosts)
3. Add golang runtime and set `PATH`

### Goc Adapter Work 目录结构

设置变量 `GOC_ADAPTER_HOME`，目录结构如下：

```text
- goc_adapter/
  - module_x/
    - repo/
    - cov_history/
      - module_x_1.cov
      - module_x_2.cov
    - module_x.cov
    - module_x_report.txt
    - module_x_report.html
  - module_y/
    - repo/
    - cov_history/
    - module_y.cov
    - module_y_report.txt
    - module_y_report.html
```

## Local Test Env

1. Build and start goc server

```sh
# build
cd goc/; go build .
# run
goc server
```

2. Build and tart echoserver

```sh
# build
cd echo/; goc build . -o goc_echoserver
# run
ENV=staging APPTYPE=apa REGION=th GIT_BRANCH=origin/master GIT_COMMIT=518e0a570c ./goc_echoserver
```

3. Check register service

```sh
goc list
# {
#  "staging_th_apa_goc_echoserver_master_518e0a570c":["http://127.0.0.1:49970","http://127.0.0.1:51007"],
#  "staging_th_apa_goc_echoserver_v2_master_518e0a570c":["http://127.0.0.1:51025"]
# }
```

### Goc API

- goc list

```sh
# list service
curl http://localhost:7777/v1/cover/list
# register service in goc center
curl -XPOST "http://localhost:7777//v1/cover/register?name=staging_th_apa_goc_echoserver_v1&address=http://127.0.0.1:49971"
# remove service from goc center
curl -XPOST http://localhost:7777/v1/cover/remove -H "Content-Type:application/json" -d '{"service":["staging_th_apa_goc_echoserver_v1"]}'

# attach server api
curl http://127.0.0.1:51025/v1/cover/coverage
```

## Goc Adapter API

TODO:

