FROM golang:1.15.2 AS builder

WORKDIR /go/src/demo.grpc

COPY bin/goc_linux /go/bin/goc
COPY grpc/greeter_server grpc/greeter_server
COPY grpc/proto grpc/proto
COPY go.mod .

# grpc server port
EXPOSE 50051
# goc agent port
EXPOSE 50052

RUN cd /go/src/demo.grpc/grpc/greeter_server && go mod tidy && \
  goc build --center=http://127.0.0.1:7777 --agentport :50052 -o /go/bin/greeter_server

FROM ubuntu:20.04

WORKDIR /app

COPY --from=builder /go/bin/greeter_server ./greeter_server

# CMD sh -c "while true; do echo 'hello'; sleep 10; done;"
CMD /app/greeter_server