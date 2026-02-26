package cmd

import (
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(notificationRulesCmd)
	notificationRulesCmd.AddCommand(notificationRulesListCmd)
	notificationRulesCmd.AddCommand(notificationRulesGetCmd)
	notificationRulesCmd.AddCommand(notificationRulesCreateCmd)
	notificationRulesCmd.AddCommand(notificationRulesUpdateCmd)
	notificationRulesCmd.AddCommand(notificationRulesDeleteCmd)
	notificationRulesCmd.AddCommand(notificationRulesEnableCmd)
	notificationRulesCmd.AddCommand(notificationRulesDisableCmd)

	addOutputFlags(notificationRulesListCmd)
	addOutputFlags(notificationRulesGetCmd)
	addOutputFlags(notificationRulesCreateCmd)
	addOutputFlags(notificationRulesUpdateCmd)

	// shared user flag
	for _, c := range []*cobra.Command{
		notificationRulesListCmd, notificationRulesGetCmd, notificationRulesCreateCmd,
		notificationRulesUpdateCmd, notificationRulesDeleteCmd,
		notificationRulesEnableCmd, notificationRulesDisableCmd,
	} {
		c.Flags().String("user", "", "User ID or username (required)")
		_ = c.MarkFlagRequired("user")
	}

	// subcommands that need rule ID
	for _, c := range []*cobra.Command{
		notificationRulesGetCmd, notificationRulesUpdateCmd, notificationRulesDeleteCmd,
		notificationRulesEnableCmd, notificationRulesDisableCmd,
	} {
		c.Flags().String("rule-id", "", "Notification rule ID (required)")
		_ = c.MarkFlagRequired("rule-id")
	}

	// create flags
	notificationRulesCreateCmd.Flags().String("name", "", "Rule name (required)")
	notificationRulesCreateCmd.Flags().String("action-type", "", "Action type (required, e.g. create-alert)")
	notificationRulesCreateCmd.Flags().Bool("enabled", true, "Whether rule is enabled")
	_ = notificationRulesCreateCmd.MarkFlagRequired("name")
	_ = notificationRulesCreateCmd.MarkFlagRequired("action-type")

	// update flags
	notificationRulesUpdateCmd.Flags().String("name", "", "Rule name")
	notificationRulesUpdateCmd.Flags().Bool("enabled", true, "Whether rule is enabled")
}

var notificationRulesCmd = &cobra.Command{
	Use:   "notification-rules",
	Short: "Manage OpsGenie notification rules",
	Long:  "Create, list, and manage notification rules for OpsGenie users.",
}

var notificationRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification rules for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v2/users/"+userID+"/notification-rules", &resp); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "NAME", "ACTION_TYPE", "ENABLED"}
		rows := make([][]string, 0, len(resp.Data))
		for _, r := range resp.Data {
			rows = append(rows, []string{
				stringVal(r, "id"),
				stringVal(r, "name"),
				stringVal(r, "actionType"),
				stringVal(r, "enabled"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var notificationRulesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a notification rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")
		ruleID, _ := cmd.Flags().GetString("rule-id")

		var resp struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v2/users/"+userID+"/notification-rules/"+ruleID, &resp); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"Name", stringVal(resp.Data, "name")},
			{"ActionType", stringVal(resp.Data, "actionType")},
			{"Enabled", stringVal(resp.Data, "enabled")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var notificationRulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a notification rule for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")
		name, _ := cmd.Flags().GetString("name")
		actionType, _ := cmd.Flags().GetString("action-type")
		enabled, _ := cmd.Flags().GetBool("enabled")

		body := map[string]interface{}{
			"name":       name,
			"actionType": actionType,
			"enabled":    enabled,
		}

		var result map[string]interface{}
		if err := client.Post("/v2/users/"+userID+"/notification-rules", body, &result); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Notification rule %q created for user %q", name, userID), opts)
		return output.RenderJSON(result, opts)
	},
}

var notificationRulesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a notification rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")
		ruleID, _ := cmd.Flags().GetString("rule-id")

		body := map[string]interface{}{}
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			body["enabled"] = v
		}

		var result map[string]interface{}
		if err := client.Patch("/v2/users/"+userID+"/notification-rules/"+ruleID, body, &result); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Notification rule %q updated for user %q", ruleID, userID), opts)
		return output.RenderJSON(result, opts)
	},
}

var notificationRulesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a notification rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		userID, _ := cmd.Flags().GetString("user")
		ruleID, _ := cmd.Flags().GetString("rule-id")

		if err := client.Delete("/v2/users/"+userID+"/notification-rules/"+ruleID, nil); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Notification rule %q deleted for user %q", ruleID, userID), opts)
		return nil
	},
}

var notificationRulesEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a notification rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		userID, _ := cmd.Flags().GetString("user")
		ruleID, _ := cmd.Flags().GetString("rule-id")

		if err := client.Post("/v2/users/"+userID+"/notification-rules/"+ruleID+"/enable", nil, nil); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Notification rule %q enabled for user %q", ruleID, userID), opts)
		return nil
	},
}

var notificationRulesDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a notification rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		userID, _ := cmd.Flags().GetString("user")
		ruleID, _ := cmd.Flags().GetString("rule-id")

		if err := client.Post("/v2/users/"+userID+"/notification-rules/"+ruleID+"/disable", nil, nil); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Notification rule %q disabled for user %q", ruleID, userID), opts)
		return nil
	},
}
