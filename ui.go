package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View mode
type viewMode int

const (
	configView viewMode = iota
	messageView
)

// Input field index for config view
type configField int

const (
	brokerField configField = iota
	topicField
	certField
	keyField
	caField
	keySerdeField
	valueSerdeField
	maxConfigField
)

// Input field index for message view
type messageField int

const (
	msgKeyField messageField = iota
	msgValueField
	maxMessageField
)

// Message represents a sent message with status
type Message struct {
	Timestamp time.Time
	Key       string
	Value     string
	Status    string
	Partition int32
	Offset    int64
}

// Model holds the application state
type model struct {
	config          *Config
	producer        *KafkaProducer
	currentView     viewMode
	configInputs    []textinput.Model
	messageKeyInput textinput.Model
	messageValueArea textarea.Model
	configFocus     int
	messageFocus    int
	messages        []Message
	statusMessage   string
	connected       bool
	width           int
	height          int
	err             error
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type successMsg struct{ msg string }

type connectSuccessMsg struct {
	producer *KafkaProducer
}

type messageResult struct {
	partition int32
	offset    int64
	err       error
}

// initialModel creates the initial model
func initialModel(config *Config) model {
	// Create config input fields
	configInputs := make([]textinput.Model, 7)

	// Broker input
	configInputs[brokerField] = textinput.New()
	configInputs[brokerField].Placeholder = "localhost:9092"
	configInputs[brokerField].SetValue(strings.Join(config.Brokers, ","))
	configInputs[brokerField].Focus()
	configInputs[brokerField].Width = 60

	// Topic input
	configInputs[topicField] = textinput.New()
	configInputs[topicField].Placeholder = "my-topic"
	configInputs[topicField].SetValue(config.Topic)
	configInputs[topicField].Width = 60

	// Cert input
	configInputs[certField] = textinput.New()
	configInputs[certField].Placeholder = "/path/to/cert.pem"
	configInputs[certField].SetValue(config.CertFile)
	configInputs[certField].Width = 60

	// Key input
	configInputs[keyField] = textinput.New()
	configInputs[keyField].Placeholder = "/path/to/key.pem"
	configInputs[keyField].SetValue(config.KeyFile)
	configInputs[keyField].Width = 60

	// CA input
	configInputs[caField] = textinput.New()
	configInputs[caField].Placeholder = "/path/to/ca.pem"
	configInputs[caField].SetValue(config.CAFile)
	configInputs[caField].Width = 60

	// Key Serde input
	configInputs[keySerdeField] = textinput.New()
	configInputs[keySerdeField].Placeholder = "string, json, bytearray"
	if config.KeySerde != "" {
		configInputs[keySerdeField].SetValue(config.KeySerde)
	} else {
		configInputs[keySerdeField].SetValue("json")
	}
	configInputs[keySerdeField].Width = 60

	// Value Serde input
	configInputs[valueSerdeField] = textinput.New()
	configInputs[valueSerdeField].Placeholder = "string, json, bytearray"
	if config.ValueSerde != "" {
		configInputs[valueSerdeField].SetValue(config.ValueSerde)
	} else {
		configInputs[valueSerdeField].SetValue("json")
	}
	configInputs[valueSerdeField].Width = 60

	// Create message key input
	messageKeyInput := textinput.New()
	messageKeyInput.Placeholder = "optional-key"
	messageKeyInput.Width = 100

	// Create message value textarea
	messageValueArea := textarea.New()
	messageValueArea.Placeholder = `{"example": "json"}`
	messageValueArea.SetWidth(120)
	messageValueArea.SetHeight(8)
	messageValueArea.CharLimit = 0

	return model{
		config:           config,
		currentView:      configView,
		configInputs:     configInputs,
		messageKeyInput:  messageKeyInput,
		messageValueArea: messageValueArea,
		configFocus:      0,
		messageFocus:     0,
		messages:         []Message{},
		connected:        false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.producer != nil {
				m.producer.Close()
			}
			return m, tea.Quit

		case "tab":
			if m.currentView == configView {
				m.configInputs[m.configFocus].Blur()
				m.configFocus = (m.configFocus + 1) % int(maxConfigField)
				m.configInputs[m.configFocus].Focus()
			} else {
				if m.messageFocus == int(msgKeyField) {
					m.messageKeyInput.Blur()
					m.messageFocus = int(msgValueField)
					m.messageValueArea.Focus()
				} else {
					m.messageValueArea.Blur()
					m.messageFocus = int(msgKeyField)
					m.messageKeyInput.Focus()
				}
			}
			return m, nil

		case "shift+tab":
			if m.currentView == configView {
				m.configInputs[m.configFocus].Blur()
				if m.configFocus == 0 {
					m.configFocus = int(maxConfigField) - 1
				} else {
					m.configFocus--
				}
				m.configInputs[m.configFocus].Focus()
			} else {
				if m.messageFocus == int(msgValueField) {
					m.messageValueArea.Blur()
					m.messageFocus = int(msgKeyField)
					m.messageKeyInput.Focus()
				} else {
					m.messageKeyInput.Blur()
					m.messageFocus = int(msgValueField)
					m.messageValueArea.Focus()
				}
			}
			return m, nil

		case "f2":
			// Toggle between views
			if m.currentView == configView {
				if m.connected {
					m.configInputs[m.configFocus].Blur()
					m.currentView = messageView
					if m.messageFocus == int(msgKeyField) {
						m.messageKeyInput.Focus()
					} else {
						m.messageValueArea.Focus()
					}
				} else {
					m.statusMessage = "Please connect to Kafka first (F5)"
				}
			} else {
				if m.messageFocus == int(msgKeyField) {
					m.messageKeyInput.Blur()
				} else {
					m.messageValueArea.Blur()
				}
				m.currentView = configView
				m.configInputs[m.configFocus].Focus()
			}
			return m, nil

		case "f5":
			// Connect/Reconnect to Kafka
			return m, m.connect()

		case "f9":
			// Save config
			return m, m.saveConfig()

		case "enter":
			if m.currentView == messageView {
				// Send message
				return m, m.sendMessage()
			}
			return m, nil

		case "f10":
			// Format JSON in message value field
			if m.currentView == messageView && m.messageFocus == int(msgValueField) {
				return m, m.formatJSON()
			}
			return m, nil
		}

		// Delegate to textinput/textarea for handling
		var cmd tea.Cmd
		if m.currentView == configView {
			m.configInputs[m.configFocus], cmd = m.configInputs[m.configFocus].Update(msg)
		} else {
			if m.messageFocus == int(msgKeyField) {
				m.messageKeyInput, cmd = m.messageKeyInput.Update(msg)
			} else {
				m.messageValueArea, cmd = m.messageValueArea.Update(msg)
			}
		}
		return m, cmd

	case successMsg:
		m.statusMessage = msg.msg
		return m, nil

	case connectSuccessMsg:
		m.producer = msg.producer
		m.connected = true
		m.statusMessage = "Successfully connected to Kafka"
		return m, nil

	case errMsg:
		m.err = msg.err
		m.statusMessage = fmt.Sprintf("Error: %v", msg.err)
		return m, nil

	case messageResult:
		if msg.err != nil {
			m.messages = append(m.messages, Message{
				Timestamp: time.Now(),
				Key:       m.messageKeyInput.Value(),
				Value:     m.messageValueArea.Value(),
				Status:    fmt.Sprintf("Failed: %v", msg.err),
			})
		} else {
			m.messages = append(m.messages, Message{
				Timestamp: time.Now(),
				Key:       m.messageKeyInput.Value(),
				Value:     m.messageValueArea.Value(),
				Status:    "Success",
				Partition: msg.partition,
				Offset:    msg.offset,
			})
			// Clear message fields after successful send
			m.messageKeyInput.SetValue("")
			m.messageValueArea.SetValue("")
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string
	if m.currentView == configView {
		content = m.renderConfigView()
	} else {
		content = m.renderMessageView()
	}

	// Adaptive status bar
	var statusStyle lipgloss.Style
	if m.connected {
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FAFAFA"}).
			Background(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"}).
			Bold(true).
			Padding(0, 2)
	} else {
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FAFAFA"}).
			Background(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}).
			Padding(0, 2)
	}

	status := m.statusMessage
	if status == "" {
		if m.connected {
			status = "● Connected to Kafka"
		} else {
			status = "○ Not connected"
		}
	}

	// Adaptive help bar
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}).
		Background(lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#1F2937"}).
		Padding(0, 2)

	help := "󰌌 F2: Switch │ 󰛐 F5: Connect │ 󰆓 F9: Save │ 󰉢 F10: Format │  Enter: Send │ 󰩈 Esc: Quit"

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		"",
		statusStyle.Render(status),
		helpStyle.Render(help),
	)
}

