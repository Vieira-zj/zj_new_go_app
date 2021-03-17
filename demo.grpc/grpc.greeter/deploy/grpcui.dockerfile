FROM ubuntu:20.04

COPY bin/grpcui_linux /app/grpcui

EXPOSE 8080

CMD /app/grpcui -plaintext -bind 0.0.0.0 -port 8080 localhost:50051