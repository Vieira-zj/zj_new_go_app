#!/bin/bash
set -eu

#
# Http Server Topic: httpserver_important, httpserver_access_log
#
# Command Line:
# 
# broker config info:
# kafka-configs.sh --bootstrap-server 127.0.0.1:9092 --entity-type brokers --entity-name 1 --all --describe
#
# topic info:
# kafka-topics.sh --bootstrap-server 127.0.0.1:9092 --describe --topic httpserver_access_log
#
# product and consume:
# kafka-console-producer.sh --bootstrap-server 127.0.0.1:9092 --topic httpserver_access_log
# kafka-console-consumer.sh --bootstrap-server 127.0.0.1:9092 --topic httpserver_access_log --from-beginning
#
# get messages count in a topic (include deleted msg):
# kafka-run-class.sh kafka.tools.GetOffsetShell --broker-list 127.0.0.1:9092 --topic httpserver_access_log | awk -F  ":" '{sum += $3} END {print sum}'
# 

function product_messages {
    echo "product kafka messages by access http server."
    for i in $(seq 30 40); do
        curl "http://localhost:8080/?data${i}"
        # curl "http://127.0.0.1:8080/?data${i}"
    done
}

if [[ $1 == "product" ]]; then
    product_messages
    exit 0
fi

echo "done"
