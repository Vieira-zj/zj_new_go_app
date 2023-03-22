install:
	go get gitlab.com/golang-commonmark/markdown@bf3e522c626a

clean:
	rm -r bin
	mkdir bin

build-nopgo:
	go build -o bin/markdown_nopgo

build-pgo:
	go build -pgo=auto -o bin/markdown_pgo

test:
	curl http://localhost:8080/render --data-binary @data/README.md

load-test:
	./load.sh cpu_nopgo.prof

open-prof-text:
	go tool pprof --text bin/cpu_nopgo.prof

open-prof-html:
	go tool pprof -http :8081 bin/cpu_nopgo.prof
