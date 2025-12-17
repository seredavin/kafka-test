package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	configFocus     configField
	messageFocus    messageField
	inputs          map[string]string
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

type messageResult struct {
	partition int32
	offset    int64
	err       error
}

// initialModel creates the initial model
func initialModel(config *Config) model {
	inputs := make(map[string]string)
	inputs["broker"] = strings.Join(config.Brokers, ",")
	inputs["topic"] = config.Topic
	inputs["cert"] = config.CertFile
	inputs["key"] = config.KeyFile
	inputs["ca"] = config.CAFile
	inputs["msgkey"] = ""
	inputs["msgvalue"] = ""

	return model{
		config:       config,
		currentView:  configView,
		configFocus:  brokerField,
		messageFocus: msgKeyField,
		inputs:       inputs,
		messages:     []Message{},
		connected:    false,
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
				m.configFocus = (m.configFocus + 1) % maxConfigField
			} else {
				m.messageFocus = (m.messageFocus + 1) % maxMessageField
			}
			return m, nil

		case "shift+tab":
			if m.currentView == configView {
				if m.configFocus == 0 {
					m.configFocus = maxConfigField - 1
				} else {
					m.configFocus--
				}
			} else {
				if m.messageFocus == 0 {
					m.messageFocus = maxMessageField - 1
				} else {
					m.messageFocus--
				}
			}
			return m, nil

		case "f2":
			// Toggle between views
			if m.currentView == configView {
				if m.connected {
					m.currentView = messageView
				} else {
					m.statusMessage = "Please connect to Kafka first (F5)"
				}
			} else {
				m.currentView = configView
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
			if m.currentView == messageView && m.messageFocus == msgValueField {
				return m, m.formatJSON()
			}
			return m, nil

		case "backspace":
			field := m.getCurrentField()
			if len(m.inputs[field]) > 0 {
				m.inputs[field] = m.inputs[field][:len(m.inputs[field])-1]
			}
			return m, nil

		default:
			// Handle text input
			if len(msg.String()) == 1 {
				field := m.getCurrentField()
				m.inputs[field] += msg.String()
			}
			return m, nil
		}

	case successMsg:
		m.statusMessage = msg.msg
		return m, nil

	case errMsg:
		m.err = msg.err
		m.statusMessage = fmt.Sprintf("Error: %v", msg.err)
		return m, nil

	case messageResult:
		if msg.err != nil {
			m.messages = append(m.messages, Message{
				Timestamp: time.Now(),
				Key:       m.inputs["msgkey"],
				Value:     m.inputs["msgvalue"],
				Status:    fmt.Sprintf("Failed: %v", msg.err),
			})
		} else {
			m.messages = append(m.messages, Message{
				Timestamp: time.Now(),
				Key:       m.inputs["msgkey"],
				Value:     m.inputs["msgvalue"],
				Status:    "Success",
				Partition: msg.partition,
				Offset:    msg.offset,
			})
			// Clear message fields after successful send
			m.inputs["msgkey"] = ""
			m.inputs["msgvalue"] = ""
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

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	status := m.statusMessage
	if status == "" {
		if m.connected {
			status = "Connected to Kafka"
		} else {
			status = "Not connected"
		}
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	help := "F2: Switch View | F5: Connect | F9: Save Config | F10: Format JSON | Enter: Send | Esc: Quit"

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		statusStyle.Render(status),
		helpStyle.Render(help),
	)
}

func (m model) renderConfigView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Padding(1, 0)

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(60)

	focusedInputStyle := inputStyle.Copy().
		Background(lipgloss.Color("238")).
		Foreground(lipgloss.Color("229"))

	title := titleStyle.Render("Kafka Producer Configuration")

	fields := []struct {
		label string
		key   string
		field configField
	}{
		{"Brokers (comma-separated)", "broker", brokerField},
		{"Topic", "topic", topicField},
		{"Client Certificate Path", "cert", certField},
		{"Client Key Path", "key", keyField},
		{"CA Certificate Path", "ca", caField},
	}

	var rows []string
	rows = append(rows, title)
	rows = append(rows, "")

	for _, f := range fields {
		label := fieldStyle.Render(f.label + ":")
		if m.configFocus == f.field {
			label = focusedStyle.Render(f.label + ":")
		}

		input := m.inputs[f.key]
		if input == "" {
			input = " "
		}

		var inputBox string
		if m.configFocus == f.field {
			inputBox = focusedInputStyle.Render(input + "▌")
		} else {
			inputBox = inputStyle.Render(input)
		}

		rows = append(rows, label)
		rows = append(rows, inputBox)
		rows = append(rows, "")
	}

	useAuthLabel := "Use mTLS Authentication: "
	useAuthValue := "NO"
	if m.config.UseAuth || (m.inputs["cert"] != "" && m.inputs["key"] != "" && m.inputs["ca"] != "") {
		useAuthValue = "YES"
	}
	rows = append(rows, fieldStyle.Render(useAuthLabel+useAuthValue))

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m model) renderMessageView() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Padding(1, 0)

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(70)

	focusedInputStyle := inputStyle.Copy().
		Background(lipgloss.Color("238")).
		Foreground(lipgloss.Color("229"))

	valueInputStyle := inputStyle.Copy().Height(5)
	focusedValueInputStyle := focusedInputStyle.Copy().Height(5)

	title := titleStyle.Render(fmt.Sprintf("Send Message to Topic: %s", m.config.Topic))

	var rows []string
	rows = append(rows, title)
	rows = append(rows, "")

	// Key field
	keyLabel := fieldStyle.Render("Message Key (optional):")
	if m.messageFocus == msgKeyField {
		keyLabel = focusedStyle.Render("Message Key (optional):")
	}

	keyInput := m.inputs["msgkey"]
	if keyInput == "" {
		keyInput = " "
	}

	var keyBox string
	if m.messageFocus == msgKeyField {
		keyBox = focusedInputStyle.Render(keyInput + "▌")
	} else {
		keyBox = inputStyle.Render(keyInput)
	}

	rows = append(rows, keyLabel)
	rows = append(rows, keyBox)
	rows = append(rows, "")

	// Value field
	valueLabel := fieldStyle.Render("Message Value:")
	if m.messageFocus == msgValueField {
		valueLabel = focusedStyle.Render("Message Value:")
	}

	valueInput := m.inputs["msgvalue"]
	if valueInput == "" {
		valueInput = " "
	}

	var valueBox string
	if m.messageFocus == msgValueField {
		valueBox = focusedValueInputStyle.Render(valueInput + "▌")
	} else {
		valueBox = valueInputStyle.Render(valueInput)
	}

	rows = append(rows, valueLabel)
	rows = append(rows, valueBox)
	rows = append(rows, "")

	// Message history
	historyTitle := titleStyle.Render("Message History:")
	rows = append(rows, historyTitle)

	if len(m.messages) == 0 {
		rows = append(rows, fieldStyle.Render("No messages sent yet"))
	} else {
		// Show last 5 messages
		start := 0
		if len(m.messages) > 5 {
			start = len(m.messages) - 5
		}

		for i := start; i < len(m.messages); i++ {
			msg := m.messages[i]
			statusColor := "34"
			if strings.HasPrefix(msg.Status, "Failed") {
				statusColor = "196"
			}

			msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
			msgStr := fmt.Sprintf("[%s] Key: %s | Status: %s",
				msg.Timestamp.Format("15:04:05"),
				truncate(msg.Key, 20),
				msg.Status)

			if msg.Status == "Success" {
				msgStr += fmt.Sprintf(" | Partition: %d, Offset: %d", msg.Partition, msg.Offset)
			}

			rows = append(rows, msgStyle.Render(msgStr))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m model) getCurrentField() string {
	if m.currentView == configView {
		switch m.configFocus {
		case brokerField:
			return "broker"
		case topicField:
			return "topic"
		case certField:
			return "cert"
		case keyField:
			return "key"
		case caField:
			return "ca"
		}
	} else {
		switch m.messageFocus {
		case msgKeyField:
			return "msgkey"
		case msgValueField:
			return "msgvalue"
		}
	}
	return ""
}

func (m *model) connect() tea.Cmd {
	return func() tea.Msg {
		// Close existing producer
		if m.producer != nil {
			m.producer.Close()
		}

		// Update config from inputs
		brokers := strings.Split(m.inputs["broker"], ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}

		m.config.Brokers = brokers
		m.config.Topic = m.inputs["topic"]
		m.config.CertFile = m.inputs["cert"]
		m.config.KeyFile = m.inputs["key"]
		m.config.CAFile = m.inputs["ca"]

		// Enable mTLS if certificates are provided
		m.config.UseAuth = m.inputs["cert"] != "" && m.inputs["key"] != "" && m.inputs["ca"] != ""

		// Create new producer
		producer, err := NewKafkaProducer(m.config)
		if err != nil {
			return errMsg{err}
		}

		m.producer = producer
		m.connected = true

		return successMsg{"Successfully connected to Kafka"}
	}
}

func (m *model) saveConfig() tea.Cmd {
	return func() tea.Msg {
		// Update config from inputs
		brokers := strings.Split(m.inputs["broker"], ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}

		m.config.Brokers = brokers
		m.config.Topic = m.inputs["topic"]
		m.config.CertFile = m.inputs["cert"]
		m.config.KeyFile = m.inputs["key"]
		m.config.CAFile = m.inputs["ca"]
		m.config.UseAuth = m.inputs["cert"] != "" && m.inputs["key"] != "" && m.inputs["ca"] != ""

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

		key := m.inputs["msgkey"]
		value := m.inputs["msgvalue"]

		if value == "" {
			return errMsg{fmt.Errorf("message value cannot be empty")}
		}

		partition, offset, err := m.producer.SendMessage(key, value)
		return messageResult{partition, offset, err}
	}
}

func (m *model) formatJSON() tea.Cmd {
	return func() tea.Msg {
		value := m.inputs["msgvalue"]
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

		m.inputs["msgvalue"] = string(formatted)
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
