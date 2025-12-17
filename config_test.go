package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Default(t *testing.T) {
	// Use a non-existent path to force default config
	homeDir := t.TempDir()

	// Temporarily change home dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(config.Brokers) != 1 || config.Brokers[0] != "localhost:9092" {
		t.Errorf("Expected default broker localhost:9092, got %v", config.Brokers)
	}

	if config.Topic != "test-topic" {
		t.Errorf("Expected default topic test-topic, got %s", config.Topic)
	}

	if config.UseAuth {
		t.Error("Expected UseAuth to be false by default")
	}

	if config.KeySerde != "json" {
		t.Errorf("Expected default KeySerde json, got %s", config.KeySerde)
	}

	if config.ValueSerde != "json" {
		t.Errorf("Expected default ValueSerde json, got %s", config.ValueSerde)
	}
}

func TestLoadConfig_ExistingFile(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, ".kafka-producer.json")

	// Create test config
	testConfig := &Config{
		Brokers:    []string{"broker1:9092", "broker2:9092"},
		Topic:      "my-topic",
		CertFile:   "/path/to/cert.pem",
		KeyFile:    "/path/to/key.pem",
		CAFile:     "/path/to/ca.pem",
		UseAuth:    true,
		KeySerde:   "string",
		ValueSerde: "bytearray",
	}

	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Temporarily change home dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(config.Brokers) != 2 {
		t.Errorf("Expected 2 brokers, got %d", len(config.Brokers))
	}

	if config.Topic != "my-topic" {
		t.Errorf("Expected topic my-topic, got %s", config.Topic)
	}

	if !config.UseAuth {
		t.Error("Expected UseAuth to be true")
	}

	if config.KeySerde != "string" {
		t.Errorf("Expected KeySerde string, got %s", config.KeySerde)
	}

	if config.ValueSerde != "bytearray" {
		t.Errorf("Expected ValueSerde bytearray, got %s", config.ValueSerde)
	}
}

func TestSaveConfig(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, ".kafka-producer.json")

	// Temporarily change home dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	testConfig := &Config{
		Brokers:    []string{"test-broker:9092"},
		Topic:      "test-topic",
		CertFile:   "/cert.pem",
		KeyFile:    "/key.pem",
		CAFile:     "/ca.pem",
		UseAuth:    true,
		KeySerde:   "json",
		ValueSerde: "json",
	}

	if err := SaveConfig(testConfig); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var savedConfig Config
	if err := json.Unmarshal(data, &savedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if savedConfig.Topic != "test-topic" {
		t.Errorf("Expected topic test-topic, got %s", savedConfig.Topic)
	}

	if !savedConfig.UseAuth {
		t.Error("Expected UseAuth to be true")
	}

	// Check file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, ".kafka-producer.json")

	// Write invalid JSON
	if err := os.WriteFile(configPath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Temporarily change home dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
