default:
	go version

# go plugin 主要支持 linux 系统, 在 mac 系统上不完善
setup:
	docker run -it --rm --name golang -v $(shell pwd):/src -p 8080:8080 golang:1.21 sh

build-plugin:
	go build -buildmode=plugin -o greeter.so greeter.go

run:
	go run main.go

test-api:
	curl http://localhost:8080/healthz
