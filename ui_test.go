package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	config := &Config{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test-topic",
		CertFile:   "/cert.pem",
		KeyFile:    "/key.pem",
		CAFile:     "/ca.pem",
		UseAuth:    true,
		KeySerde:   "json",
		ValueSerde: "json",
	}

	m := initialModel(config)

	if m.currentView != configView {
		t.Errorf("Expected configView, got %v", m.currentView)
	}

	if len(m.configInputs) != 7 {
		t.Errorf("Expected 7 config inputs, got %d", len(m.configInputs))
	}

	if m.configFocus != 0 {
		t.Error("Expected configFocus to be 0")
	}

	if m.messageFocus != 0 {
		t.Error("Expected messageFocus to be 0")
	}

	if m.connected {
		t.Error("Expected connected to be false initially")
	}

	// Check that inputs are initialized with config values
	brokerValue := m.configInputs[brokerField].Value()
	if brokerValue != "localhost:9092" {
		t.Errorf("Expected broker localhost:9092, got %s", brokerValue)
	}

	topicValue := m.configInputs[topicField].Value()
	if topicValue != "test-topic" {
		t.Errorf("Expected topic test-topic, got %s", topicValue)
	}
}

func TestModel_Init(t *testing.T) {
	m := initialModel(&Config{})
	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init() to return nil")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := initialModel(&Config{})

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	updatedModel := newModel.(model)

	if updatedModel.width != 100 {
		t.Errorf("Expected width 100, got %d", updatedModel.width)
	}

	if updatedModel.height != 50 {
		t.Errorf("Expected height 50, got %d", updatedModel.height)
	}
}

func TestModel_Update_Tab(t *testing.T) {
	m := initialModel(&Config{})
	m.currentView = configView
	m.configFocus = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel := newModel.(model)

	if updatedModel.configFocus != 1 {
		t.Errorf("Expected configFocus 1 after tab, got %d", updatedModel.configFocus)
	}
}

func TestModel_Update_ShiftTab(t *testing.T) {
	m := initialModel(&Config{})
	m.currentView = configView
	m.configFocus = 1

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updatedModel := newModel.(model)

	if updatedModel.configFocus != 0 {
		t.Errorf("Expected configFocus 0 after shift+tab, got %d", updatedModel.configFocus)
	}
}

func TestModel_Update_ShiftTab_Wraparound(t *testing.T) {
	m := initialModel(&Config{})
	m.currentView = configView
	m.configFocus = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updatedModel := newModel.(model)

	expectedFocus := int(maxConfigField) - 1
	if updatedModel.configFocus != expectedFocus {
		t.Errorf("Expected configFocus %d after shift+tab wraparound, got %d", expectedFocus, updatedModel.configFocus)
	}
}

func TestModel_Update_F2_NotConnected(t *testing.T) {
	m := initialModel(&Config{})
	m.currentView = configView
	m.connected = false

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyF2})
	updatedModel := newModel.(model)

	if updatedModel.currentView != configView {
		t.Error("Should not switch to messageView when not connected")
	}

	if !strings.Contains(updatedModel.statusMessage, "connect") {
		t.Errorf("Expected connection warning in status, got: %s", updatedModel.statusMessage)
	}
}

func TestModel_Update_F2_Connected(t *testing.T) {
	m := initialModel(&Config{})
	m.currentView = configView
	m.connected = true

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyF2})
	updatedModel := newModel.(model)

	if updatedModel.currentView != messageView {
		t.Error("Should switch to messageView when connected")
	}
}

func TestModel_Update_F2_BackToConfig(t *testing.T) {
	m := initialModel(&Config{})
	m.currentView = messageView
	m.connected = true

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyF2})
	updatedModel := newModel.(model)

	if updatedModel.currentView != configView {
		t.Error("Should switch back to configView from messageView")
	}
}

func TestModel_Update_SuccessMsg(t *testing.T) {
	m := initialModel(&Config{})

	newModel, _ := m.Update(successMsg{msg: "Test success"})
	updatedModel := newModel.(model)

	if updatedModel.statusMessage != "Test success" {
		t.Errorf("Expected status 'Test success', got %s", updatedModel.statusMessage)
	}
}

func TestModel_Update_ErrMsg(t *testing.T) {
	m := initialModel(&Config{})

	testErr := errMsg{err: &testError{msg: "Test error"}}
	newModel, _ := m.Update(testErr)
	updatedModel := newModel.(model)

	if !strings.Contains(updatedModel.statusMessage, "Test error") {
		t.Errorf("Expected error in status, got %s", updatedModel.statusMessage)
	}
}

