FROM golang:1.24.2-alpine3.21 AS builder
RUN apk add --no-cache git

WORKDIR /build
COPY --from=cmd_server . .
COPY go.mod go.sum .
COPY /server ./server/
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /main main.go

FROM alpine/mkcert AS certs
RUN mkcert -install && \
    mkcert -key-file key.pem -cert-file cert.pem 0.0.0.0 localhost 127.0.0.1 ::1

FROM alpine:3.21
COPY --from=builder /build/server/db/postgres/migrations /bin/maindir/migrations
COPY --from=builder main /bin/maindir
COPY --from=certs /key.pem /cert.pem .
ENTRYPOINT ["/bin/maindir/main"]
