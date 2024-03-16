# Go Embed

## Build

Build http server bin, and move to a tmp dir:

```sh
go build -o embed_server .

# run in tmp dir
mv embed_server /tmp/test
cd /tmp/test; ./embed_server
```

## API Test

Test server health:

```sh
curl http://localhost:8080/healthz
```

Get embed static `index.html` file:

```sh
curl http://localhost:8080/

# first, http server sends response 301 with redirect to "http://localhost:8080/";
# here add "-L" option to resolve redirect.
curl -L http://localhost:8080/index.html
```

Get embed static `raw.json` file:

```sh
curl "http://localhost:8080/raw.json"
```