func TestModel_Update_ConnectSuccessMsg(t *testing.T) {
	m := initialModel(&Config{})
	m.connected = false

	mockProducer := &KafkaProducer{
		config: &Config{},
	}

	newModel, _ := m.Update(connectSuccessMsg{producer: mockProducer})
	updatedModel := newModel.(model)

	if !updatedModel.connected {
		t.Error("Expected connected to be true after connectSuccessMsg")
	}

	if updatedModel.producer != mockProducer {
		t.Error("Expected producer to be set")
	}

	if !strings.Contains(strings.ToLower(updatedModel.statusMessage), "connected") {
		t.Errorf("Expected connection success message, got %s", updatedModel.statusMessage)
	}
}

func TestModel_Update_MessageResult_Success(t *testing.T) {
	m := initialModel(&Config{})
	m.messageKeyInput.SetValue("test-key")
	m.messageValueArea.SetValue("test-value")

	result := messageResult{
		partition: 1,
		offset:    100,
		err:       nil,
	}

	newModel, _ := m.Update(result)
	updatedModel := newModel.(model)

	if len(updatedModel.messages) != 1 {
		t.Fatalf("Expected 1 message in history, got %d", len(updatedModel.messages))
	}

	msg := updatedModel.messages[0]
	if msg.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got %s", msg.Key)
	}

	if msg.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got %s", msg.Value)
	}

	if msg.Status != "Success" {
		t.Errorf("Expected status 'Success', got %s", msg.Status)
	}

	if msg.Partition != 1 {
		t.Errorf("Expected partition 1, got %d", msg.Partition)
	}

	if msg.Offset != 100 {
		t.Errorf("Expected offset 100, got %d", msg.Offset)
	}

	// Check that inputs were cleared
	if updatedModel.messageKeyInput.Value() != "" {
		t.Error("Expected key input to be cleared after success")
	}

	if updatedModel.messageValueArea.Value() != "" {
		t.Error("Expected value input to be cleared after success")
	}
}

func TestModel_Update_MessageResult_Error(t *testing.T) {
	m := initialModel(&Config{})
	m.messageKeyInput.SetValue("test-key")
	m.messageValueArea.SetValue("test-value")

	result := messageResult{
		partition: 0,
		offset:    0,
		err:       &testError{msg: "Send failed"},
	}

	newModel, _ := m.Update(result)
	updatedModel := newModel.(model)

	if len(updatedModel.messages) != 1 {
		t.Fatalf("Expected 1 message in history, got %d", len(updatedModel.messages))
	}

	msg := updatedModel.messages[0]
	if !strings.Contains(msg.Status, "Failed") {
		t.Errorf("Expected status to contain 'Failed', got %s", msg.Status)
	}

	// Check that inputs were NOT cleared on failure
	if updatedModel.messageKeyInput.Value() == "" {
		t.Error("Expected key input to remain after failure")
	}
}

func TestModel_View_Loading(t *testing.T) {
	m := initialModel(&Config{})
	m.width = 0

	view := m.View()

	if !strings.Contains(view, "Loading") {
		t.Error("Expected 'Loading...' message when width is 0")
	}
}

func TestModel_View_ConfigView(t *testing.T) {
	m := initialModel(&Config{Topic: "test-topic"})
	m.width = 100
	m.height = 50
	m.currentView = configView

	view := m.View()

	if !strings.Contains(view, "Configuration") {
		t.Error("Expected 'Configuration' in config view")
	}

	if !strings.Contains(view, "Brokers") {
		t.Error("Expected 'Brokers' field in config view")
	}
}

func TestModel_View_MessageView(t *testing.T) {
	config := &Config{Topic: "my-topic"}
	m := initialModel(config)
	m.width = 100
	m.height = 50
	m.currentView = messageView
	m.connected = true

	view := m.View()

	if !strings.Contains(view, "Send Message") {
		t.Error("Expected 'Send Message' in message view")
	}

	if !strings.Contains(view, "my-topic") {
		t.Error("Expected topic name in message view")
	}

	if !strings.Contains(view, "Message History") {
		t.Error("Expected 'Message History' in message view")
	}
}

func TestModel_View_Connected(t *testing.T) {
	m := initialModel(&Config{})
	m.width = 100
	m.height = 50
	m.connected = true

	view := m.View()

	if !strings.Contains(view, "Connected") {
		t.Error("Expected 'Connected' status when connected")
	}
}

