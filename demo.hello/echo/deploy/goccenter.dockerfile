FROM golang:1.15.2

COPY bin/goc_linux /go/bin/goc

# go center port
EXPOSE 7777

CMD goc server
