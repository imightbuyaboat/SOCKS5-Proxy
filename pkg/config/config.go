package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

type Config struct {
	ListenAddress    string `json:"listen_address"`
	RemoteTCPAddress string `json:"remote_tcp_address"`
	RemoteUDPAddress string `json:"remote_udp_address"`
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
	host, portStr, err := net.SplitHostPort(c.ListenAddress)
	if err != nil {
		return err
	}
	if err = validateHost(host); err != nil {
		return err
	}
	if err = validatePort(portStr); err != nil {
		return err
	}

	host, portStr, err = net.SplitHostPort(c.RemoteTCPAddress)
	if err != nil {
		return err
	}
	if err = validateHost(host); err != nil {
		return err
	}
	if err = validatePort(portStr); err != nil {
		return err
	}

	host, portStr, err = net.SplitHostPort(c.RemoteUDPAddress)
	if err != nil {
		return err
	}
	if err = validateHost(host); err != nil {
		return err
	}
	if err = validatePort(portStr); err != nil {
		return err
	}

	return nil
}

func validateHost(host string) error {
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("invalid ip")
	}
	return nil
}

func validatePort(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %v", err)
	}

	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port")
	}
	return nil
}
