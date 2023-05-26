# E2E Code Coverage Test

## E2E Cover Test Workflow

1. Build http serve bin with cover injected.

```text
./run.sh build-cover
httpserve-cover generated.
```

Actually, `httpserve-cover` is a bin with test wrapped and cover probe injected, see help:

```text
./httpserve-cover -h
Usage of ./httpserve-cover:
  -test.bench regexp
    	run only benchmarks matching regexp
  -test.benchmem
    	print memory allocations for benchmarks
...
```

2. Build api test bin.

```text
./run.sh build-test
httpserve-test generated.
```

3. Start http serve.

`cd /tmp/test; ./httpserve-cover -test.coverprofile=results.cov`

4. Run e2e api test.

`cd /tmp/test; ENV=test ./httpserve-test`

5. Stop cover test by call an injected rest api.

`curl http://localhost:8080/cover`

Then http serve app will exit, and cover profile is generated.

6. Create func and html cover report from profile.

`./run.sh cover-report`

