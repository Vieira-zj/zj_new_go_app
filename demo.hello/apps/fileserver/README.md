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

Upload report html file:

```sh
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "X-Component: spba" \
    -F "file=@./lint_report.html"
```

Browser html file by `http://localhost:8081/public/lint/spba/lint_report.html`.

- Register static router:

```sh
# register
curl "http://localhost:8081/register?module=goc"
# upload
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "X-Component: goc" \
    -F "file=@./test.txt"
# download
curl http://localhost:8081/public/lint/goc/test.txt
```

