package cli

import (
	"fmt"
	"net/http"
	"net/url"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newCompanyCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "company", Short: "Manage companies"}
	cmd.AddCommand(newCompanyCreateCmd(g), newCompanyListCmd(g), newCompanyGetCmd(g))
	return cmd
}

func newCompanyCreateCmd(g *globalFlags) *cobra.Command {
	var name, slug, repoOwner, repoName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a company",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out companyView
			err = client.do(cmd.Context(), http.MethodPost, "/api/v1/companies", createCompanyBody{
				Name:      name,
				Slug:      slug,
				RepoOwner: repoOwner,
				RepoName:  repoName,
			}, &out)
			if err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "created company %s (%s)\n", out.ID, out.Slug)
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name (required)")
	cmd.Flags().StringVar(&slug, "slug", "", "url-safe slug (required)")
	cmd.Flags().StringVar(&repoOwner, "repo-owner", "", "GitHub owner login (required)")
	cmd.Flags().StringVar(&repoName, "repo-name", "", "GitHub repository name (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("slug")
	_ = cmd.MarkFlagRequired("repo-owner")
	_ = cmd.MarkFlagRequired("repo-name")
	return cmd
}

func newCompanyListCmd(g *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List companies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out []companyView
			if err := client.do(cmd.Context(), http.MethodGet, "/api/v1/companies", nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() { printCompanyTable(cmd, out) })
		},
	}
}

func newCompanyGetCmd(g *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show one company",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out companyView
			path := "/api/v1/companies/" + url.PathEscape(args[0])
			if err := client.do(cmd.Context(), http.MethodGet, path, nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s/%s\n", out.ID, out.Name, out.RepoOwner, out.RepoName)
			})
		},
	}
}

func printCompanyTable(cmd *cobra.Command, companies []companyView) {
	if len(companies) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no companies")
		return
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSLUG\tREPO")
	for _, c := range companies {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s/%s\n", c.ID, c.Name, c.Slug, c.RepoOwner, c.RepoName)
	}
	_ = w.Flush()
}
