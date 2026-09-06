package cli

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/spf13/cobra"
)

const (
	defaultAPIURL = "http://localhost:8080"
	envAPIURL     = "FOREMAN_API_URL"
	envAPIKey     = "FOREMAN_API_KEY"
)

// globalFlags holds the options shared by every subcommand.
type globalFlags struct {
	apiURL string
	apiKey string
	json   bool
}

// NewRootCmd builds the `foreman` command tree.
func NewRootCmd() *cobra.Command {
	g := &globalFlags{}

	root := &cobra.Command{
		Use:           "foreman",
		Short:         "Client for the autodev harness API",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&g.apiURL, "api-url", envOr(envAPIURL, defaultAPIURL), "base URL of the harness API")
	root.PersistentFlags().StringVar(&g.apiKey, "api-key", os.Getenv(envAPIKey), "API key sent as the X-Api-Key header")
	root.PersistentFlags().BoolVar(&g.json, "json", false, "print the raw JSON response")

	root.AddCommand(
		newStatusCmd(g),
		newDashboardCmd(g),
		newCompanyCmd(g),
		newAgentCmd(g),
		newChannelCmd(g),
		newTaskCmd(g),
		newMessageCmd(g),
		newCommandCmd(g),
		newKillCmd(g),
		newPauseCmd(g),
		newResumeCmd(g),
	)
	return root
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (g *globalFlags) client() (*apiClient, error) {
	if g.apiKey == "" {
		return nil, errors.New("api key required: pass --api-key or set $" + envAPIKey)
	}
	return newAPIClient(g.apiURL, g.apiKey), nil
}

// render writes v as indented JSON when --json is set, otherwise runs text.
func render(cmd *cobra.Command, g *globalFlags, v any, text func()) error {
	if g.json {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	text()
	return nil
}
