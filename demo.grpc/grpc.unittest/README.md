# Grpc Mock

> Refer: <https://github.com/tokopedia/gripmock>

## Sample

1. Pull image

```sh
docker pull tkpd/gripmock
```

2. Start grpc mock server

```sh
mkdir -p /tmp/test/proto
cp proto/account/deposit.proto /tmp/test/proto
docker run --name grpcmock --rm -p 4770:4770 -p 4771:4771 -v /tmp/test/proto:/proto tkpd/gripmock /proto/deposit.proto
# Starting GripMock
# Serving stub admin on http://:4771
# grpc server pid: 769
# Serving gRPC on tcp://:4770
```

3. Register 2 grpc stubs

```sh
curl -X POST "localhost:4771/add" -d '@/tmp/test/data.json'
# Success add stub
```

Stub data:

```json
{
  "service": "DepositService",
  "method": "Deposit",
  "input": {
    "equals": {
      "amount": 100
    }
  },
  "output": {
    "data": {
      "ok": true
    },
    "error": "Invalid input"
  }
}
```

4. List grpc stubs

```sh
curl "localhost:4771/"
```

5. Run grpc client with port 4770

```sh
./client -p 4770 -m 100
./client -p 4770 -m -101
```

6. Clear stubs

```sh
curl "localhost:4771/clear"
```

