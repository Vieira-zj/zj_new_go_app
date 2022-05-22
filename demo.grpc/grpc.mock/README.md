# Demo: openmock

> Refer: <https://github.com/checkr/openmock>
>

## Http Mock

1. build openmock

```sh
cd openmock/
make build_swagger
```

2. create html mock template `templates/http.yaml`

3. run openmock server with http enabled (default)

```sh
OPENMOCK_HTTP_PORT=8081 OPENMOCK_TEMPLATES_DIR=./templates ./bin/openmock
```

4. test mock rest api

```sh
curl -i http://localhost:8081/ping
curl -i http://localhost:8081/hello
curl -i http://localhost:8081/slow_endpoint
curl -i -XPOST http://localhost:8081/query_string?foo=bar&test=true
curl -i http://localhost:8081/token -H "X-Token:t1234" -H "Y-Token:t1234"
curl -i -XPOST http://localhost:8081/query_body -d '{"foo":123}'

# abstract behavior
curl http://localhost:8081/fruit-of-the-day?day=monday | jq .
# {"fruit": "apple"}
curl http://localhost:8081/fruit-of-the-day?day=tuesday | jq .
# {"fruit": "potato"}
```

## Grpc Mock

1. add custom openmock functionality in `openmock.go`

```golang
import (
	// ...
	"github.com/checkr/openmock/demo_protobuf"
	"github.com/golang/protobuf/proto"
)

func (om *OpenMock) Start() {
	// ...
	if om.GRPCEnabled {
		om.GRPCServiceMap = registerCustomGRPCMockSvc()
		go om.startGRPC()
	}
	//  ...
}

func registerCustomGRPCMockSvc() map[string]GRPCService {
	return map[string]GRPCService{
		"demo_protobuf.ExampleService": {
			"ExampleMethod": GRPCRequestResponsePair{
				Request:  proto.MessageV2(&demo_protobuf.ExampleRequest{}),
				Response: proto.MessageV2(&demo_protobuf.ExampleResponse{}),
			},
		},
	}
}
```

2. build openmock

```sh
cd openmock/
make build_swagger
```

3. create grpc mock template `templates/grpc.yaml`

4. run openmock server with grpc enabled

```sh
OPENMOCK_HTTP_ENABLED=false OPENMOCK_GRPC_ENABLED=true OPENMOCK_GRPC_PORT=50051 \
  OPENMOCK_TEMPLATES_DIR=./templates ./bin/openmock
```

5. test mock grpc api

```sh
cd grpc.mock/client
go run main.go
```

## Http Mock with Redis Ops

Use Redis for stateful things (by default, OpenMock uses an in-memory miniredis).

1. create html mock with redis ops template `templates/redis.yaml`

2. run openmock server

```sh
OPENMOCK_HTTP_PORT=8081 OPENMOCK_REDIS_URL=redis://127.0.0.1:6379 \
  OPENMOCK_TEMPLATES_DIR=./templates ./bin/openmock --port 9998
```

3. test mock rest api

```sh
curl http://localhost:8081/test_redis -H "X-TOKEN:t123"  | jq .
```

