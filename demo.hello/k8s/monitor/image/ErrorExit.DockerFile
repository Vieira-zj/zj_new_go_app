FROM busybox:1.30

CMD ["sh", "-c", "for i in {1..10}; do echo 'hello'; sleep 1; done; echo 'exit'; exit 99"]
