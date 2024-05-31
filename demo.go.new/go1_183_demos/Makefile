build_flag := -ldflags='-extldflags -static' -tags='embed kodo'

default: version

echo/%:  # example: make echo/all, make echo/go
	@echo arg=$*
ifdef build_flag
	@echo "build flag: $(build_flag)"
else
	@echo "no build flag found"
endif

check-env:
ifndef ENV
	@echo "var 'ENV' is not set, and default to 'dev'."
else
	@echo "run env $(ENV)."
endif

version:
	@go version
	@echo "project root: $(shell pwd)"

debug:
	@GOMEMLIMIT=512MiB go run main.go

govet:
	@go vet ./...

golint-delta:
	@./run.sh golint-delta

golint:
	@golangci-lint run -c .golangci.yaml ./...
