package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/parser"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/postgres"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/socks5"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/udp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/web_gui"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
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

	postgresCtx, postgresCancel := context.WithCancel(context.Background())
	defer postgresCancel()
	postgresStorage, err := postgres.NewPostgresStorage(postgresCtx, cfg.Storage.URL, cfg.Storage.MigrationsPath)
	if err != nil {
		zapLogger.Fatal("PostgresStorage error",
			zap.Error(err))
	}

	p := parser.NewInMemoryParser()
	tcpAssociateHandler := tcp.NewTCPAssociateHandler(zapLogger)
	udpAssociateHandler := udp.NewUDPAssociateHandler(zapLogger)

	listener := socks5.NewSOCKS5Listener(&cfg.SOCKS5Listener, p, postgresStorage, tcpAssociateHandler, udpAssociateHandler, zapLogger)

	webGUI := web_gui.NewWebGUI(cfg.WebGUI.Port, listener)

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

	postgresCancel()

	zapLogger.Info("server gracefully stopped")
}
