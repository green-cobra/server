FROM alpine:latest

COPY go-green-cobra-server /

USER 999
ENTRYPOINT ["/go-green-cobra-server"]
STOPSIGNAL SIGINT