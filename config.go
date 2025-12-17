package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Brokers    []string `json:"brokers"`
	Topic      string   `json:"topic"`
	CertFile   string   `json:"cert_file"`
	KeyFile    string   `json:"key_file"`
	CAFile     string   `json:"ca_file"`
	KeySerde   string   `json:"key_serde"`   // "string", "json", "bytearray"
	ValueSerde string   `json:"value_serde"` // "string", "json", "bytearray"
	UseAuth    bool     `json:"use_auth"`
}

// LoadConfig loads configuration from file
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".kafka-producer.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &Config{
				Brokers:    []string{"localhost:9092"},
				Topic:      "test-topic",
				UseAuth:    false,
				KeySerde:   "json",
				ValueSerde: "json",
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".kafka-producer.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}
