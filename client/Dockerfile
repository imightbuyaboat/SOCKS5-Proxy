FROM golang:1.24.4 AS builder

WORKDIR /SOCKS5-proxy

COPY . .

WORKDIR /SOCKS5-proxy/client

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd
RUN chmod +x main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /SOCKS5-proxy/client/main .
COPY config.json .

CMD ["./main"]