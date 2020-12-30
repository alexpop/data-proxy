# README #

## Dependencies

* [Go](https://golang.org/doc/install)

Tested with these versions:
```bash
$ go version
go version go1.14.8 linux/amd64
```

## Start the API locally

```bash
go get

# Compile and run it with go:
go run *.go config-sample.yaml

# or you can compile the binary and then execute it:
./make.sh build-linux   # tested on Ubuntu
./make.sh build-mac     # tested on macOS Catalina
./data-proxy config-sample.yaml
```

## Call the API
```bash
curl -X GET http://127.0.0.1:4000/v1/version

curl -X POST http://127.0.0.1:4000/v1/azure/workspace/wks-test/log/WebTest_Log -d '{ "hello": "world" }'
```