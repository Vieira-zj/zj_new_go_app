# Goc Watch Dog

## Overview

1. 更新服务覆盖率 raw 数据
  - 获取指定服务的覆盖率 raw 数据
  - 与历史数据合并

2. 定时清除已下线服务
  - 根据 pod name 或 ip, 通过调用 pod monitor 服务接口查询服务状态
  - 从 goc server list 中删除已下线的服务（pod）

3. 获取服务覆盖率结果
  - 根据 goc server list 中注册的信息，获取 git branch 和 commit 信息
  - 更新本地 git repo 代码
  - 更新服务覆盖率 raw 数据，并生成覆盖率结果
  - 根据 server name 来获取覆盖率结果
  - 历史数据
    - 获取服务最近 5 个版本的覆盖率结果
    - 如果服务下线，则使用历史覆盖率数据

4. 定时拉取 goc server list 中服务的覆盖率数据
  - 检查服务是否下线，如果已下线，则该服务从 goc server list 中删除
  - 更新服务覆盖率 raw 数据

5. 生成增量覆盖率报告
  - diff_cover 镜像

## Goc Watcher 目录结构

设置变量 `goc_watcher_home`，目录结果如下：

```text
- goc_watcher/
  - bin/
    - goc
  run.sh
  - goc_coverage/
    - payment/
      - cur_cov.txt
      - history/
        - pre_cov_ts1.txt
        - pre_cov_ts2.txt
      - reports/
        - cov_func_report_ts1.txt
        - cov_func_report_ts2.txt
        - cov_html_report_ts1.txt
        - cov_html_report_ts2.txt
        - cov_history_report.txt
    - wallet/
      - history/
      - reports/
```

## Goc Watcher API

### Env

```sh
```

