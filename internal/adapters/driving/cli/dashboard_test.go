package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestDashboardQuitKeys(t *testing.T) {
	for _, k := range []string{"q", "ctrl+c", "esc"} {
		t.Run(k, func(t *testing.T) {
			_, cmd := dashboardModel{}.Update(key(k))
			require.NotNil(t, cmd)
			require.IsType(t, tea.QuitMsg{}, cmd())
		})
	}
}

func TestDashboardRefreshKeyTriggersFetch(t *testing.T) {
	_, cmd := dashboardModel{refresh: dashboardRefresh}.Update(key("r"))
	require.NotNil(t, cmd)
}

func TestDashboardStoresReport(t *testing.T) {
	report := statusReport{
		Companies: []companyStatus{{
			Company: companyView{Name: "Demo", Slug: "demo", RepoOwner: "me", RepoName: "r"},
			Agents:  []agentView{{Name: "Ada", Role: "coder", Status: "working"}},
			States:  map[string]int{"queued": 1, "merged": 2},
		}},
		PendingCommands: []commandView{{Kind: "kill"}},
	}

	updated, _ := dashboardModel{}.Update(dashReportMsg{report: report})
	view := updated.View()

	require.Contains(t, view, "Demo (demo)  me/r")
	require.Contains(t, view, "merged 2")
	require.Contains(t, view, "Ada")
	require.Contains(t, view, "pending commands: 1")
}

func TestDashboardShowsError(t *testing.T) {
	updated, _ := dashboardModel{}.Update(dashErrMsg{err: errStub("boom")})
	require.Contains(t, updated.View(), "error: boom")
}

func TestDashboardEmptyView(t *testing.T) {
	require.Contains(t, dashboardModel{}.View(), "(no companies)")
}

type errStub string

func (e errStub) Error() string { return string(e) }
