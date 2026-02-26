package cmd

import (
	"encoding/json"
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var schedulesCmd = &cobra.Command{
	Use:   "schedules",
	Short: "Manage OpsGenie schedules",
}

var schedulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all schedules",
	Example: `  # List all schedules
  opsgenie-cli schedules list

  # List schedules as JSON with only id and name fields
  opsgenie-cli schedules list --json --fields id,name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []api.ScheduleResponse `json:"data"`
		}
		if err := client.Get("/v2/schedules", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "Name", "Timezone", "Enabled"}
		rows := make([][]string, len(resp.Data))
		for i, s := range resp.Data {
			enabled := "false"
			if s.Enabled {
				enabled = "true"
			}
			rows[i] = []string{s.ID, s.Name, s.Timezone, enabled}
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var schedulesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a schedule by ID or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data api.ScheduleResponse `json:"data"`
		}
		if err := client.Get("/v2/schedules/"+args[0], &resp); err != nil {
			return err
		}

		return output.RenderJSON(resp.Data, opts)
	},
}

var schedulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		timezone, _ := cmd.Flags().GetString("timezone")
		description, _ := cmd.Flags().GetString("description")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body := map[string]interface{}{
			"name":        name,
			"timezone":    timezone,
			"description": description,
		}

		var resp struct {
			Data api.ScheduleResponse `json:"data"`
		}
		if err := client.Post("/v2/schedules", body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Schedule %q created (id: %s)", resp.Data.Name, resp.Data.ID), opts)
		return output.RenderJSON(resp.Data, opts)
	},
}

var schedulesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a schedule by ID or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		body := map[string]interface{}{}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			body["name"] = name
		}
		if tz, _ := cmd.Flags().GetString("timezone"); tz != "" {
			body["timezone"] = tz
		}
		if desc, _ := cmd.Flags().GetString("description"); desc != "" {
			body["description"] = desc
		}
		if cmd.Flags().Changed("enabled") {
			enabled, _ := cmd.Flags().GetBool("enabled")
			body["enabled"] = enabled
		}

		var resp struct {
			Data api.ScheduleResponse `json:"data"`
		}
		if err := client.Patch("/v2/schedules/"+args[0], body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Schedule %s updated", args[0]), opts)
		if resp.Data.ID != "" {
			return output.RenderJSON(resp.Data, opts)
		}
		return nil
	},
}

var schedulesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a schedule by ID or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var result json.RawMessage
		if err := client.Delete("/v2/schedules/"+args[0], &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Schedule %s deleted", args[0]), opts)
		return nil
	},
}

func init() {
	schedulesCreateCmd.Flags().String("name", "", "Schedule name (required)")
	schedulesCreateCmd.Flags().String("timezone", "UTC", "Schedule timezone")
	schedulesCreateCmd.Flags().String("description", "", "Schedule description")

	schedulesUpdateCmd.Flags().String("name", "", "New schedule name")
	schedulesUpdateCmd.Flags().String("timezone", "", "New timezone")
	schedulesUpdateCmd.Flags().String("description", "", "New description")
	schedulesUpdateCmd.Flags().Bool("enabled", true, "Enable or disable the schedule")

	addOutputFlags(schedulesListCmd)
	addOutputFlags(schedulesGetCmd)

	schedulesCmd.AddCommand(schedulesListCmd)
	schedulesCmd.AddCommand(schedulesGetCmd)
	schedulesCmd.AddCommand(schedulesCreateCmd)
	schedulesCmd.AddCommand(schedulesUpdateCmd)
	schedulesCmd.AddCommand(schedulesDeleteCmd)

	rootCmd.AddCommand(schedulesCmd)
}
