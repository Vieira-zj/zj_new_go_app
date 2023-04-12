setup-cluster:
	docker-compose up -d

clearup-cluster:
	docker-compose down

start-cluster:
	docker-compose start zookeeper kafka

stop-cluster:
	docker-compose stop

run-http-srv-debug:
	go run main.go

run-consumer-group-client-debug:
	go run consumer.group/main.go -topics=httpserver_access_log -group=consumer-group-access-log

run-consume-all:
	go test -timeout 30s -run TestConsumeAll go1_1711_demo/middlewares/kafka -v -count=1 -timeout=15s

run-product-msgs-test:
	./run.sh product
