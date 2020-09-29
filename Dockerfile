FROM golang:1.15.2-alpine3.12 AS builder

RUN apk update && \
    apk add --no-cache --update make bash ca-certificates && \
    update-ca-certificates

WORKDIR /app
ENV GO111MODULE=on
COPY . .
RUN go mod download
RUN make build-docker

FROM alpine:3.12

COPY --from=builder /app/build/*/* /usr/local/bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 7373

ENV VISIGOTH_TCP_BIND_TO=:7374
ENV VISIGOTH_HTTP_BIND_TO=7373

CMD /usr/local/bin/server -tcp $VISIGOTH_TCP_BIND_TO -http $VISIGOTH_HTTP_BIND_TO
