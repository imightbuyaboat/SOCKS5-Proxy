package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

type Config struct {
	TCPListener TCPListenerConfig
	UDPListener UDPListenerConfig
	WebGUI      WebGUIConfig
	Logger      LoggerConfig
}

type TCPListenerConfig struct {
	TCPRelayServerAddress string `json:"tcp_relay_server_address"`
}

type UDPListenerConfig struct {
	UDPRelayServerAddress string `json:"udp_relay_server_address"`
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

	var tcpListenerConfig TCPListenerConfig
	if err = json.NewDecoder(file).Decode(&tcpListenerConfig); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}
	file.Close()

	file, err = os.OpenFile("config.json", os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var udpListenerConfig UDPListenerConfig
	if err = json.NewDecoder(file).Decode(&udpListenerConfig); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}
	file.Close()

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
		TCPListener: tcpListenerConfig,
		UDPListener: udpListenerConfig,
		WebGUI:      webGUIConfig,
		Logger: LoggerConfig{
			LogLevel: os.Getenv("LOG_LEVEL"),
		},
	}

	if err = config.validateConfig(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %v", err)
	}

	return &config, nil
}

func (c *Config) validateConfig() error {
	if err := ValidateAddress(c.TCPListener.TCPRelayServerAddress); err != nil {
		return err
	}
	if err := ValidateAddress(c.UDPListener.UDPRelayServerAddress); err != nil {
		return err
	}
	if c.WebGUI.Port < 0 || c.WebGUI.Port > 65535 {
		return fmt.Errorf("invalid socks5_web_gui_port")
	}

	return nil
}

func ValidateAddress(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port")
	}

	ip := net.ParseIP(host).To4()
	if ip == nil {
		return fmt.Errorf("invalid format of address")
	}

	return nil
}
