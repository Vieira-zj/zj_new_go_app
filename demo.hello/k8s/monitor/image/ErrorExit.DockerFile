FROM busybox:1.30

CMD ["sh", "-c", "for i in $(seq 15); do echo 'hello'; sleep 1; done; echo 'exit'; exit 99"]
