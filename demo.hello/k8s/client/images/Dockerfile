FROM debian:buster-slim

RUN apt-get update && apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux procps curl \
  && rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

RUN curl -L https://github.com/chaos-mesh/toda/releases/download/v0.2.0/toda-linux-amd64.tar.gz -o /usr/local/bin/toda.tar.gz \
  && tar xvf /usr/local/bin/toda.tar.gz -C /usr/local/bin \
  && rm /usr/local/bin/toda.tar.gz 

RUN curl -L https://github.com/chaos-mesh/nsexec/releases/download/v0.1.5/nsexec-linux-amd64.tar.gz -o /usr/local/bin/nsexec.tar.gz \
  && tar xvf /usr/local/bin/nsexec.tar.gz -C /usr/local/bin \
  && rm /usr/local/bin/nsexec.tar.gz \
  && cp /usr/local/bin/libnsenter.so /usr/local/lib/libnsenter.so

RUN curl -L https://github.com/chaos-mesh/rs-tproxy/releases/download/v0.2.3/tproxy-linux-amd64.tar.gz -o /usr/local/bin/tproxy.tar.gz \
  && tar xvf /usr/local/bin/tproxy.tar.gz -C /usr/local/bin \
  && rm /usr/local/bin/tproxy.tar.gz 

CMD ["sh", "-c", "while true; do echo 'hello'; sleep 10; done;"]
