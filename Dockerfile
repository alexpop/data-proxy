FROM alpine:latest

COPY ./dp /service-bin

CMD echo && if [ -z "$CONFIG_YAML" ]; then \
      echo "This container requires ENV variable CONFIG_YAML to be set, aborting!"; \
      exit 1; \
    else \
      echo "$CONFIG_YAML" > /config.yaml; \
      /service-bin /config.yaml; \
    fi