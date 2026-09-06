package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newChannelCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "channel", Short: "Manage channels"}
	cmd.AddCommand(newChannelCreateCmd(g), newChannelListCmd(g), newChannelGetCmd(g), newChannelTailCmd(g))
	return cmd
}

func newChannelCreateCmd(g *globalFlags) *cobra.Command {
	var company, name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Open a channel in a company",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out channelView
			err = client.do(cmd.Context(), http.MethodPost, "/api/v1/channels", createChannelBody{
				CompanyID: company,
				Name:      name,
			}, &out)
			if err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "created channel %s (%s)\n", out.ID, out.Name)
			})
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "company id (required)")
	cmd.Flags().StringVar(&name, "name", "", "channel name (required)")
	_ = cmd.MarkFlagRequired("company")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newChannelListCmd(g *globalFlags) *cobra.Command {
	var company string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a company's channels",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			q := url.Values{"company": {company}}
			var out []channelView
			if err := client.do(cmd.Context(), http.MethodGet, "/api/v1/channels?"+q.Encode(), nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() { printChannelTable(cmd, out) })
		},
	}

	cmd.Flags().StringVar(&company, "company", "", "company id (required)")
	_ = cmd.MarkFlagRequired("company")
	return cmd
}

func newChannelGetCmd(g *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show one channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			var out channelView
			path := "/api/v1/channels/" + url.PathEscape(args[0])
			if err := client.do(cmd.Context(), http.MethodGet, path, nil, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", out.ID, out.Name)
			})
		},
	}
}

func newChannelTailCmd(g *globalFlags) *cobra.Command {
	var (
		channel string
		history int
	)

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Follow a channel's messages in real time",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/api/v1/channels/%s/messages/stream?history=%d",
				url.PathEscape(channel), history)
			return client.stream(cmd.Context(), path, func(data []byte) error {
				printTailLine(cmd, data)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&channel, "channel", "", "channel id (required)")
	cmd.Flags().IntVar(&history, "history", 20, "backlog messages to print before the live feed")
	_ = cmd.MarkFlagRequired("channel")
	return cmd
}

func printTailLine(cmd *cobra.Command, data []byte) {
	var m messageView
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", data)
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s  %-8s %s\n", m.CreatedAt.Format("15:04:05"), m.Type, m.Body)
}

func printChannelTable(cmd *cobra.Command, channels []channelView) {
	if len(channels) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no channels")
		return
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCOMPANY")
	for _, ch := range channels {
		fmt.Fprintf(w, "%s\t%s\t%s\n", ch.ID, ch.Name, ch.CompanyID)
	}
	_ = w.Flush()
}
