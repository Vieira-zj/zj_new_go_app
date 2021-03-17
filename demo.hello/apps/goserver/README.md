# Go simple server for goc coverage

> Refer: <https://github.com/qiniu/goc>

## goc Hello Demo

goc help:

```sh
goc -h
goc server -h
```

1. Use `goc server` to start a service registry center:

```sh
goc server
```

2. Use `goc build` to build the target service, and run the generated binary.

```sh
cd simple-go-server

# build
goc build .
# build by specified agent port (same with server port is ok)
goc build . --agentport :17891

# run server
./simple-go-server

# use "goc list" to check all registered services
goc list
```

3. Use `goc profile` to get the code coverage profile of the started simple server above:

```sh
goc profile -o coverprofile.cov

# with specified address port (agent port)
goc profile --address "http://127.0.0.1:17891" -o coverprofile.cov
```

4. Do test:

```sh
curl -v "http://localhost:17891/"
```

5. Generate code coverage report.

```sh
cd simple-go-server
cp [goc-profile.cov] .

# open in browser (html)
go tool cover -html=coverprofile.cov
# open in command line
go tool cover -func=coverprofile.cov
```

## Tips

1. You can use `--agentport` flag to specify a fixed port when calling goc build or goc install.

2. The coverage data is stored on each covered service side, so if one service needs to restart during test, this service's coverage data will be lost. For this case, you can use following steps to handle:

1) Before the service restarts, collect coverage with `goc profile -o a.cov`
2) After service restarted and test finished, collect coverage again with `goc profile -o b.cov`
3) Merge two coverage profiles together: `goc merge a.cov b.cov -o merge.cov`

## Docker + goc Hello Demo

### Build go server + goc center env

1. Verify build go server by goc in docker:

```sh
# pull go image and run container
docker pull golang:1.15.2
docker run --name golang -it --rm golang:1.15.2 sh

# build goc for linux
GOOS=linux GOARCH=amd64 go build goc.go

# copy goc and src/main.go into container
docker cp goc golang:/go/bin
docker cp main.go golang:/go/src

cd /go/src
goc build --center=http://goccenter:8080 --agentport :17890 -o goserver
```

2. Build go server and goc center image:

```sh
docker build -t goserver-test:v1.0 -f deploy/dockerfile_goserver .
docker build -t goccenter-test:v1.0 -f deploy/dockerfile_goccenter .
```

3. Run go server and goc center instance:

```sh
cd deploy
docker-compose -f deploy_compose.yaml up -d
```

4. Test go server APIs:

```sh
curl "http://localhost:17891/"
curl "http://localhost:17891/healthz"
```

### Create go coverage report

1. Generate go coverage file in docker:

From go server:

```sh
docker exec -it goserver sh
goc list --center=http://goccenter:8080
# output: {"goserver":["http://172.18.0.3:17890"]}
goc profile --center=http://goccenter:8080 --address=http://172.18.0.3:17890 -o /tmp/coverprofile.cov
```

From go center:

```sh
docker exec -it goccenter sh
goc list --center=http://localhost:8080
# output: {"goserver":["http://172.18.0.3:17890"]}
goc profile --center=http://localhost:8080 --address=http://172.18.0.3:17890 -o /tmp/coverprofile.cov
```

2. Copy go coverage file to local go project, and validate.

```sh
cd ${goserver-project-dir}
docker cp goserver:/tmp/coverprofile.cov .
go tool cover -func coverprofile.cov

# if go package of coverprofile.cov not matched, replace it with "demo.hello/apps/goserver" (by "go list")
go list
```

3. Output html coverage report:

```sh
go tool cover -html=coverprofile.cov -o coverage.html
```

### Create go cobertura coverage report

1. Create cobertura xml file from go coverage file.

gocover-cobertura: <https://github.com/t-yuki/gocover-cobertura>

```sh
gocover-cobertura < coverprofile.cov > coverage.xml
```

2. Install pycobertura to parse cobertura xml file.

pycobertura: <https://pypi.org/project/pycobertura/>

```sh
pip3 install pycobertura
```

3. Output cobertura html report:

```sh
cd ${HOME}/Workspaces/zj_repos/zj_go2_project
cp coverage.xml .
pycobertura show --format html --output coverage.html coverage.xml
```
