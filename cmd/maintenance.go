package cmd

import (
	"fmt"

	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(maintenanceCmd)
	maintenanceCmd.AddCommand(maintenanceListCmd)
	maintenanceCmd.AddCommand(maintenanceGetCmd)
	maintenanceCmd.AddCommand(maintenanceCreateCmd)
	maintenanceCmd.AddCommand(maintenanceUpdateCmd)
	maintenanceCmd.AddCommand(maintenanceDeleteCmd)
	maintenanceCmd.AddCommand(maintenanceCancelCmd)

	addOutputFlags(maintenanceListCmd)
	addOutputFlags(maintenanceGetCmd)
	addOutputFlags(maintenanceCreateCmd)
	addOutputFlags(maintenanceUpdateCmd)

	// create/update flags
	for _, c := range []*cobra.Command{maintenanceCreateCmd, maintenanceUpdateCmd} {
		c.Flags().String("description", "", "Maintenance description")
		c.Flags().String("start-date", "", "Start date (RFC3339)")
		c.Flags().String("end-date", "", "End date (RFC3339)")
		c.Flags().String("type", "schedule-based", "Maintenance type (schedule-based)")
	}
}

var maintenanceCmd = &cobra.Command{
	Use:   "maintenance",
	Short: "Manage OpsGenie maintenance windows",
	Long:  "Create, list, and manage maintenance windows that suppress alerts.",
}

var maintenanceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all maintenance windows",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v1/maintenance", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "DESCRIPTION", "STATUS", "START_DATE", "END_DATE"}
		rows := make([][]string, 0, len(resp.Data))
		for _, m := range resp.Data {
			rows = append(rows, []string{
				stringVal(m, "id"),
				stringVal(m, "description"),
				stringVal(m, "status"),
				stringVal(m, "startDate"),
				stringVal(m, "endDate"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var maintenanceGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a maintenance window by ID",
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
		if err := client.Get("/v1/maintenance/"+args[0], &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"Description", stringVal(resp.Data, "description")},
			{"Status", stringVal(resp.Data, "status")},
			{"StartDate", stringVal(resp.Data, "startDate")},
			{"EndDate", stringVal(resp.Data, "endDate")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var maintenanceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a maintenance window",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		description, _ := cmd.Flags().GetString("description")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		mType, _ := cmd.Flags().GetString("type")

		body := map[string]interface{}{
			"description": description,
			"time": map[string]interface{}{
				"type":      mType,
				"startDate": startDate,
				"endDate":   endDate,
			},
		}

		var result map[string]interface{}
		if err := client.Post("/v1/maintenance", body, &result); err != nil {
			return err
		}

		output.Success("Maintenance window created", opts)
		return output.RenderJSON(result, opts)
	},
}

var maintenanceUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a maintenance window",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		body := map[string]interface{}{}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			body["description"] = v
		}
		timeMap := map[string]interface{}{}
		if cmd.Flags().Changed("start-date") {
			v, _ := cmd.Flags().GetString("start-date")
			timeMap["startDate"] = v
		}
		if cmd.Flags().Changed("end-date") {
			v, _ := cmd.Flags().GetString("end-date")
			timeMap["endDate"] = v
		}
		if cmd.Flags().Changed("type") {
			v, _ := cmd.Flags().GetString("type")
			timeMap["type"] = v
		}
		if len(timeMap) > 0 {
			body["time"] = timeMap
		}

		var result map[string]interface{}
		if err := client.Put("/v1/maintenance/"+args[0], body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Maintenance window %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var maintenanceDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a maintenance window",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v1/maintenance/"+args[0], nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Maintenance window %q deleted", args[0]), opts)
		return nil
	},
}

var maintenanceCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a maintenance window",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Post("/v1/maintenance/"+args[0]+"/cancel", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Maintenance window %q cancelled", args[0]), opts)
		return nil
	},
}

// stringVal safely extracts a string value from a map[string]interface{}.
func stringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}
