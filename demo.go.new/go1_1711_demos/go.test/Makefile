default: go-version

go-version:
	go version

go-test-race:
	go test -run=TestBatchOp -count=1 -timeout=3s -v -race

go-test-fuzz:
	go test -fuzz=FuzzGetRecord -fuzztime=5s -v

go-test-benchmark:
	# go test -bench=BenchmarkConvStr2bytes -run=^$ -benchtime=5s -benchmem -v
	go test -bench=BenchmarkConv -run=^$ -benchtime=5s -benchmem -v
