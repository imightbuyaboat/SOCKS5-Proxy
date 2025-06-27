package main

import (
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/udp"
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

	go tcp.NewTCPAssociateListener(config, zapLogger).Start()
	udp.NewUDPAssociateListener(config, zapLogger).Start()
}
