package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev" // will be replaced by build ldflags

func main() {
	// Handle command-line flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("kafka-producer-ui version %s\n", version)
			os.Exit(0)
		case "--help", "-h":
			fmt.Println("Kafka Producer UI - Terminal UI for Apache Kafka")
			fmt.Println("\nUsage:")
			fmt.Println("  kafka-producer-ui          Start the interactive UI")
			fmt.Println("  kafka-producer-ui --version Show version")
			fmt.Println("  kafka-producer-ui --help    Show this help")
			fmt.Println("\nConfiguration file: ~/.kafka-producer.json")
			fmt.Println("Documentation: https://github.com/seredavin/kafka-test")
			os.Exit(0)
		}
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create Bubble Tea program
	p := tea.NewProgram(
		initialModel(config),
		tea.WithAltScreen(),
		tea.WithInputTTY(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
