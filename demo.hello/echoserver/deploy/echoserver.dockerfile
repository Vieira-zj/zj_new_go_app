FROM golang:1.15.2

WORKDIR /go/src/demo.hello

COPY bin/goc_linux /go/bin/goc
COPY echoserver echoserver
COPY go.mod .

# goc agent port
EXPOSE 8080
# server port
EXPOSE 8081

RUN cd /go/src/demo.hello/echoserver && go mod tidy && \
  goc build --center=http://127.0.0.1:7777 --agentport :8080 -o /go/bin/echoserver

# CMD sh -c "while true; do echo 'hello'; sleep 10; done;"
CMD echoserver