func (m model) renderConfigView() string {
	// Adaptive title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}).
		Background(lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#1F2937"}).
		Padding(1, 2).
		MarginBottom(1)

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}).
		MarginTop(1)

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}).
		Bold(true).
		MarginTop(1)

	title := titleStyle.Render("⚡ Kafka Producer Configuration")

	fields := []struct {
		label string
		icon  string
		field configField
	}{
		{"Brokers (comma-separated)", "󰒋", brokerField},
		{"Topic", "󰏫", topicField},
		{"Client Certificate Path", "󰄤", certField},
		{"Client Key Path", "󰌆", keyField},
		{"CA Certificate Path", "󰷛", caField},
		{"Key Serde", "󰘦", keySerdeField},
		{"Value Serde", "󰘦", valueSerdeField},
	}

	var rows []string
	rows = append(rows, title)
	rows = append(rows, "")

	for _, f := range fields {
		var label string
		if m.configFocus == int(f.field) {
			label = focusedStyle.Render(f.icon + " " + f.label + " ›")
		} else {
			label = fieldStyle.Render(f.icon + " " + f.label + ":")
		}

		rows = append(rows, label)
		rows = append(rows, m.configInputs[f.field].View())
	}

	// Adaptive mTLS status badge
	certVal := m.configInputs[certField].Value()
	keyVal := m.configInputs[keyField].Value()
	caVal := m.configInputs[caField].Value()

	var authBadge string
	if m.config.UseAuth || (certVal != "" && keyVal != "" && caVal != "") {
		authBadgeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#10B981"}).
			Background(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#064E3B"}).
			Padding(0, 1).
			MarginTop(1)
		authBadge = authBadgeStyle.Render("🔒 mTLS Enabled")
	} else {
		authBadgeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#6B7280"}).
			Background(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#1F2937"}).
			Padding(0, 1).
			MarginTop(1)
		authBadge = authBadgeStyle.Render("🔓 mTLS Disabled")
	}

	rows = append(rows, "")
	rows = append(rows, authBadge)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m model) renderMessageView() string {
	// Adaptive title with topic badge
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}).
		Background(lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#1F2937"}).
		Padding(1, 2).
		MarginBottom(1)

	topicBadge := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FBBF24"}).
		Background(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#451A03"}).
		Padding(0, 1).
		Bold(true).
		Render(m.config.Topic)

	title := titleStyle.Render("󰭻 Send Message") + " " + topicBadge

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}).
		MarginTop(1)

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}).
		Bold(true).
		MarginTop(1)

	var rows []string
	rows = append(rows, title)
	rows = append(rows, "")

	// Key field
	var keyLabel string
	if m.messageFocus == int(msgKeyField) {
		keyLabel = focusedStyle.Render("󰌆 Message Key (optional) ›")
	} else {
		keyLabel = fieldStyle.Render("󰌆 Message Key (optional):")
	}

	rows = append(rows, keyLabel)
	rows = append(rows, m.messageKeyInput.View())

	// Value field
	var valueLabel string
	if m.messageFocus == int(msgValueField) {
		valueLabel = focusedStyle.Render("󰗀 Message Value (JSON) ›")
	} else {
		valueLabel = fieldStyle.Render("󰗀 Message Value (JSON):")
	}

	rows = append(rows, valueLabel)
	rows = append(rows, m.messageValueArea.View())
	rows = append(rows, "")

	// Adaptive message history section
	historyHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}).
		Background(lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#1F2937"}).
		Padding(0, 2).
		MarginTop(1).
		MarginBottom(1)

	rows = append(rows, historyHeaderStyle.Render("󰋼 Message History"))

	if len(m.messages) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}).
			Italic(true)
		rows = append(rows, emptyStyle.Render("  No messages sent yet"))
	} else {
		// Show last 5 messages
		start := 0
		if len(m.messages) > 5 {
			start = len(m.messages) - 5
		}

		for i := start; i < len(m.messages); i++ {
			msg := m.messages[i]

			var statusBadge string
			var msgStyle lipgloss.Style

			if strings.HasPrefix(msg.Status, "Failed") {
				statusBadge = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#EF4444"}).
					Background(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#450A0A"}).
					Padding(0, 1).
					Render("✗ FAILED")
				msgStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#FCA5A5"})
			} else {
				statusBadge = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#10B981"}).
					Background(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#064E3B"}).
					Padding(0, 1).
					Render("✓ SUCCESS")
				msgStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#6EE7B7"})
			}

			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"})
			keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"})

			msgStr := fmt.Sprintf("  %s %s │ Key: %s",
				timeStyle.Render(msg.Timestamp.Format("15:04:05")),
				statusBadge,
				keyStyle.Render(truncate(msg.Key, 20)))

			if msg.Status == "Success" {
				partitionStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"})
				msgStr += partitionStyle.Render(fmt.Sprintf(" │ P:%d O:%d", msg.Partition, msg.Offset))
			}

			rows = append(rows, msgStyle.Render(msgStr))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *model) connect() tea.Cmd {
	return func() tea.Msg {
		// Close existing producer
		if m.producer != nil {
			m.producer.Close()
		}

		// Update config from inputs
		brokers := strings.Split(m.configInputs[brokerField].Value(), ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}

		m.config.Brokers = brokers
		m.config.Topic = m.configInputs[topicField].Value()
		m.config.CertFile = m.configInputs[certField].Value()
		m.config.KeyFile = m.configInputs[keyField].Value()
		m.config.CAFile = m.configInputs[caField].Value()
		m.config.KeySerde = m.configInputs[keySerdeField].Value()
		m.config.ValueSerde = m.configInputs[valueSerdeField].Value()

		// Enable mTLS if certificates are provided
		m.config.UseAuth = m.configInputs[certField].Value() != "" &&
			m.configInputs[keyField].Value() != "" &&
			m.configInputs[caField].Value() != ""

		// Create new producer
		producer, err := NewKafkaProducer(m.config)
		if err != nil {
			return errMsg{err}
		}

		return connectSuccessMsg{producer: producer}
	}
}

func (m *model) saveConfig() tea.Cmd {
	return func() tea.Msg {
		// Update config from inputs
		brokers := strings.Split(m.configInputs[brokerField].Value(), ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}

		m.config.Brokers = brokers
		m.config.Topic = m.configInputs[topicField].Value()
		m.config.CertFile = m.configInputs[certField].Value()
		m.config.KeyFile = m.configInputs[keyField].Value()
		m.config.CAFile = m.configInputs[caField].Value()
		m.config.KeySerde = m.configInputs[keySerdeField].Value()
		m.config.ValueSerde = m.configInputs[valueSerdeField].Value()
		m.config.UseAuth = m.configInputs[certField].Value() != "" &&
			m.configInputs[keyField].Value() != "" &&
			m.configInputs[caField].Value() != ""

		if err := SaveConfig(m.config); err != nil {
			return errMsg{err}
		}

		return successMsg{"Configuration saved successfully"}
	}
}

func (m *model) sendMessage() tea.Cmd {
	return func() tea.Msg {
		if m.producer == nil {
			return errMsg{fmt.Errorf("not connected to Kafka")}
		}

		key := m.messageKeyInput.Value()
		value := m.messageValueArea.Value()

		if value == "" {
			return errMsg{fmt.Errorf("message value cannot be empty")}
		}

		partition, offset, err := m.producer.SendMessage(key, value)
		return messageResult{partition, offset, err}
	}
}

func (m *model) formatJSON() tea.Cmd {
	return func() tea.Msg {
		value := m.messageValueArea.Value()
		if value == "" {
			return successMsg{"Nothing to format"}
		}

		var obj interface{}
		if err := json.Unmarshal([]byte(value), &obj); err != nil {
			return errMsg{fmt.Errorf("invalid JSON: %w", err)}
		}

		formatted, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return errMsg{err}
		}

		m.messageValueArea.SetValue(string(formatted))
		return successMsg{"JSON formatted successfully"}
	}
}

func truncate(s string, maxLen int) string {
	if s == "" {
		return "(empty)"
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
