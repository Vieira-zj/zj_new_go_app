# App Profile Monitor

> Refer: <https://mp.weixin.qq.com/s/EPCu5E82LKkhayxyvs2EGA>

1. Start monitor app.

```sh
go run main.go | tee /tmp/test/profile.txt
```

2. Check realtime app run state like GC and MemStats.

Open: <http://localhost:6060/debug/statsviz/>

3. Check heap profile:

> PProf user guide: <https://pkg.go.dev/net/http/pprof>

```sh
curl http://127.0.0.1:6060/debug/pprof/heap -o heap_info.out

go tool pprof --http :9091 heap_info.out
```

