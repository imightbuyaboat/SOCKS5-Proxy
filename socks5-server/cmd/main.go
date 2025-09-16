package main

import (
	"context"
	"fmt"
	"os"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/postgres"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/socks5"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	zapLogger := logger.GetLogger()

	config, err := config.LoadConfig()
	if err != nil {
		zapLogger.Fatal("failed to load config",
			zap.Error(err))
	}

	if err = godotenv.Load(); err != nil {
		zapLogger.Fatal("failed to load .env",
			zap.Error(err))
	}

	postgresUrl := fmt.Sprintf("postgresql://%s:%s@127.0.0.1:5432/%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))

	postgresStorage, err := postgres.NewPostgresStorage(context.Background(), postgresUrl)
	if err != nil {
		zapLogger.Fatal("failed to create PostgresStorage",
			zap.Error(err))
	}

	listener := socks5.NewSOCKS5Listener(config, zapLogger, postgresStorage)
	listener.Start()
}
