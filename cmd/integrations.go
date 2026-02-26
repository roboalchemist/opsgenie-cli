package cmd

import (
	"fmt"
	"strconv"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(integrationsCmd)
	integrationsCmd.AddCommand(integrationsListCmd)
	integrationsCmd.AddCommand(integrationsGetCmd)
	integrationsCmd.AddCommand(integrationsCreateCmd)
	integrationsCmd.AddCommand(integrationsUpdateCmd)
	integrationsCmd.AddCommand(integrationsDeleteCmd)
	integrationsCmd.AddCommand(integrationsEnableCmd)
	integrationsCmd.AddCommand(integrationsDisableCmd)

	addOutputFlags(integrationsListCmd)
	addOutputFlags(integrationsGetCmd)
	addOutputFlags(integrationsCreateCmd)
	addOutputFlags(integrationsUpdateCmd)

	// create flags
	integrationsCreateCmd.Flags().String("name", "", "Integration name (required)")
	integrationsCreateCmd.Flags().String("type", "", "Integration type (required)")
	integrationsCreateCmd.Flags().Bool("enabled", true, "Whether integration is enabled")
	_ = integrationsCreateCmd.MarkFlagRequired("name")
	_ = integrationsCreateCmd.MarkFlagRequired("type")

	// update flags
	integrationsUpdateCmd.Flags().String("name", "", "Integration name")
	integrationsUpdateCmd.Flags().String("type", "", "Integration type")
	integrationsUpdateCmd.Flags().Bool("enabled", true, "Whether integration is enabled")
}

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Manage OpsGenie integrations",
	Long:  "Create, list, and manage OpsGenie integrations.",
}

var integrationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []api.IntegrationResponse `json:"data"`
		}
		if err := client.Get("/v2/integrations", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "NAME", "TYPE", "ENABLED"}
		rows := make([][]string, 0, len(resp.Data))
		for _, i := range resp.Data {
			rows = append(rows, []string{
				i.ID,
				i.Name,
				i.Type,
				strconv.FormatBool(i.Enabled),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var integrationsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an integration by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data api.IntegrationResponse `json:"data"`
		}
		if err := client.Get("/v2/integrations/"+args[0], &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", resp.Data.ID},
			{"Name", resp.Data.Name},
			{"Type", resp.Data.Type},
			{"Enabled", strconv.FormatBool(resp.Data.Enabled)},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var integrationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new integration",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		intType, _ := cmd.Flags().GetString("type")
		enabled, _ := cmd.Flags().GetBool("enabled")

		body := map[string]interface{}{
			"name":    name,
			"type":    intType,
			"enabled": enabled,
		}

		var result map[string]interface{}
		if err := client.Post("/v2/integrations", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Integration %q created", name), opts)
		return output.RenderJSON(result, opts)
	},
}

var integrationsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an integration",
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
		if err := client.Put("/v2/integrations/"+args[0], body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Integration %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var integrationsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an integration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v2/integrations/"+args[0], nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Integration %q deleted", args[0]), opts)
		return nil
	},
}

var integrationsEnableCmd = &cobra.Command{
	Use:   "enable <id>",
	Short: "Enable an integration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Post("/v2/integrations/"+args[0]+"/enable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Integration %q enabled", args[0]), opts)
		return nil
	},
}

var integrationsDisableCmd = &cobra.Command{
	Use:   "disable <id>",
	Short: "Disable an integration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Post("/v2/integrations/"+args[0]+"/disable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Integration %q disabled", args[0]), opts)
		return nil
	},
}
