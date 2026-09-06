package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// statusReport is the aggregate shape rendered by `foreman status --json`.
type statusReport struct {
	Companies       []companyStatus `json:"companies"`
	PendingCommands []commandView   `json:"pending_commands"`
}

type companyStatus struct {
	Company companyView    `json:"company"`
	Tasks   []taskView     `json:"tasks"`
	Agents  []agentView    `json:"agents"`
	States  map[string]int `json:"task_states"`
}

func newStatusCmd(g *globalFlags) *cobra.Command {
	var company string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Snapshot of companies, tasks, agents and pending commands",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			report, err := fetchStatusReport(cmd.Context(), client, company)
			if err != nil {
				return err
			}
			return render(cmd, g, report, func() { printStatus(cmd, report) })
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "limit the snapshot to one company id")
	return cmd
}

// fetchStatusReport gathers the snapshot used by `status` and the dashboard.
func fetchStatusReport(ctx context.Context, client *apiClient, companyFilter string) (statusReport, error) {
	companies, err := statusCompanies(ctx, client, companyFilter)
	if err != nil {
		return statusReport{}, err
	}

	report := statusReport{}
	for _, c := range companies {
		cs := companyStatus{Company: c, States: map[string]int{}}

		q := url.Values{"company": {c.ID}}.Encode()
		if err := client.do(ctx, http.MethodGet, "/api/v1/tasks?"+q, nil, &cs.Tasks); err != nil {
			return statusReport{}, err
		}
		if err := client.do(ctx, http.MethodGet, "/api/v1/agents?"+q, nil, &cs.Agents); err != nil {
			return statusReport{}, err
		}
		for _, t := range cs.Tasks {
			cs.States[t.State]++
		}
		report.Companies = append(report.Companies, cs)
	}

	if err := client.do(ctx, http.MethodGet, "/api/v1/commands?status=pending", nil, &report.PendingCommands); err != nil {
		return statusReport{}, err
	}
	return report, nil
}

func statusCompanies(ctx context.Context, client *apiClient, companyID string) ([]companyView, error) {
	if companyID != "" {
		var one companyView
		if err := client.do(ctx, http.MethodGet, "/api/v1/companies/"+url.PathEscape(companyID), nil, &one); err != nil {
			return nil, err
		}
		return []companyView{one}, nil
	}
	var all []companyView
	err := client.do(ctx, http.MethodGet, "/api/v1/companies", nil, &all)
	return all, err
}

func printStatus(cmd *cobra.Command, report statusReport) {
	out := cmd.OutOrStdout()
	if len(report.Companies) == 0 {
		fmt.Fprintln(out, "no companies")
	}

	for _, cs := range report.Companies {
		fmt.Fprintf(out, "%s (%s)  %s/%s\n", cs.Company.Name, cs.Company.Slug, cs.Company.RepoOwner, cs.Company.RepoName)
		fmt.Fprintf(out, "  tasks: %s\n", formatStates(cs.States))

		w := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
		for _, a := range cs.Agents {
			fmt.Fprintf(w, "  agent\t%s\t%s\t%s\n", a.Name, a.Role, a.Status)
		}
		_ = w.Flush()
	}

	fmt.Fprintf(out, "pending commands: %d\n", len(report.PendingCommands))
	for _, c := range report.PendingCommands {
		reason := c.Reason
		if reason == "" {
			reason = "-"
		}
		fmt.Fprintf(out, "  %s  %-6s %s\n", c.ID, c.Kind, reason)
	}
}

func formatStates(states map[string]int) string {
	if len(states) == 0 {
		return "(none)"
	}
	keys := make([]string, 0, len(states))
	for k := range states {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, states[k]))
	}
	return strings.Join(parts, "  ")
}
