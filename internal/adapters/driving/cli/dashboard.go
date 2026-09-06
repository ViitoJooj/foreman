package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const dashboardRefresh = 2 * time.Second

func newDashboardCmd(g *globalFlags) *cobra.Command {
	var company string

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Live read-only dashboard of companies, tasks, agents and commands",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			program := tea.NewProgram(
				dashboardModel{client: client, company: company, refresh: dashboardRefresh},
				tea.WithContext(cmd.Context()),
				tea.WithAltScreen(),
			)
			_, err = program.Run()
			return err
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "limit the dashboard to one company id")
	return cmd
}

type dashReportMsg struct{ report statusReport }
type dashErrMsg struct{ err error }
type dashTickMsg struct{}

type dashboardModel struct {
	client  *apiClient
	company string
	refresh time.Duration

	report  statusReport
	err     error
	updated time.Time
}

func (m dashboardModel) Init() tea.Cmd {
	return tea.Batch(m.fetch(), m.scheduleTick())
}

func (m dashboardModel) scheduleTick() tea.Cmd {
	return tea.Tick(m.refresh, func(time.Time) tea.Msg { return dashTickMsg{} })
}

func (m dashboardModel) fetch() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.refresh+5*time.Second)
		defer cancel()

		report, err := fetchStatusReport(ctx, m.client, m.company)
		if err != nil {
			return dashErrMsg{err: err}
		}
		return dashReportMsg{report: report}
	}
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			return m, m.fetch()
		}
	case dashTickMsg:
		return m, tea.Batch(m.fetch(), m.scheduleTick())
	case dashReportMsg:
		m.report = msg.report
		m.err = nil
		m.updated = time.Now()
	case dashErrMsg:
		m.err = msg.err
	}
	return m, nil
}

func (m dashboardModel) View() string {
	var b strings.Builder
	b.WriteString("foreman dashboard\n\n")

	if m.err != nil {
		fmt.Fprintf(&b, "error: %v\n\n", m.err)
	}
	if len(m.report.Companies) == 0 && m.err == nil {
		b.WriteString("(no companies)\n")
	}

	for _, cs := range m.report.Companies {
		fmt.Fprintf(&b, "%s (%s)  %s/%s\n",
			cs.Company.Name, cs.Company.Slug, cs.Company.RepoOwner, cs.Company.RepoName)
		fmt.Fprintf(&b, "  tasks:  %s\n", formatStates(cs.States))
		for _, a := range cs.Agents {
			fmt.Fprintf(&b, "  agent:  %-24s %-12s %s\n", a.Name, a.Role, a.Status)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "pending commands: %d\n", len(m.report.PendingCommands))
	for _, c := range m.report.PendingCommands {
		fmt.Fprintf(&b, "  %-6s %s\n", c.Kind, c.ID)
	}

	updated := "never"
	if !m.updated.IsZero() {
		updated = m.updated.Format("15:04:05")
	}
	fmt.Fprintf(&b, "\nupdated %s  ·  refresh %s  ·  [r] refresh   [q] quit\n", updated, m.refresh)
	return b.String()
}
