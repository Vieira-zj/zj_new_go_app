#!/bin/bash
set -eu

#
# page:
# http://localhost:8081/static
# http://localhost:8081/public/page_basic.html
#
# rest api:
# curl http://localhost:8081/
# curl http://localhost:8081/ping
#
# curl -XPOST http://localhost:8081/echo -D '{"msg":"hello"}'
# curl -XPOST http://localhost:8081/echo?timeout=1s -H "X-Test:OptionHeader" -D '{"msg":"hello"}'
#
# curl -XPOST http://localhost:8081/upload -H "X-Md5:c233b531f7d18e4a1e3b94c8c7bd337a" -F "files=@/tmp/test/raw.txt"
# F/--form <name=content> Specify HTTP multipart POST data.
#
# curl -v -XPOST http://localhost:8081/user -d '{"birthday":"10/07","timezone":"Asia/Shanghai"}'
#

function http_serve {
    go run .
}

http_serve
