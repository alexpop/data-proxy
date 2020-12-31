FROM alpine:latest

COPY ./dp /root/
COPY ./config.yaml /root/config.yaml

CMD /root/dp /root/config.yaml
