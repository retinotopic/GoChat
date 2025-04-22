FROM golang:1.24.2-alpine3.21 AS builder
RUN apk add --no-cache git

WORKDIR /build
COPY --from=cmd_server . .
COPY go.mod go.sum .
COPY /server ./server/
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /main main.go

FROM alpine/mkcert@sha256:a8f4f5af61908b4c79c2e9d1e5f23e747f29de174649209ebafcab03d4f6d5fd AS certs
RUN mkcert -install && \
    mkcert -key-file key.pem -cert-file cert.pem 0.0.0.0 localhost backend 127.0.0.1 ::1

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
COPY --from=certs /key.pem /etc/ssl/private/
COPY --from=certs /cert.pem /etc/ssl/certs/
COPY --from=certs /root/.local/share/mkcert/rootCA.pem /usr/local/share/ca-certificates/mkcert-ca.crt
RUN update-ca-certificates

COPY --from=builder /build/server/db/postgres/migrations /bin/maindir/migrations
COPY --from=builder main /bin/maindir
# COPY --from=certs /key.pem /cert.pem /bin/maindir
ENTRYPOINT ["/bin/maindir/main"]
