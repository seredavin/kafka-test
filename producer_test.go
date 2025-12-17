package main

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/sarama"
)

func TestEncodeValue_String(t *testing.T) {
	config := &Config{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test",
		KeySerde:   "string",
		ValueSerde: "string",
	}

	producer := &KafkaProducer{
		config: config,
	}

	encoder := producer.encodeValue("test-value", "string")

	// Encode and check
	bytes, err := encoder.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if string(bytes) != "test-value" {
		t.Errorf("Expected 'test-value', got %s", string(bytes))
	}
}

func TestEncodeValue_JSON(t *testing.T) {
	config := &Config{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test",
		KeySerde:   "json",
		ValueSerde: "json",
	}

	producer := &KafkaProducer{
		config: config,
	}

	jsonData := `{"key":"value"}`
	encoder := producer.encodeValue(jsonData, "json")

	bytes, err := encoder.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if string(bytes) != jsonData {
		t.Errorf("Expected %s, got %s", jsonData, string(bytes))
	}
}

func TestEncodeValue_ByteArray(t *testing.T) {
	config := &Config{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test",
		KeySerde:   "bytearray",
		ValueSerde: "bytearray",
	}

	producer := &KafkaProducer{
		config: config,
	}

	testData := "test data"
	encoder := producer.encodeValue(testData, "bytearray")

	bytes, err := encoder.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if string(bytes) != testData {
		t.Errorf("Expected %s, got %s", testData, string(bytes))
	}
}

func TestEncodeValue_Default(t *testing.T) {
	config := &Config{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test",
		KeySerde:   "unknown",
		ValueSerde: "unknown",
	}

	producer := &KafkaProducer{
		config: config,
	}

	// Unknown serde should default to ByteEncoder
	encoder := producer.encodeValue("test", "unknown")

	bytes, err := encoder.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if string(bytes) != "test" {
		t.Errorf("Expected 'test', got %s", string(bytes))
	}
}

func TestCreateTLSConfig_InvalidCert(t *testing.T) {
	_, err := createTLSConfig("nonexistent.pem", "nonexistent.key", "nonexistent.ca")
	if err == nil {
		t.Error("Expected error for nonexistent certificates, got nil")
	}
}

func TestCreateTLSConfig_InvalidCA(t *testing.T) {
	// Create temp cert and key files
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")
	caFile := filepath.Join(tempDir, "ca.pem")

	// Write minimal valid cert (this will fail at CA stage)
	certPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)

	keyPEM := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)

	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatal(err)
	}

	// Write invalid CA
	if err := os.WriteFile(caFile, []byte("invalid ca"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := createTLSConfig(certFile, keyFile, caFile)
	if err == nil {
		t.Error("Expected error for invalid CA certificate, got nil")
	}
}

func TestCreateTLSConfig_Success(t *testing.T) {
	// Create temp cert, key, and CA files
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")
	caFile := filepath.Join(tempDir, "ca.pem")

	certPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)

	keyPEM := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)

	caPEM := certPEM // Use same cert as CA for testing

	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caFile, caPEM, 0600); err != nil {
		t.Fatal(err)
	}

	tlsConfig, err := createTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		t.Fatalf("createTLSConfig() error = %v", err)
	}

	if tlsConfig == nil {
		t.Fatal("Expected non-nil TLS config")
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("Expected 1 certificate, got %d", len(tlsConfig.Certificates))
	}

	if tlsConfig.RootCAs == nil {
		t.Error("Expected non-nil RootCAs")
	}

	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinVersion TLS12, got %d", tlsConfig.MinVersion)
	}

	// Verify CA pool contains certificate
	if tlsConfig.RootCAs != nil {
		cert, err := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
		if err != nil {
			t.Fatalf("Failed to parse certificate: %v", err)
		}
		if cert == nil {
			t.Error("Expected non-nil parsed certificate")
		}
	}
}

func TestNewKafkaProducer_InvalidBrokers(t *testing.T) {
	config := &Config{
		Brokers:    []string{"invalid-broker:9999999"},
		Topic:      "test",
		UseAuth:    false,
		KeySerde:   "json",
		ValueSerde: "json",
	}

	_, err := NewKafkaProducer(config)
	if err == nil {
		t.Error("Expected error for invalid broker, got nil")
	}
}

func TestNewKafkaProducer_WithInvalidTLS(t *testing.T) {
	config := &Config{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test",
		UseAuth:    true,
		CertFile:   "nonexistent.pem",
		KeyFile:    "nonexistent.key",
		CAFile:     "nonexistent.ca",
		KeySerde:   "json",
		ValueSerde: "json",
	}

	_, err := NewKafkaProducer(config)
	if err == nil {
		t.Error("Expected error for invalid TLS config, got nil")
	}
}

func TestKafkaProducer_Close(t *testing.T) {
	// Test closing nil producer
	producer := &KafkaProducer{
		producer: nil,
		config:   &Config{},
	}

	err := producer.Close()
	if err != nil {
		t.Errorf("Close() with nil producer error = %v, want nil", err)
	}
}

// Mock sync producer for testing
type mockSyncProducer struct {
	sendMessageFunc func(*sarama.ProducerMessage) (int32, int64, error)
	closeFunc       func() error
}

func (m *mockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(msg)
	}
	return 0, 0, nil
}

func (m *mockSyncProducer) SendMessages([]*sarama.ProducerMessage) error {
	return nil
}

func (m *mockSyncProducer) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// Additional methods required by sarama.SyncProducer interface
func (m *mockSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return 0
}

func (m *mockSyncProducer) IsTransactional() bool {
	return false
}

func (m *mockSyncProducer) BeginTxn() error {
	return nil
}

func (m *mockSyncProducer) CommitTxn() error {
	return nil
}

func (m *mockSyncProducer) AbortTxn() error {
	return nil
}

func (m *mockSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	return nil
}

func (m *mockSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	return nil
}

func TestKafkaProducer_SendMessage_Success(t *testing.T) {
	config := &Config{
		Topic:      "test-topic",
		KeySerde:   "string",
		ValueSerde: "string",
	}

	mockProducer := &mockSyncProducer{
		sendMessageFunc: func(msg *sarama.ProducerMessage) (int32, int64, error) {
			if msg.Topic != "test-topic" {
				t.Errorf("Expected topic test-topic, got %s", msg.Topic)
			}
			return 1, 100, nil
		},
	}

	producer := &KafkaProducer{
		producer: mockProducer,
		config:   config,
	}

	partition, offset, err := producer.SendMessage("test-key", "test-value")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	if partition != 1 {
		t.Errorf("Expected partition 1, got %d", partition)
	}

	if offset != 100 {
		t.Errorf("Expected offset 100, got %d", offset)
	}
}

func TestKafkaProducer_SendMessage_EmptyKey(t *testing.T) {
	config := &Config{
		Topic:      "test-topic",
		KeySerde:   "string",
		ValueSerde: "string",
	}

	mockProducer := &mockSyncProducer{
		sendMessageFunc: func(msg *sarama.ProducerMessage) (int32, int64, error) {
			if msg.Key != nil {
				t.Error("Expected nil key for empty string")
			}
			return 0, 0, nil
		},
	}

	producer := &KafkaProducer{
		producer: mockProducer,
		config:   config,
	}

	_, _, err := producer.SendMessage("", "test-value")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
}
