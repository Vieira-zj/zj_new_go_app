# Kafka Middleware Demo

## Project

- `client.go` / `client_test.go`: kafka admin client to list/create/delete topic.

- `main.go`: a http server as kafka producer.

- `consumer.group/main.go`: a kafka consumer group to parallel consume message.
  - parallel is equal to partition number of topic
  - mark consumed messages, and will continue from last offset

## Env

Topics/ConsumerGroup for test:

- `httpserver_important/consumer-group-test`: sync producer sends url.query messages.

- `httpserver_access_log/consumer-group-access-log`: async producer sends http server access log messages by middleware.

## Kafka Cli

- Kafka version

```sh
kafka-topics.sh --version
```

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

- Consumer Group

```sh
kafka-consumer-groups.sh --list --bootstrap-server 127.0.0.1:9092
kafka-consumer-groups.sh --describe --group consumer-group-access-log --bootstrap-server 127.0.0.1:9092
```

