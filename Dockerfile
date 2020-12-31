FROM alpine:latest

COPY ./dp /root/
COPY ./config-sample.yaml /root/

CMD /root/dp /root/config-sample.yaml
