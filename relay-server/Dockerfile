FROM golang:1.24.4 AS builder

WORKDIR /SOCKS5-proxy

COPY . .

WORKDIR /SOCKS5-proxy/relay-server

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd
RUN chmod +x main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /SOCKS5-proxy/relay-server/main .
COPY --from=builder /SOCKS5-proxy/relay-server/internal/web_gui/template ./template
COPY --from=builder /SOCKS5-proxy/relay-server/internal/web_gui/static ./static

COPY config.json .

CMD ["./main"]