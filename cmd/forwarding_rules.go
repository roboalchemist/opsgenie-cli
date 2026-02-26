package cmd

import (
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(forwardingRulesCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesListCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesGetCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesCreateCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesUpdateCmd)
	forwardingRulesCmd.AddCommand(forwardingRulesDeleteCmd)

	addOutputFlags(forwardingRulesListCmd)
	addOutputFlags(forwardingRulesGetCmd)
	addOutputFlags(forwardingRulesCreateCmd)
	addOutputFlags(forwardingRulesUpdateCmd)

	// create flags
	forwardingRulesCreateCmd.Flags().String("from-user", "", "Username to forward from (required)")
	forwardingRulesCreateCmd.Flags().String("to-user", "", "Username to forward to (required)")
	forwardingRulesCreateCmd.Flags().String("start-date", "", "Start date (RFC3339)")
	forwardingRulesCreateCmd.Flags().String("end-date", "", "End date (RFC3339)")
	_ = forwardingRulesCreateCmd.MarkFlagRequired("from-user")
	_ = forwardingRulesCreateCmd.MarkFlagRequired("to-user")

	// update flags
	forwardingRulesUpdateCmd.Flags().String("from-user", "", "Username to forward from")
	forwardingRulesUpdateCmd.Flags().String("to-user", "", "Username to forward to")
	forwardingRulesUpdateCmd.Flags().String("start-date", "", "Start date (RFC3339)")
	forwardingRulesUpdateCmd.Flags().String("end-date", "", "End date (RFC3339)")
}

var forwardingRulesCmd = &cobra.Command{
	Use:   "forwarding-rules",
	Short: "Manage OpsGenie forwarding rules",
	Long:  "Create, list, and manage notification forwarding rules between users.",
}

var forwardingRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all forwarding rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v2/forwarding-rules", &resp); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "FROM_USER", "TO_USER", "START_DATE", "END_DATE"}
		rows := make([][]string, 0, len(resp.Data))
		for _, r := range resp.Data {
			fromUser := nestedStringVal(r, "fromUser", "username")
			toUser := nestedStringVal(r, "toUser", "username")
			rows = append(rows, []string{
				stringVal(r, "id"),
				fromUser,
				toUser,
				stringVal(r, "startDate"),
				stringVal(r, "endDate"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var forwardingRulesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a forwarding rule by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v2/forwarding-rules/"+args[0], &resp); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"FromUser", nestedStringVal(resp.Data, "fromUser", "username")},
			{"ToUser", nestedStringVal(resp.Data, "toUser", "username")},
			{"StartDate", stringVal(resp.Data, "startDate")},
			{"EndDate", stringVal(resp.Data, "endDate")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var forwardingRulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a forwarding rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		fromUser, _ := cmd.Flags().GetString("from-user")
		toUser, _ := cmd.Flags().GetString("to-user")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")

		body := map[string]interface{}{
			"fromUser":  map[string]string{"username": fromUser},
			"toUser":    map[string]string{"username": toUser},
			"startDate": startDate,
			"endDate":   endDate,
		}

		var result map[string]interface{}
		if err := client.Post("/v2/forwarding-rules", body, &result); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Forwarding rule from %q to %q created", fromUser, toUser), opts)
		return output.RenderJSON(result, opts)
	},
}

var forwardingRulesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		body := map[string]interface{}{}
		if cmd.Flags().Changed("from-user") {
			v, _ := cmd.Flags().GetString("from-user")
			body["fromUser"] = map[string]string{"username": v}
		}
		if cmd.Flags().Changed("to-user") {
			v, _ := cmd.Flags().GetString("to-user")
			body["toUser"] = map[string]string{"username": v}
		}
		if cmd.Flags().Changed("start-date") {
			v, _ := cmd.Flags().GetString("start-date")
			body["startDate"] = v
		}
		if cmd.Flags().Changed("end-date") {
			v, _ := cmd.Flags().GetString("end-date")
			body["endDate"] = v
		}

		var result map[string]interface{}
		if err := client.Put("/v2/forwarding-rules/"+args[0], body, &result); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Forwarding rule %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var forwardingRulesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v2/forwarding-rules/"+args[0], nil); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Forwarding rule %q deleted", args[0]), opts)
		return nil
	},
}

// nestedStringVal extracts a string from a nested map[string]interface{}.
func nestedStringVal(m map[string]interface{}, outerKey, innerKey string) string {
	if v, ok := m[outerKey]; ok {
		if inner, ok := v.(map[string]interface{}); ok {
			return stringVal(inner, innerKey)
		}
	}
	return ""
}
