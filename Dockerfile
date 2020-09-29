FROM golang:1.14.6-alpine3.12 AS builder

RUN apk update && \
    apk add --no-cache --update make bash ca-certificates && \
    update-ca-certificates

WORKDIR /app
ENV GO111MODULE=on
COPY . .
RUN go mod download
RUN make build-docker

FROM scratch

COPY --from=builder /app/build/*/* /usr/local/bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 7373
ENTRYPOINT [ "/usr/local/bin/server" ]
