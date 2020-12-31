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
curl -X GET http://127.0.0.1:4000/version

curl -X POST http://127.0.0.1:4000/azure/workspace/wks-test/log/WebTest_Log -d '{ "hello": "world" }'
```

## Build docker image
```bash
./make build-alpine
docker build --tag data-proxy-image  .
docker run -d --name=dp-container -p 127.0.0.1:4000:4000 data-proxy-image
```
