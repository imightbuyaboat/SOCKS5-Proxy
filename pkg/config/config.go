package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

type Config struct {
	SOCKS5ServerAddress   string `json:"socks5_server_address"`
	TCPRelayServerAddress string `json:"tcp_relay_server_address"`
	UDPRelayServerAddress string `json:"udp_relay_server_address"`
}

func LoadConfig() (*Config, error) {
	file, err := os.OpenFile("config.json", os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	if err = config.validateConfig(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %v", err)
	}

	return &config, nil
}

func (c *Config) validateConfig() error {
	if err := validateAddress(c.SOCKS5ServerAddress); err != nil {
		return err
	}
	if err := validateAddress(c.TCPRelayServerAddress); err != nil {
		return err
	}
	if err := validateAddress(c.UDPRelayServerAddress); err != nil {
		return err
	}

	return nil
}

func validateAddress(addr string) error {
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
