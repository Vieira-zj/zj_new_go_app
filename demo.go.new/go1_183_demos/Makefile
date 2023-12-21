build_flag := -ldflags='-extldflags -static' -tags='embed kodo'

default: version

echo:
ifdef build_flag
	@echo "build flag: $(build_flag)"
else
	@echo "no build flag found"
endif

version:
	@go version
	@echo "project: $(shell pwd)"

debug:
	@go run main.go

govet:
	@go vet ./...

golint-delta:
	@./run.sh golint-delta

golint:
	@golangci-lint run -c .golangci.yaml ./...
