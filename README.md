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

## API endpoints in the form of HTTP `VERB: URL`

1. `GET: $URI/version`

Shows version of the service and sha256 sum of the api binary.

2. `POST: $URI/azure/workspace/:workspace_id/log/:log_name`

Forwards the payload of the request to Azure Log Analytics, where:

 * `:workspace_id` is the UUID of the workspace. Must match one of the workspace `id`s in the config file that the service is started with. A name, as defined in the workspace `name` can be used here as well.
 * `:log_name` is the name of the Azure Analytics log where data will end up. For a `log_name` of `ProxyTest`, your data will be visible under `ProxyTest_CL`. It can take up to five minutes for a new log to be visible in the Log Analytics dashboard.


## Example API calls with _curl_
```bash
# Get the version of the service
curl -X GET http://127.0.0.1:4000/version

# Send json data to the 'ProxyTest_CL' log using the workspace id
curl -X POST http://127.0.0.1:4000/azure/workspace/01234567-8383-3ca5-4b65-d12a5cda0a55/log/ProxyTest -d '{"hello":"world1"}'

# Send json data to the 'ProxyTest_CL' log using the workspace name
curl -X POST http://127.0.0.1:4000/azure/workspace/wks-americas/log/ProxyTest -d '{"hello":"world2"}'
```

## Build docker image

This repository contains a `Dockerfile` that uses a minimal `alpine` container to copy the compiled binary in.
To run the container, you have to build the binary and provide the yaml content of the config file of the service via the `CONFIG_YAML` ENV variable.

```bash
# Build the binary that will be copied in the container
./make.sh build-alpine

# Build the docker file locally
docker build --tag data-proxy .

# Create a config file for the service (e.g. ~/data-proxy-config.yaml) with your azure workspaces. Use config-sample.yaml as an example

# Run the API interactively in a container and listen on localhost (127.0.0.1), TCP port 4000
docker run --name=data-proxy-container --rm -p 127.0.0.1:4000:4000 -e "CONFIG_YAML=$(cat ~/data-proxy-config.yaml)" data-proxy
```

## Or use the public docker image published


