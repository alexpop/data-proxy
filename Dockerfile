FROM alpine:latest

# ./dp is a static go binary of the service built using: ./make.sh build-alpine
COPY ./dp /service-bin

# Don't start the service if the CONFIG_YAML env variable is undefined
CMD echo && if [ -z "$CONFIG_YAML" ]; then \
      echo "This container requires ENV variable CONFIG_YAML to be set, aborting!"; \
      exit 1; \
    else \
      echo "$CONFIG_YAML" > /config.yaml; \
      /service-bin /config.yaml; \
    fi
