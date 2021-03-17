# demo.hello

> golang demos.

## go modules

go mod cli:

```sh
go mod
```

init package:

```sh
cd [package_dir]
go mod init [package_name]
```

update dependency packages:

```sh
go mod tidy # 更新依赖
go list -m all # 查看项目所有依赖的包
go list ./...
```

## echoserver

echoserver APIs:

'''sh
curl -v "http://localhost:8081/"
curl -v "http://localhost:8081/ping"
curl -v "http://localhost:8081/users"

curl -v "http://localhost:8081/sample/01?base=5"
curl -v "http://localhost:8081/cover?cond1=true&cond2=false"
'''

## goc code coverage

1. push echoserver and goccenter image to k8s master repo

```sh
minikube mount ${HOME}/Workspaces/zj_repos/zj_go2_project/demo.hello:/tmp/app/demo.hello

minikube ssh
# run docker build
cd /tmp/app/demo.hello && ./run.sh
```

2. deploy echoserver components (deploy,service,ingress)

```sh
cd demo.hello/echoserver/deploy
kubectl apply -f deploy.yaml
```

3. check deploy service in k8s

```sh
kubectl get svc
# output: echoserver   ClusterIP   10.110.113.179

kubectl run debug-pod -it --rm --restart=Never --image=goccenter-test:v1.0 sh
# test echoserver api
cd /tmp && wget 'http://10.110.113.179:8081/' -S -O output; cat output

# check agent register in goccenter
goc list --center=http://goccenter.default.k8s
# output: {"echoserver":["http://127.0.0.1:8080"]}
```

4. check deploy service exposed by ingress (on local)

```sh
# test echoserver api
curl http://echoserver.default.k8s/
curl http://echoserver.default.k8s/ping
curl http://echoserver.default.k8s/users/
```

5. create goc coverage results

```sh
cd demo.hello/echoserver
goc profile --center=http://goccenter.default.k8s --address=http://127.0.0.1:8080 -o coverprofile.cov

# display coverage data
go tool cover -func=coverprofile.cov
# go tool cover -func=coverprofile.cov -o cover_summary.txt
# generate coverage html
go tool cover -html=coverprofile.cov -o coverage.html
```

6. generate goc coverage only includes files matching the regexp patterns

```sh
# include ping.go and user.go
goc profile --coverfile="echoserver/handlers/ping.go,users.go" \
  --center=http://goccenter.default.k8s --address=http://127.0.0.1:8080 -o coverprofile.cov

# include go files in dir "handlers"
goc profile --coverfile="handlers/.*\.go" \
  --center=http://goccenter.default.k8s --address=http://127.0.0.1:8080 -o coverprofile.cov
```

7. create cobertura html report

```sh
gocover-cobertura < coverprofile.cov > cobertura.xml
mv cobertura.xml ${HOME}/Workspaces/zj_repos/zj_go2_project
pycobertura show --format html --output cobertura.html cobertura.xml
```

## coverage results

### go原生cover工具

1. `go tool cover -func` 生成 method-level 覆盖率数据，统计结果基于行（statements）。
2. `go tool cover -html` 生成 file-level 覆盖率数据，达到分支覆盖，统计结果基于行（statements）。**行染色结果可达到路径覆盖。**

> 注：需要把`coverprofile.cov`文件放在`echoserver`项目目录，保证`go list ./...`可正常执行来解析覆盖率数据。
>
> 检查`coverprofile.cov`文件中的记录路径与`go.mod`中的module名。

### cobertura报告

cobertura的xml覆盖率数据可以与jenkins集成，生成覆盖率报告。

cobertura html 生成 file-level 覆盖率数据，与染色结果一致，达到分支覆盖，统计结果为`(stmts-miss)/stmts`。

> 注：保证`pycobertura`命令的执行路径与`cobertura.xml`中`package`的文件路径一致。

### 测试脚本

```sh
# cd project
cd echoserver/

# clear
goc clear
goc init

# check goc center and agent
go list
# dump coverage data
goc profile --coverfile="handlers/cover.go,handlers/ping.go" --service=echoserver_cover -o coverprofile.cov

# create cover func and html results
go tool cover -func=coverprofile.cov
go tool cover -html=coverprofile.cov

# api test
curl "http://localhost:8081/"
# have 4 conditions to test coverage level (语句覆盖、分支覆盖、路径覆盖)
curl "http://localhost:8081/cover?cond1=true&cond2=false"
```

