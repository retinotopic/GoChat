FROM golang:1.24.2-alpine3.21 as builder

WORKDIR /build
COPY --from=cmd_app . .
COPY go.mod go.sum .
COPY /app ./app/
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /main main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
COPY --from=backendimage /usr/local/share/ca-certificates/mkcert-ca.crt /usr/local/share/ca-certificates/mkcert-ca.crt
RUN update-ca-certificates

COPY --from=builder main /bin/main
ENTRYPOINT ["/bin/main"]
