FROM alpine:latest

COPY ./dp /root/
COPY ./config-sample.yaml /root/config.yaml

CMD /root/dp /root/config.yaml
