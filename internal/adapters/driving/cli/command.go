package cli

import (
	"fmt"
	"net/http"
	"net/url"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// issueCommand posts a control action and prints the recorded command.
func issueCommand(cmd *cobra.Command, g *globalFlags, body issueCommandBody) error {
	client, err := g.client()
	if err != nil {
		return err
	}

	var out commandView
	if err := client.do(cmd.Context(), http.MethodPost, "/api/v1/commands", body, &out); err != nil {
		return err
	}
	return render(cmd, g, out, func() {
		fmt.Fprintf(cmd.OutOrStdout(), "issued %s command %s (%s)\n", out.Kind, out.ID, out.Status)
	})
}

func newKillCmd(g *globalFlags) *cobra.Command {
	var (
		panicMode bool
		company   string
		reason    string
	)

	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Record a graceful stop; add --panic to also close agent PRs and branches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			kind := "kill"
			if panicMode {
				kind = "panic"
			}
			return issueCommand(cmd, g, issueCommandBody{Kind: kind, CompanyID: company, Reason: reason})
		},
	}

	cmd.Flags().BoolVar(&panicMode, "panic", false, "panic mode: also close agent PRs and delete their branches")
	cmd.Flags().StringVar(&company, "company", "", "scope the stop to one company (default: all)")
	cmd.Flags().StringVar(&reason, "reason", "", "why the stop was issued")
	return cmd
}

func newPauseCmd(g *globalFlags) *cobra.Command {
	var agent, reason string

	cmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause a single agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return issueCommand(cmd, g, issueCommandBody{Kind: "pause", TargetAgentID: agent, Reason: reason})
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "", "agent id (required)")
	cmd.Flags().StringVar(&reason, "reason", "", "why the agent was paused")
	_ = cmd.MarkFlagRequired("agent")
	return cmd
}

func newResumeCmd(g *globalFlags) *cobra.Command {
	var agent, reason string

	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a paused agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return issueCommand(cmd, g, issueCommandBody{Kind: "resume", TargetAgentID: agent, Reason: reason})
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "", "agent id (required)")
	cmd.Flags().StringVar(&reason, "reason", "", "why the agent was resumed")
	_ = cmd.MarkFlagRequired("agent")
	return cmd
}

func newCommandCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "command", Short: "Inspect recorded control actions"}
	cmd.AddCommand(newCommandListCmd(g), newCommandGetCmd(g))
	return cmd
}

func newCommandListCmd(g *globalFlags) *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recorded commands (the API defaults to pending)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			path := "/api/v1/commands"
			if status != "" {
				path += "?" + url.Values{"status": {status}}.Encode()
			}
			var out []commandView
			if err := client.do(cmd.Context(), http.MethodGet, path, nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() { printCommandTable(cmd, out) })
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "filter by status: pending, applied or failed")
	return cmd
}

func newCommandGetCmd(g *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show one command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out commandView
			path := "/api/v1/commands/" + url.PathEscape(args[0])
			if err := client.do(cmd.Context(), http.MethodGet, path, nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", out.ID, out.Kind, out.Status)
			})
		},
	}
}

func printCommandTable(cmd *cobra.Command, commands []commandView) {
	if len(commands) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no commands")
		return
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tKIND\tSTATUS\tREASON")
	for _, c := range commands {
		reason := c.Reason
		if reason == "" {
			reason = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID, c.Kind, c.Status, reason)
	}
	_ = w.Flush()
}
