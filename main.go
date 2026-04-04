package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/robmerrell/eldritch/internal/app"
)

func main() {
	p := tea.NewProgram(app.Init())

	// debug logging
	f, err := tea.LogToFile("/tmp/eld_debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting Eldritch: %v", err)
		os.Exit(1)
	}
}
