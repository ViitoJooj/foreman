package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

func newMessageCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "message", Short: "Post channel messages"}
	cmd.AddCommand(newMessagePostCmd(g))
	return cmd
}

func newMessagePostCmd(g *globalFlags) *cobra.Command {
	var (
		channel          string
		msgType          string
		from             string
		to               string
		inReplyTo        string
		body             string
		payload          string
		requiresResponse bool
	)

	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post a message to a channel",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := g.client()
			if err != nil {
				return err
			}

			reqBody := postMessageBody{
				Type:             msgType,
				FromAgentID:      from,
				ToAgentID:        to,
				InReplyTo:        inReplyTo,
				RequiresResponse: requiresResponse,
				Body:             body,
			}
			if payload != "" {
				if !json.Valid([]byte(payload)) {
					return errors.New("--payload is not valid JSON")
				}
				reqBody.Payload = json.RawMessage(payload)
			}

			path := "/api/v1/channels/" + url.PathEscape(channel) + "/messages"
			var out messageView
			if err := client.do(cmd.Context(), http.MethodPost, path, reqBody, &out); err != nil {
				return err
			}
			return render(cmd, g, out, func() {
				fmt.Fprintf(cmd.OutOrStdout(), "posted %s message %s to channel %s\n", out.Type, out.ID, out.ChannelID)
			})
		},
	}

	cmd.Flags().StringVar(&channel, "channel", "", "channel id (required)")
	cmd.Flags().StringVar(&msgType, "type", "", "message type: request, response or status (required)")
	cmd.Flags().StringVar(&from, "from", "", "sending agent id (required)")
	cmd.Flags().StringVar(&to, "to", "", "recipient agent id")
	cmd.Flags().StringVar(&inReplyTo, "in-reply-to", "", "id of the message this replies to")
	cmd.Flags().StringVar(&body, "body", "", "message body")
	cmd.Flags().StringVar(&payload, "payload", "", "JSON payload")
	cmd.Flags().BoolVar(&requiresResponse, "requires-response", false, "mark that a response is expected")
	_ = cmd.MarkFlagRequired("channel")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("from")
	return cmd
}
