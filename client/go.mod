module github.com/imightbuyaboat/SOCKS5-Proxy/client

go 1.24.4

replace github.com/imightbuyaboat/SOCKS5-Proxy/pkg => ../pkg

require (
	github.com/imightbuyaboat/SOCKS5-Proxy/pkg v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)
