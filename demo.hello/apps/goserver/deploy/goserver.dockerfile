FROM golang:1.15.2

COPY bin/goc_linux /go/bin/goc
COPY main.go /go/src

# goc agent port
EXPOSE 17890
# server port
EXPOSE 17891

RUN cd /go/src && goc build --center=http://goccenter:8080 --agentport :17890 -o /go/bin/goserver

CMD goserver
