## About

Send data to Azure Log Analytics with one line of code:

```bash
curl -X POST http://127.0.0.1:4000/azure/workspace/wks-americas/log/WebTest -d '{"hello":"2021"}'
```

^ this assumes the data-proxy service is listening for requests on TCP port 4000. See more details below.

## Spin the service quickly via docker

 * [using the public image](https://hub.docker.com/r/alexpop/data-proxy):

```bash
docker pull alexpop/data-proxy:latest

docker run --name=data-proxy-container --rm -p 127.0.0.1:4000:4000 -e "CONFIG_YAML=$(cat ~/data-proxy-config.yaml)" alexpop/data-proxy:latest
```

^ where `~/data-proxy-config.yaml` is a local config file with the workspaces to allow the service to send data to. Use `config-sample.yaml` in this repo as an example.

 * or build the service from source with the instructions below

## Build Dependencies

  * [Go](https://golang.org/doc/install)
  * [git](https://git-scm.com/downloads)

Tested with these versions:
```bash
$ go version
go version go1.14.8 linux/amd64

$ git --version
git version 2.27.0
```

## Start the API locally

```bash
go get

# Compile and run it with go:
go run cmd/*.go config-sample.yaml

# or you can compile the binary and then execute it:
./make.sh build-linux   # tested on Ubuntu
./make.sh build-mac     # tested on macOS Catalina
./data-proxy config-sample.yaml
```

## Deploy the service to Google Cloud
```bash
gcloud auth login
gcloud config set project myproject
gcloud functions list

# new deployment or update existing one. All options here: https://cloud.google.com/sdk/gcloud/reference/functions/deploy
gcloud functions deploy myfunction --region us-central1 --runtime go113 --memory=128MB --trigger-http --entry-point HttpHook --allow-unauthenticated --source ~/b/data-proxy --set-env-vars=CONFIG_YAML_CONTENT=`base64 -w 0 ../config-proxy.yaml`
```


## API endpoints in the form of HTTP `VERB: URL`

### GET: $URI/version

Shows version of the service and sha256 sum of the api binary, for example:
```json
{
  "data": {
    "binary_version": "v1.1.0-1-gd8d0b37",
    "binary_sha256": "e5b4a33ca29d21d25f7fb211315293705844e8b76bbe99f58ad5d688bd72d520"
  }
}
```

### POST: $URI/azure/workspace/:workspace_id/log/:log_name

Forwards the payload of the request to Azure Log Analytics, where:

 * `:workspace_id` is the UUID of the workspace. Must match one of the workspace `id`s in the config file that the service is started with. A name, as defined in the workspace `name` can be used here as well.
 * `:log_name` is the name of the Azure Analytics log where data will end up. For a `log_name` of `ProxyTest`, your data will be visible under `ProxyTest_CL`. It can take up to five minutes for a new log to be visible in the Log Analytics dashboard.

### GET: $URI/stats

Shows basic stats related to the response codes of the service since start time, for example:
```json
{
  "data": {
    "start_time": "2021-01-03T22:11:32Z",
    "response_codes": {
      "200": 234,
      "404": 11,
      "503": 4
    }
  }
}
```

## Example API calls with _curl_
```bash
# Get the version of the service
curl -X GET http://127.0.0.1:4000/version

# Send json data to the 'ProxyTest_CL' log using the workspace id
curl -X POST http://127.0.0.1:4000/azure/workspace/01234567-8383-3ca5-4b65-d12a5cda0a55/log/ProxyTest -d '{"hello":"world1"}'

# Send json data to the 'ProxyTest_CL' log using the workspace name
curl -X POST http://127.0.0.1:4000/azure/workspace/wks-americas/log/ProxyTest -d '{"hello":"world2"}'

# Get service stats
curl -X GET http://127.0.0.1:4000/stats
```

## Build docker image

This repository contains a `Dockerfile` that uses a minimal `alpine` container to copy the compiled binary in.
To run the container, you have to build the binary and provide the yaml content of the config file of the service via the `CONFIG_YAML` ENV variable.

```bash
# Build the binary that will be copied in the container
./make.sh build-alpine

# Build the docker file locally
docker build --tag data-proxy .

# Create a config file for the service (e.g. ~/data-proxy-config.yaml) with your azure workspaces. Use config-sample.yaml in this repo as an example.

# Run the API interactively in a container and listen on localhost (127.0.0.1), TCP port 4000
docker run --name=data-proxy-container --rm -p 127.0.0.1:4000:4000 -e "CONFIG_YAML=$(cat ~/data-proxy-config.yaml)" data-proxy
```
