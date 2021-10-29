FROM busybox:1.30

COPY bin/pod_monitor /bin/monitor

EXPOSE 8081

CMD ["/bin/monitor", "-debug", "-mode", "cluster", "-ns", "default,k8s-test"]
