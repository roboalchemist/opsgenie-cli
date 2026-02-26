package cmd

import (
	"fmt"
	"strings"

	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(customRolesCmd)
	customRolesCmd.AddCommand(customRolesListCmd)
	customRolesCmd.AddCommand(customRolesGetCmd)
	customRolesCmd.AddCommand(customRolesCreateCmd)
	customRolesCmd.AddCommand(customRolesUpdateCmd)
	customRolesCmd.AddCommand(customRolesDeleteCmd)

	addOutputFlags(customRolesListCmd)
	addOutputFlags(customRolesGetCmd)
	addOutputFlags(customRolesCreateCmd)
	addOutputFlags(customRolesUpdateCmd)

	// create flags
	customRolesCreateCmd.Flags().String("name", "", "Role name (required)")
	customRolesCreateCmd.Flags().String("extended-role", "user", "Base role to extend (admin, user, observer)")
	customRolesCreateCmd.Flags().StringSlice("granted-rights", nil, "Comma-separated list of rights to grant")
	_ = customRolesCreateCmd.MarkFlagRequired("name")

	// update flags
	customRolesUpdateCmd.Flags().String("name", "", "Role name")
	customRolesUpdateCmd.Flags().String("extended-role", "", "Base role to extend")
	customRolesUpdateCmd.Flags().StringSlice("granted-rights", nil, "Comma-separated list of rights to grant")
}

var customRolesCmd = &cobra.Command{
	Use:   "custom-roles",
	Short: "Manage OpsGenie custom roles",
	Long:  "Create, list, and manage custom roles with specific granted rights.",
}

var customRolesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all custom roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v2/roles", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "NAME", "EXTENDED_ROLE"}
		rows := make([][]string, 0, len(resp.Data))
		for _, r := range resp.Data {
			rows = append(rows, []string{
				stringVal(r, "id"),
				stringVal(r, "name"),
				stringVal(r, "extendedRole"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var customRolesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a custom role by ID",
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
		if err := client.Get("/v2/roles/"+args[0], &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"Name", stringVal(resp.Data, "name")},
			{"ExtendedRole", stringVal(resp.Data, "extendedRole")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var customRolesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new custom role",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		extendedRole, _ := cmd.Flags().GetString("extended-role")
		grantedRights, _ := cmd.Flags().GetStringSlice("granted-rights")

		body := map[string]interface{}{
			"name":         name,
			"extendedRole": extendedRole,
		}
		if len(grantedRights) > 0 {
			body["grantedRights"] = grantedRights
		}

		var result map[string]interface{}
		if err := client.Post("/v2/roles", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Custom role %q created", name), opts)
		return output.RenderJSON(result, opts)
	},
}

var customRolesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a custom role",
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
		if cmd.Flags().Changed("extended-role") {
			v, _ := cmd.Flags().GetString("extended-role")
			body["extendedRole"] = v
		}
		if cmd.Flags().Changed("granted-rights") {
			v, _ := cmd.Flags().GetStringSlice("granted-rights")
			body["grantedRights"] = v
		}

		var result map[string]interface{}
		if err := client.Put("/v2/roles/"+args[0], body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Custom role %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var customRolesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a custom role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v2/roles/"+args[0], nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Custom role %q deleted", args[0]), opts)
		return nil
	},
}

// joinStrings joins a slice of strings with a separator, handling type assertions.
func joinStrings(v interface{}, sep string) string {
	if arr, ok := v.([]interface{}); ok {
		parts := make([]string, 0, len(arr))
		for _, item := range arr {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		return strings.Join(parts, sep)
	}
	return ""
}
