package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/validator"
)

type Config struct {
	Storage        StorageConfig
	SOCKS5Listener SOCKS5ListenerConfig
	WebGUI         WebGUIConfig
	Logger         LoggerConfig
}

type StorageConfig struct {
	URL            string
	MigrationsPath string
}

type SOCKS5ListenerConfig struct {
	SOCKS5ServerAddress   string `json:"socks5_server_address"`
	TCPRelayServerAddress string `json:"tcp_relay_server_address"`
	UDPRelayServerAddress string `json:"udp_relay_server_address"`
	AllowNoAuth           bool   `json:"allow_no_auth"`
	SOCKS5WebGUIPort      int    `json:"socks5_web_gui_port"`
}

type WebGUIConfig struct {
	Port int `json:"web_gui_port"`
}

type LoggerConfig struct {
	LogLevel string
}

func LoadConfig() (*Config, error) {
	file, err := os.OpenFile("config.json", os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var socks5ListenerConfig SOCKS5ListenerConfig
	if err = json.NewDecoder(file).Decode(&socks5ListenerConfig); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}
	file.Close()

	if err = socks5ListenerConfig.validateConfig(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %v", err)
	}

	file, err = os.OpenFile("config.json", os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var webGUIConfig WebGUIConfig
	if err = json.NewDecoder(file).Decode(&webGUIConfig); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	config := Config{
		Storage: StorageConfig{
			URL: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
				os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB")),
			MigrationsPath: os.Getenv("MIGRATIONS_PATH"),
		},
		SOCKS5Listener: socks5ListenerConfig,
		WebGUI:         webGUIConfig,
		Logger: LoggerConfig{
			LogLevel: os.Getenv("LOG_LEVEL"),
		},
	}

	return &config, nil
}

func (c *SOCKS5ListenerConfig) validateConfig() error {
	if err := validator.ValidateAddress(c.SOCKS5ServerAddress); err != nil {
		return err
	}
	if err := validator.ValidateAddress(c.TCPRelayServerAddress); err != nil {
		return err
	}
	if err := validator.ValidateAddress(c.UDPRelayServerAddress); err != nil {
		return err
	}
	if c.SOCKS5WebGUIPort < 0 || c.SOCKS5WebGUIPort > 65535 {
		return fmt.Errorf("invalid socks5_web_gui_port")
	}

	return nil
}
