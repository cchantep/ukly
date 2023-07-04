FROM golang:1.20-alpine3.17 as go-builder

COPY ukli.go /tmp/

RUN cd /tmp && go build ukli.go

FROM alpine:3.17.4

COPY --from=go-builder /tmp/ukli /usr/local/bin/

RUN chmod a+x /usr/local/bin/ukli

ENTRYPOINT ["ukli"]
