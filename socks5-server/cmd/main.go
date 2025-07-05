package main

import (
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/socks5"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
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

	listener := socks5.NewSOCKS5Listener(config, zapLogger)
	listener.Start()
}
