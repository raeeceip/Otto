#!/bin/bash

# Upgrade Otto Load Balancer with enhanced Bubble Tea TUI

# Install dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss

# Create new files for TUI
cat << EOF > tui.go
package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	healthyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	unhealthyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
)

type model struct {
	table     table.Model
	spinner   spinner.Model
	isLoading bool
}

func initialModel() model {
	columns := []table.Column{
		{Title: "Server", Width: 30},
		{Title: "Status", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		table:     t,
		spinner:   sp,
		isLoading: true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadData)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.isLoading = true
			return m, loadData
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case dataLoadedMsg:
		m.isLoading = false
		updateTableData(&m, msg.servers)
		return m, nil
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.isLoading {
		return fmt.Sprintf("\n\n   %s Loading server status...\n\n", m.spinner.View())
	}
	return baseStyle.Render(m.table.View()) + "\n\nPress 'r' to refresh, 'q' to quit\n"
}

func updateTableData(m *model, servers []*Server) {
	var rows []table.Row
	for _, server := range servers {
		status := healthyStyle.Render("✓ Healthy")
		if !server.IsHealthy {
			status = unhealthyStyle.Render("✗ Down")
		}
		rows = append(rows, table.Row{server.URL.String(), status})
	}
	m.table.SetRows(rows)
}

type dataLoadedMsg struct {
	servers []*Server
}

func loadData() tea.Msg {
	time.Sleep(2 * time.Second) // Simulate loading time
	return dataLoadedMsg{servers: servers}
}

func runTUI() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
EOF

# Modify main.go to incorporate TUI
awk '
/^import \(/ { print; print "\t\"github.com/charmbracelet/bubbletea\""; next }
/^func main\(\) {/ { print; print "\tgo runTUI()"; next }
/log.Println\("Starting Otto load balancer on port", config.Port\)/ {
    print "\tgo func() {"
    print "\t\tif err := http.ListenAndServe(config.Port, nil); err != nil {"
    print "\t\t\tlog.Fatalf(\"Error starting load balancer: %s\\n\", err.Error())"
    print "\t\t}"
    print "\t}()"
    next
}
/err = http.ListenAndServe\(config.Port, nil\)/ { next }
/if err != nil {/ { next }
/log.Fatalf\("Error starting load balancer: %s\\n", err.Error\(\)\)/ { next }
/}/ && !p { p=1; next }
{ print }
' otto.go > main.go.tmp && mv main.go.tmp otto.go

echo "Otto has been upgraded with an enhanced Bubble Tea TUI. Run 'go run .' to start the application."