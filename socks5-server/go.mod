module github.com/imightbuyaboat/SOCKS5-Proxy/client

go 1.24.4

replace github.com/imightbuyaboat/SOCKS5-Proxy/pkg => ../pkg

require (
	github.com/gorilla/mux v1.8.1
	github.com/imightbuyaboat/SOCKS5-Proxy/pkg v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.7.6
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)
