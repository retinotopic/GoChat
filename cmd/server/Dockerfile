FROM golang:1.24.2-alpine3.21 AS builder

RUN apk add --no-cache mkcert ca-certificates && \
    mkdir -p /ssl && \
    mkcert -install && \
    mkcert -cert-file /ssl/localhost.pem -key-file /ssl/localhost-key.pem 0.0.0.0 localhost 127.0.0.1 ::1

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /build/cmd .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /ssl /ssl
COPY --from=builder /build/cmd /cmd

ENTRYPOINT ["/cmd"]
