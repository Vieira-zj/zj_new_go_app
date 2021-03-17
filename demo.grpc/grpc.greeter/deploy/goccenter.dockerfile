FROM ubuntu:20.04

COPY bin/goc_linux /app/goc

EXPOSE 7777

CMD /app/goc server