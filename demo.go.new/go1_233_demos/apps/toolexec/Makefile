test_dir = /tmp/test/go-app

default: version

version:
	@go version

prepare-toolexec:
	rm -r $(test_dir); cp -r ./tmpl $(test_dir)
	echo "module zjin.goapp.demo.toolexec\n\ngo 1.23.3" > $(test_dir)/go.mod
	go build -o $(test_dir)/toolexec_test main.go

run-toolexec:
	cd $(test_dir); go build -a -toolexec=$(test_dir)/toolexec_test app.go
