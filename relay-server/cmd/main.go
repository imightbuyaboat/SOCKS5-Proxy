package main

import (
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/udp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/web_gui"
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

	listenerTCP := tcp.NewTCPAssociateListener(config, zapLogger)
	listenerUDP := udp.NewUDPAssociateListener(config, zapLogger)

	var ui UI = web_gui.NewWebGUI(config.RelayWebGUIPort, listenerTCP, listenerUDP)
	ui.Start()
}
