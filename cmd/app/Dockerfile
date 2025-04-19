FROM golang:1.24.2-alpine3.21 as builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM alpine:3.21
COPY --from=builder /build/cmd /cmd
ENTRYPOINT ["/cmd"]
