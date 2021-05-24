FROM busybox:1.30

COPY k8s-simple-ingress-controller /bin/k8s-simple-ingress-controller

CMD ["/bin/k8s-simple-ingress-controller", "-port", "8080", "-tls-port", "8443"]
