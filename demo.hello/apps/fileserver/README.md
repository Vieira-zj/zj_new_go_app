# File Server

1. Upload and download files with auth.
2. Clear expired history files.

## APIs

- Run server:

```sh
./fileserver -m=test,testfs -t=hello
```

- Test API:

```sh
curl http://localhost:8081/
curl http://localhost:8081/ping | jq .

curl http://localhost:8081/modules | jq .
```

- File server API:

Upload and download file:

```sh
# invalid token
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "Token: aGVsbG8K" -F "file=@./report.html"
# invalid file type
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "Token: dGVzdGZz" -F "file=@./report.txt"

# valid token
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "Token: dGVzdGZz" -F "file=@./report.html"
# browser file
curl -v http://localhost:8081/public/lint/testfs/report.html
```

- Register static router:

```sh
# register by invalid token
curl "http://localhost:8081/register?module=goc" -H "Token: test"
# register existing module
curl "http://localhost:8081/register?module=test" -H "Token: hello"

# register
curl "http://localhost:8081/register?module=goc" -H "Token: hello"
# upload
curl -XPOST http://localhost:8081/upload -H "Content-Type: multipart/form-data" -H "Token: Z29j" -F "file=@./test.html"
# download
curl -v http://localhost:8081/public/lint/goc/test.html
```

