FROM golang:1.15.2

COPY bin/goc_linux /go/bin/goc

# go center port
EXPOSE 8080

CMD goc server --port=:8080
