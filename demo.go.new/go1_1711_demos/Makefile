VERSION := DEV-$(shell date +%s)

default: debug

lint:
	@echo "run go lint"
	@go vet ./...

fmt:
	@echo "run go format"
	@go fmt ./...

version:
	@echo "ver: $(VERSION)"

debug:
	@go run main.go
