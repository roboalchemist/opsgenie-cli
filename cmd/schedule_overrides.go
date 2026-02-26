package cmd

import (
	"encoding/json"
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var scheduleOverridesCmd = &cobra.Command{
	Use:   "schedule-overrides",
	Short: "Manage schedule overrides",
}

var scheduleOverridesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List overrides for a schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}

		var resp struct {
			Data []api.ScheduleOverrideResponse `json:"data"`
		}
		if err := client.Get("/v2/schedules/"+scheduleID+"/overrides", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"Alias", "StartDate", "EndDate"}
		rows := make([][]string, len(resp.Data))
		for i, o := range resp.Data {
			rows[i] = []string{o.Alias, o.StartDate, o.EndDate}
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var scheduleOverridesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a schedule override by alias",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		alias, _ := cmd.Flags().GetString("alias")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}
		if alias == "" {
			return fmt.Errorf("--alias is required")
		}

		var resp struct {
			Data api.ScheduleOverrideResponse `json:"data"`
		}
		if err := client.Get("/v2/schedules/"+scheduleID+"/overrides/"+alias, &resp); err != nil {
			return err
		}

		return output.RenderJSON(resp.Data, opts)
	},
}

var scheduleOverridesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an override for a schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}

		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		userID, _ := cmd.Flags().GetString("user")
		rotationsJSON, _ := cmd.Flags().GetString("rotations")

		if startDate == "" || endDate == "" {
			return fmt.Errorf("--start-date and --end-date are required")
		}

		body := map[string]interface{}{
			"user":      map[string]string{"id": userID},
			"startDate": startDate,
			"endDate":   endDate,
		}

		if rotationsJSON != "" {
			var rotations interface{}
			if err := json.Unmarshal([]byte(rotationsJSON), &rotations); err != nil {
				return fmt.Errorf("invalid --rotations JSON: %w", err)
			}
			body["rotations"] = rotations
		}

		var resp struct {
			Data api.ScheduleOverrideResponse `json:"data"`
		}
		if err := client.Post("/v2/schedules/"+scheduleID+"/overrides", body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Override created for schedule %s", scheduleID), opts)
		return output.RenderJSON(resp.Data, opts)
	},
}

var scheduleOverridesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a schedule override by alias",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		alias, _ := cmd.Flags().GetString("alias")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}
		if alias == "" {
			return fmt.Errorf("--alias is required")
		}

		body := map[string]interface{}{}
		if startDate, _ := cmd.Flags().GetString("start-date"); startDate != "" {
			body["startDate"] = startDate
		}
		if endDate, _ := cmd.Flags().GetString("end-date"); endDate != "" {
			body["endDate"] = endDate
		}
		if userID, _ := cmd.Flags().GetString("user"); userID != "" {
			body["user"] = map[string]string{"id": userID}
		}
		if rotationsJSON, _ := cmd.Flags().GetString("rotations"); rotationsJSON != "" {
			var rotations interface{}
			if err := json.Unmarshal([]byte(rotationsJSON), &rotations); err != nil {
				return fmt.Errorf("invalid --rotations JSON: %w", err)
			}
			body["rotations"] = rotations
		}

		var resp struct {
			Data api.ScheduleOverrideResponse `json:"data"`
		}
		if err := client.Put("/v2/schedules/"+scheduleID+"/overrides/"+alias, body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Override %s updated", alias), opts)
		return nil
	},
}

var scheduleOverridesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a schedule override by alias",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		alias, _ := cmd.Flags().GetString("alias")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}
		if alias == "" {
			return fmt.Errorf("--alias is required")
		}

		var result json.RawMessage
		if err := client.Delete("/v2/schedules/"+scheduleID+"/overrides/"+alias, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Override %s deleted", alias), opts)
		return nil
	},
}

func init() {
	scheduleOverridesListCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	addOutputFlags(scheduleOverridesListCmd)

	scheduleOverridesGetCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleOverridesGetCmd.Flags().String("alias", "", "Override alias (required)")
	addOutputFlags(scheduleOverridesGetCmd)

	scheduleOverridesCreateCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleOverridesCreateCmd.Flags().String("start-date", "", "Override start date (ISO 8601, required)")
	scheduleOverridesCreateCmd.Flags().String("end-date", "", "Override end date (ISO 8601, required)")
	scheduleOverridesCreateCmd.Flags().String("user", "", "User ID for the override")
	scheduleOverridesCreateCmd.Flags().String("rotations", "", "JSON array of rotation references")

	scheduleOverridesUpdateCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleOverridesUpdateCmd.Flags().String("alias", "", "Override alias (required)")
	scheduleOverridesUpdateCmd.Flags().String("start-date", "", "New start date (ISO 8601)")
	scheduleOverridesUpdateCmd.Flags().String("end-date", "", "New end date (ISO 8601)")
	scheduleOverridesUpdateCmd.Flags().String("user", "", "New user ID")
	scheduleOverridesUpdateCmd.Flags().String("rotations", "", "JSON array of rotation references")

	scheduleOverridesDeleteCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleOverridesDeleteCmd.Flags().String("alias", "", "Override alias (required)")

	scheduleOverridesCmd.AddCommand(scheduleOverridesListCmd)
	scheduleOverridesCmd.AddCommand(scheduleOverridesGetCmd)
	scheduleOverridesCmd.AddCommand(scheduleOverridesCreateCmd)
	scheduleOverridesCmd.AddCommand(scheduleOverridesUpdateCmd)
	scheduleOverridesCmd.AddCommand(scheduleOverridesDeleteCmd)

	rootCmd.AddCommand(scheduleOverridesCmd)
}