func TestModel_View_NotConnected(t *testing.T) {
	m := initialModel(&Config{})
	m.width = 100
	m.height = 50
	m.connected = false

	view := m.View()

	if !strings.Contains(view, "Not connected") {
		t.Error("Expected 'Not connected' status when not connected")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"", 10, "(empty)"},
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a very long string", 10, "this is a ..."},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// Helper types for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestModel_MessageHistory_Display(t *testing.T) {
	m := initialModel(&Config{})
	m.width = 100
	m.height = 50
	m.currentView = messageView

	// Add test messages
	m.messages = []Message{
		{
			Timestamp: time.Now(),
			Key:       "key1",
			Value:     "value1",
			Status:    "Success",
			Partition: 0,
			Offset:    100,
		},
		{
			Timestamp: time.Now(),
			Key:       "key2",
			Value:     "value2",
			Status:    "Failed: timeout",
			Partition: 0,
			Offset:    0,
		},
	}

	view := m.View()

	if !strings.Contains(view, "key1") {
		t.Error("Expected key1 in message history")
	}

	if !strings.Contains(view, "key2") {
		t.Error("Expected key2 in message history")
	}

	if !strings.Contains(view, "SUCCESS") {
		t.Error("Expected SUCCESS status in history")
	}

	if !strings.Contains(view, "FAILED") {
		t.Error("Expected FAILED status in history")
	}
}

func TestModel_RenderConfigView_MTLSBadge(t *testing.T) {
	config := &Config{
		UseAuth: true,
	}
	m := initialModel(config)
	m.width = 100

	view := m.renderConfigView()

	if !strings.Contains(view, "mTLS Enabled") {
		t.Error("Expected 'mTLS Enabled' badge when UseAuth is true")
	}
}

func TestModel_TabNavigation_MessageView(t *testing.T) {
	m := initialModel(&Config{})
	m.currentView = messageView
	m.messageFocus = int(msgKeyField)

	// Tab from key to value
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel := newModel.(model)

	if updatedModel.messageFocus != int(msgValueField) {
		t.Errorf("Expected messageFocus to be msgValueField, got %d", updatedModel.messageFocus)
	}

	// Tab again should wrap to key
	newModel2, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedModel2 := newModel2.(model)

	if updatedModel2.messageFocus != int(msgKeyField) {
		t.Errorf("Expected messageFocus to wrap to msgKeyField, got %d", updatedModel2.messageFocus)
	}
}

func TestModel_EmptyMessageHistory(t *testing.T) {
	m := initialModel(&Config{})
	m.width = 100
	m.currentView = messageView

	view := m.View()

	if !strings.Contains(view, "No messages sent yet") {
		t.Error("Expected 'No messages sent yet' message when history is empty")
	}
}

func TestModel_MessageHistory_Limit(t *testing.T) {
	m := initialModel(&Config{})

	// Add more than 5 messages with unique keys
	for i := 0; i < 10; i++ {
		m.messages = append(m.messages, Message{
			Timestamp: time.Now(),
			Key:       "key-" + string(rune('0'+i)),
			Value:     "value",
			Status:    "Success",
			Partition: 0,
			Offset:    int64(i),
		})
	}

	m.width = 100
	m.currentView = messageView
	view := m.View()

	// Should only show last 5 messages (key-5 through key-9)
	if !strings.Contains(view, "key-5") {
		t.Error("Expected message 'key-5' in history (6th message)")
	}

	// Should not show first message
	if strings.Contains(view, "key-0") {
		t.Error("Should not show message 'key-0' (1st message) when more than 5 exist")
	}
}

// Test that the model properly initializes textinput and textarea
func TestModel_InputsInitialization(t *testing.T) {
	config := &Config{
		Brokers:    []string{"test:9092"},
		Topic:      "test-topic",
		KeySerde:   "json",
		ValueSerde: "json",
	}

	m := initialModel(config)

	// Check textinput initialization
	if m.messageKeyInput.Value() != "" {
		t.Error("Expected messageKeyInput to be empty initially")
	}

	// Check textarea initialization
	if m.messageValueArea.Value() != "" {
		t.Error("Expected messageValueArea to be empty initially")
	}

	// Verify they are proper types
	var _ textinput.Model = m.messageKeyInput
	var _ textarea.Model = m.messageValueArea
}

func TestModel_Connect_Success(t *testing.T) {
	homeDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	m := initialModel(&Config{
		Brokers:  []string{"localhost:9092"},
		Topic:    "test",
		UseAuth:  false,
		KeySerde: "json",
	})

	m.configInputs[brokerField].SetValue("localhost:9092")
	m.configInputs[topicField].SetValue("test-topic")

	cmd := m.connect()
	if cmd == nil {
		t.Fatal("Expected non-nil command from connect()")
	}

	// Execute the command
	msg := cmd()

	// Should return error since we can't actually connect
	if _, ok := msg.(errMsg); !ok {
		t.Errorf("Expected errMsg, got %T", msg)
	}
}

