package main

import (
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/socks5"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	zapLogger := logger.GetLogger()

	if err := godotenv.Load(); err != nil {
		zapLogger.Fatal("failed to load .env",
			zap.Error(err))
	}

	config, err := config.LoadConfig()
	if err != nil {
		zapLogger.Fatal("failed to load config",
			zap.Error(err))
	}

	listener := socks5.NewSOCKS5Listener(config, zapLogger)
	listener.Start()
}
