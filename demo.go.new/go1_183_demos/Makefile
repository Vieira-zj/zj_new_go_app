default: version

version:
	@go version

debug:
	@go run main.go

govet:
	@go vet ./...

golint:
	@golangci-lint run -c .golangci.yaml ./...
