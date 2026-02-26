package cmd

import (
	"fmt"

	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(policiesCmd)
	policiesCmd.AddCommand(policiesListCmd)
	policiesCmd.AddCommand(policiesGetCmd)
	policiesCmd.AddCommand(policiesCreateCmd)
	policiesCmd.AddCommand(policiesUpdateCmd)
	policiesCmd.AddCommand(policiesDeleteCmd)
	policiesCmd.AddCommand(policiesEnableCmd)
	policiesCmd.AddCommand(policiesDisableCmd)

	addOutputFlags(policiesListCmd)
	addOutputFlags(policiesGetCmd)
	addOutputFlags(policiesCreateCmd)
	addOutputFlags(policiesUpdateCmd)

	// create flags
	policiesCreateCmd.Flags().String("name", "", "Policy name (required)")
	policiesCreateCmd.Flags().String("type", "alert", "Policy type (alert, notification)")
	policiesCreateCmd.Flags().Bool("enabled", true, "Whether policy is enabled")
	_ = policiesCreateCmd.MarkFlagRequired("name")

	// update flags
	policiesUpdateCmd.Flags().String("name", "", "Policy name")
	policiesUpdateCmd.Flags().String("type", "", "Policy type")
	policiesUpdateCmd.Flags().Bool("enabled", true, "Whether policy is enabled")
}

var policiesCmd = &cobra.Command{
	Use:   "policies",
	Short: "Manage OpsGenie alert and notification policies",
	Long:  "Create, list, and manage alert and notification policies. Note: v1 endpoints, deprecated but functional.",
}

var policiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v1/policies", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "NAME", "TYPE", "ENABLED"}
		rows := make([][]string, 0, len(resp.Data))
		for _, p := range resp.Data {
			rows = append(rows, []string{
				stringVal(p, "id"),
				stringVal(p, "name"),
				stringVal(p, "type"),
				stringVal(p, "enabled"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var policiesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a policy by ID",
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
		if err := client.Get("/v1/policies/"+args[0], &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"Name", stringVal(resp.Data, "name")},
			{"Type", stringVal(resp.Data, "type")},
			{"Enabled", stringVal(resp.Data, "enabled")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var policiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		pType, _ := cmd.Flags().GetString("type")
		enabled, _ := cmd.Flags().GetBool("enabled")

		body := map[string]interface{}{
			"name":    name,
			"type":    pType,
			"enabled": enabled,
		}

		var result map[string]interface{}
		if err := client.Post("/v1/policies", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Policy %q created", name), opts)
		return output.RenderJSON(result, opts)
	},
}

var policiesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		body := map[string]interface{}{}
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		if cmd.Flags().Changed("type") {
			v, _ := cmd.Flags().GetString("type")
			body["type"] = v
		}
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			body["enabled"] = v
		}

		var result map[string]interface{}
		if err := client.Put("/v1/policies/"+args[0], body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Policy %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var policiesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v1/policies/"+args[0], nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Policy %q deleted", args[0]), opts)
		return nil
	},
}

var policiesEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable a policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Post("/v1/policies/"+args[0]+"/enable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Policy %q enabled", args[0]), opts)
		return nil
	},
}

var policiesDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable a policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Post("/v1/policies/"+args[0]+"/disable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Policy %q disabled", args[0]), opts)
		return nil
	},
}
