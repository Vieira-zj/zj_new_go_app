app-version = v0.0.1

default: version

version:
	@go version

env:
	@go env

build:
	go build -ldflags="-X 'main.Version=$(app-version)'" main.go

run:
	go run -ldflags="-X 'main.version=$(app-version)'" main.go
