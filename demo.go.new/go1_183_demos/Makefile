build_flag := -ldflags='-extldflags -static' -tags='embed kodo'

default: version

echo:
	@echo "build flag: $(build_flag)"

version:
	@go version

debug:
	@go run main.go

govet:
	@go vet ./...

delta-golint:
	@./run.sh delta-golint

golint:
	@golangci-lint run -c .golangci.yaml ./...
