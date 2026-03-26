package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/robmerrell/eldritch/internal/app"
)

func main() {
	p := tea.NewProgram(app.Init())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting Eldritch: %v", err)
		os.Exit(1)
	}
}
