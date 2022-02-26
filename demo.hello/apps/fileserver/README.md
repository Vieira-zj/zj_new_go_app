# File Server

1. Upload and download files.
2. Clear expired history files.

## APIs

- Test API:

```sh
curl http://localhost:8081/
curl http://localhost:8081/ping | jq .
```

- File server API:

```sh
# uplaod file
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "X-Component: spba" \
    -F "file=@./lint_report.html"
```

Browser file by `http://localhost:8081/public/lint/spba/lint_report.html`.

