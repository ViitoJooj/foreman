package cli

import (
	"fmt"
	"net/http"
	"net/url"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newAgentCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "agent", Short: "Manage agents"}
	cmd.AddCommand(newAgentCreateCmd(g), newAgentListCmd(g), newAgentGetCmd(g))
	return cmd
}

func newAgentCreateCmd(g *globalFlags) *cobra.Command {
	var company, name, role string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add an agent to a company",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out agentView
			err = client.do(cmd.Context(), http.MethodPost, "/api/v1/agents", createAgentBody{
				CompanyID: company,
				Name:      name,
				Role:      role,
			}, &out)
			if err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "created agent %s (%s, %s)\n", out.ID, out.Role, out.Status)
			})
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "company id (required)")
	cmd.Flags().StringVar(&name, "name", "", "agent name (required)")
	cmd.Flags().StringVar(&role, "role", "", "role: task_creator, coder, tester or pr_reviewer (required)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("role")
	return cmd
}

func newAgentListCmd(g *globalFlags) *cobra.Command {
	var company string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a company's agents",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			q := url.Values{"company": {company}}
			var out []agentView
			if err := client.do(cmd.Context(), http.MethodGet, "/api/v1/agents?"+q.Encode(), nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() { printAgentTable(cmd, out) })
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "company id (required)")
	_ = cmd.MarkFlagRequired("company")
	return cmd
}

func newAgentGetCmd(g *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show one agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out agentView
			path := "/api/v1/agents/" + url.PathEscape(args[0])
			if err := client.do(cmd.Context(), http.MethodGet, path, nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s  %s\n", out.ID, out.Name, out.Role, out.Status)
			})
		},
	}
}

func printAgentTable(cmd *cobra.Command, agents []agentView) {
	if len(agents) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no agents")
		return
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tROLE\tSTATUS")
	for _, a := range agents {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.ID, a.Name, a.Role, a.Status)
	}
	_ = w.Flush()
}
