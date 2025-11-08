package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/udp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/web_gui"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger.InitLogger(cfg.Logger.LogLevel)
	zapLogger := logger.GetLogger()

	tcpListener := tcp.NewTCPAssociateListener(&cfg.TCPListener, zapLogger)
	udpListener := udp.NewUDPAssociateListener(&cfg.UDPListener, zapLogger)

	webGUI := web_gui.NewWebGUI(cfg.WebGUI.Port, tcpListener, udpListener)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		zapLogger.Info("starting server", zap.Int("port", cfg.WebGUI.Port))
		if err := webGUI.Start(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-stopChan
	zapLogger.Info("shutdown signal received. Shutting down server gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := webGUI.Shutdown(shutdownCtx); err != nil {
		zapLogger.Fatal("server shutdown failed", zap.Error(err))
	}

	zapLogger.Info("server gracefully stopped")
}
