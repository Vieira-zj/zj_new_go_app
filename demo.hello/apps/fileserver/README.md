# File Server

## APIs

- API test:

```sh
curl http://localhost:8081/
curl http://localhost:8081/ping | jq .
```

- File server test:

```sh
# uplaod file
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "X-Component: spba" \
    -F "file=@./lint_report.html"
```

Browser file by `http://localhost:8081/public/lint/spba/lint_report.html`.

