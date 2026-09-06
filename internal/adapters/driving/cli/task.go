package cli

import (
	"fmt"
	"net/http"
	"net/url"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newTaskCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "Manage tasks"}
	cmd.AddCommand(newTaskCreateCmd(g), newTaskListCmd(g))
	return cmd
}

func newTaskCreateCmd(g *globalFlags) *cobra.Command {
	var (
		company     string
		title       string
		description string
		risk        string
		maxRetries  int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out taskView
			err = client.do(cmd.Context(), http.MethodPost, "/api/v1/tasks", createTaskBody{
				CompanyID:   company,
				Title:       title,
				Description: description,
				Risk:        risk,
				MaxRetries:  maxRetries,
			}, &out)
			if err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "created task %s (%s)\n", out.ID, out.State)
			})
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "company id (required)")
	cmd.Flags().StringVar(&title, "title", "", "task title (required)")
	cmd.Flags().StringVar(&description, "description", "", "task description")
	cmd.Flags().StringVar(&risk, "risk", "", "risk level: low, medium or high (required)")
	cmd.Flags().IntVar(&maxRetries, "max-retries", 0, "retry cap (0 uses the server default)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("risk")
	return cmd
}

func newTaskListCmd(g *globalFlags) *cobra.Command {
	var (
		company string
		state   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks (the API requires at least one filter)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			q := url.Values{}
			if company != "" {
				q.Set("company", company)
			}
			if state != "" {
				q.Set("state", state)
			}
			path := "/api/v1/tasks"
			if len(q) > 0 {
				path += "?" + q.Encode()
			}

			var out []taskView
			if err := client.do(cmd.Context(), http.MethodGet, path, nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() { printTaskTable(cmd, out) })
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "filter by company id")
	cmd.Flags().StringVar(&state, "state", "", "filter by lifecycle state")
	return cmd
}

func printTaskTable(cmd *cobra.Command, tasks []taskView) {
	if len(tasks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no tasks")
		return
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATE\tRISK\tRETRIES\tBRANCH")
	for _, t := range tasks {
		branch := t.Branch
		if branch == "" {
			branch = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			t.ID, t.Title, t.State, t.Risk, t.Retries, t.MaxRetries, branch)
	}
	_ = w.Flush()
}
