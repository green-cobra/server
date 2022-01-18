FROM alpine:latest

COPY server /

USER 999
ENTRYPOINT ["/server"]
STOPSIGNAL SIGINT