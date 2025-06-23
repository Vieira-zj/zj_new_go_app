# Golang: Process Big File

## Cli

```sh
# generate mock data
chmod +x generate.sh
./generate.sh 10000000

# build and run
go run main.go --file=dummy_10000000_rows.csv --chunk=1000 --workers=8
```