func TestModel_SaveConfig_Execute(t *testing.T) {
	homeDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	m := initialModel(&Config{
		Brokers:    []string{"localhost:9092"},
		Topic:      "test",
		KeySerde:   "json",
		ValueSerde: "json",
	})

	m.configInputs[brokerField].SetValue("broker1:9092,broker2:9092")
	m.configInputs[topicField].SetValue("my-topic")
	m.configInputs[keySerdeField].SetValue("string")
	m.configInputs[valueSerdeField].SetValue("bytearray")

	cmd := m.saveConfig()
	if cmd == nil {
		t.Fatal("Expected non-nil command from saveConfig()")
	}

	// Execute the command
	msg := cmd()

	// Should return success
	if successMsg, ok := msg.(successMsg); !ok {
		t.Errorf("Expected successMsg, got %T", msg)
	} else if !strings.Contains(successMsg.msg, "saved") {
		t.Errorf("Expected 'saved' in message, got %s", successMsg.msg)
	}
}

func TestModel_SendMessage_NotConnected(t *testing.T) {
	m := initialModel(&Config{})
	m.producer = nil

	m.messageKeyInput.SetValue("test-key")
	m.messageValueArea.SetValue("test-value")

	cmd := m.sendMessage()
	if cmd == nil {
		t.Fatal("Expected non-nil command from sendMessage()")
	}

	// Execute the command
	msg := cmd()

	// Should return error since not connected
	if errMsg, ok := msg.(errMsg); !ok {
		t.Errorf("Expected errMsg, got %T", msg)
	} else if !strings.Contains(errMsg.Error(), "not connected") {
		t.Errorf("Expected 'not connected' error, got %s", errMsg.Error())
	}
}

func TestModel_SendMessage_EmptyValue(t *testing.T) {
	m := initialModel(&Config{})
	m.producer = &KafkaProducer{config: &Config{}}

	m.messageKeyInput.SetValue("test-key")
	m.messageValueArea.SetValue("")

	cmd := m.sendMessage()
	if cmd == nil {
		t.Fatal("Expected non-nil command from sendMessage()")
	}

	// Execute the command
	msg := cmd()

	// Should return error for empty value
	if errMsg, ok := msg.(errMsg); !ok {
		t.Errorf("Expected errMsg, got %T", msg)
	} else if !strings.Contains(errMsg.Error(), "empty") {
		t.Errorf("Expected 'empty' error, got %s", errMsg.Error())
	}
}

func TestModel_FormatJSON_ValidJSON(t *testing.T) {
	m := initialModel(&Config{})
	m.messageValueArea.SetValue(`{"key":"value","number":123}`)

	cmd := m.formatJSON()
	if cmd == nil {
		t.Fatal("Expected non-nil command from formatJSON()")
	}

	// Execute the command
	msg := cmd()

	// Should return success
	if successMsg, ok := msg.(successMsg); !ok {
		t.Errorf("Expected successMsg, got %T", msg)
	} else if !strings.Contains(successMsg.msg, "formatted") {
		t.Errorf("Expected 'formatted' in message, got %s", successMsg.msg)
	}

	// Check that value was formatted with indentation
	formatted := m.messageValueArea.Value()
	if !strings.Contains(formatted, "\n") {
		t.Error("Expected formatted JSON to contain newlines")
	}

	if !strings.Contains(formatted, "  ") {
		t.Error("Expected formatted JSON to contain indentation")
	}
}

func TestModel_FormatJSON_InvalidJSON(t *testing.T) {
	m := initialModel(&Config{})
	m.messageValueArea.SetValue("not valid json")

	cmd := m.formatJSON()
	if cmd == nil {
		t.Fatal("Expected non-nil command from formatJSON()")
	}

	// Execute the command
	msg := cmd()

	// Should return error for invalid JSON
	if errMsg, ok := msg.(errMsg); !ok {
		t.Errorf("Expected errMsg, got %T", msg)
	} else if !strings.Contains(errMsg.Error(), "JSON") {
		t.Errorf("Expected 'JSON' error, got %s", errMsg.Error())
	}
}

func TestModel_FormatJSON_Empty(t *testing.T) {
	m := initialModel(&Config{})
	m.messageValueArea.SetValue("")

	cmd := m.formatJSON()
	if cmd == nil {
		t.Fatal("Expected non-nil command from formatJSON()")
	}

	// Execute the command
	msg := cmd()

	// Should return success with "Nothing to format"
	if successMsg, ok := msg.(successMsg); !ok {
		t.Errorf("Expected successMsg, got %T", msg)
	} else if !strings.Contains(successMsg.msg, "Nothing") {
		t.Errorf("Expected 'Nothing' in message, got %s", successMsg.msg)
	}
}

func TestErrMsg_Error(t *testing.T) {
	err := errMsg{err: &testError{msg: "test error"}}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}
