# Kafka Middleware Demo

## Http Server

Topics:

- `httpserver_important`: sync producer sends url.query messages.
- `httpserver_access_log`: async producer sends http server access log messages.

## Kafka Cli

- Broker config info

```sh
kafka-configs.sh --bootstrap-server 127.0.0.1:9092 --entity-type brokers --entity-name 1 --all --describe
```

- Topic info

```sh
kafka-topics.sh --bootstrap-server 127.0.0.1:9092 --describe --topic httpserver_access_log
```

- Product and Consume

```sh
kafka-console-producer.sh --bootstrap-server 127.0.0.1:9092 --topic httpserver_access_log
kafka-console-consumer.sh --bootstrap-server 127.0.0.1:9092 --topic httpserver_access_log --from-beginning
```

- Get messages count in a topic (include deleted msg)

```sh
kafka-run-class.sh kafka.tools.GetOffsetShell --broker-list 127.0.0.1:9092 --topic httpserver_access_log | awk -F  ":" '{sum += $3} END {print sum}'
```

