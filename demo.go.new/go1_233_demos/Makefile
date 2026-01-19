app-version := v0.0.1

default: version

version:
	@go version

build:
	go build -ldflags="-X 'main.AppVersion=$(app-version)'" main.go

run:
	go run -ldflags="-X 'main.AppVersion=$(app-version)'" main.go
