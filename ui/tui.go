package ui

import (
	"fmt"
	"time"

	"otto/models"

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
	status    string
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
		status:    "Checking prerequisites...",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, checkPrerequisites)
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
			m.status = "Refreshing server status..."
			return m, loadData
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case prerequisitesCheckedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
			return m, tea.Quit
		}
		m.status = "Loading server status..."
		return m, loadData
	case dataLoadedMsg:
		m.isLoading = false
		updateTableData(&m, msg.servers)
		m.status = "Server status loaded. Press 'r' to refresh, 'q' to quit."
		return m, nil
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.isLoading {
		return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.status)
	}
	return baseStyle.Render(m.table.View()) + "\n\n" + m.status + "\n"
}

func updateTableData(m *model, servers []*models.Server) {
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
	servers []*models.Server
}

type prerequisitesCheckedMsg struct {
	err error
}

// Add these functions to handle prerequisites and data loading
func checkPrerequisites() tea.Msg {
	// Implement prerequisite checks here
	// For now, we'll just simulate a delay
	time.Sleep(1 * time.Second)
	return prerequisitesCheckedMsg{err: nil}
}

func loadData() tea.Msg {
	// Implement data loading here
	// For now, we'll just simulate a delay and return empty data
	time.Sleep(1 * time.Second)
	return dataLoadedMsg{servers: []*models.Server{}}
}

// RunTUI starts the Terminal User Interface
func RunTUI() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
