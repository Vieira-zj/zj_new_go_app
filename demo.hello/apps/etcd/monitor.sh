#!/bin/bash
set -e

log_file="/tmp/monitor.log"
target_file="/tmp/restart.flag"

if [[ -f ${log_file} ]]; then
    rm ${log_file}
fi

 while true; do
    echo "$(date) | check for restart.flag" >> ${log_file}
    if [[ -f ${target_file} ]]; then
        # fix issue: pid太长与user显示在一起 137530root
        ps -ef | grep http_proxy | grep -v grep | awk '{print $1}' | tr "root" " " | xargs kill
        addr=$(cat ${target_file})
        echo "restart http proxy with addr ${addr}" | tee ${log_file}
        /tmp/http_proxy_linux -e ${addr} &
        rm ${target_file}
    fi
    sleep 3
done

echo "monitor exit"
