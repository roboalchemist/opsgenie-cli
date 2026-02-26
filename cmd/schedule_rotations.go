package cmd

import (
	"encoding/json"
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var scheduleRotationsCmd = &cobra.Command{
	Use:   "schedule-rotations",
	Short: "Manage schedule rotations",
}

var scheduleRotationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rotations for a schedule",
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
			Data []api.ScheduleRotationResponse `json:"data"`
		}
		if err := client.Get("/v2/schedules/"+scheduleID+"/rotations", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "Name", "Type", "Length", "StartDate"}
		rows := make([][]string, len(resp.Data))
		for i, r := range resp.Data {
			rows[i] = []string{r.ID, r.Name, r.Type, fmt.Sprintf("%d", r.Length), r.StartDate}
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var scheduleRotationsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a schedule rotation by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		rotationID, _ := cmd.Flags().GetString("id")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}
		if rotationID == "" {
			return fmt.Errorf("--id is required")
		}

		var resp struct {
			Data api.ScheduleRotationResponse `json:"data"`
		}
		if err := client.Get("/v2/schedules/"+scheduleID+"/rotations/"+rotationID, &resp); err != nil {
			return err
		}

		return output.RenderJSON(resp.Data, opts)
	},
}

var scheduleRotationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a rotation for a schedule",
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

		name, _ := cmd.Flags().GetString("name")
		rotType, _ := cmd.Flags().GetString("type")
		startDate, _ := cmd.Flags().GetString("start-date")
		length, _ := cmd.Flags().GetInt("length")
		participantsJSON, _ := cmd.Flags().GetString("participants")

		body := map[string]interface{}{
			"name":      name,
			"type":      rotType,
			"startDate": startDate,
			"length":    length,
		}

		if participantsJSON != "" {
			var participants interface{}
			if err := json.Unmarshal([]byte(participantsJSON), &participants); err != nil {
				return fmt.Errorf("invalid --participants JSON: %w", err)
			}
			body["participants"] = participants
		}

		var resp struct {
			Data api.ScheduleRotationResponse `json:"data"`
		}
		if err := client.Post("/v2/schedules/"+scheduleID+"/rotations", body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Rotation created for schedule %s", scheduleID), opts)
		return output.RenderJSON(resp.Data, opts)
	},
}

var scheduleRotationsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a schedule rotation",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		rotationID, _ := cmd.Flags().GetString("id")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}
		if rotationID == "" {
			return fmt.Errorf("--id is required")
		}

		body := map[string]interface{}{}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			body["name"] = name
		}
		if rotType, _ := cmd.Flags().GetString("type"); rotType != "" {
			body["type"] = rotType
		}
		if startDate, _ := cmd.Flags().GetString("start-date"); startDate != "" {
			body["startDate"] = startDate
		}
		if cmd.Flags().Changed("length") {
			length, _ := cmd.Flags().GetInt("length")
			body["length"] = length
		}
		if participantsJSON, _ := cmd.Flags().GetString("participants"); participantsJSON != "" {
			var participants interface{}
			if err := json.Unmarshal([]byte(participantsJSON), &participants); err != nil {
				return fmt.Errorf("invalid --participants JSON: %w", err)
			}
			body["participants"] = participants
		}

		var resp struct {
			Data api.ScheduleRotationResponse `json:"data"`
		}
		if err := client.Patch("/v2/schedules/"+scheduleID+"/rotations/"+rotationID, body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Rotation %s updated", rotationID), opts)
		return nil
	},
}

var scheduleRotationsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a schedule rotation",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		rotationID, _ := cmd.Flags().GetString("id")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}
		if rotationID == "" {
			return fmt.Errorf("--id is required")
		}

		var result json.RawMessage
		if err := client.Delete("/v2/schedules/"+scheduleID+"/rotations/"+rotationID, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Rotation %s deleted", rotationID), opts)
		return nil
	},
}

func init() {
	scheduleRotationsListCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	addOutputFlags(scheduleRotationsListCmd)

	scheduleRotationsGetCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleRotationsGetCmd.Flags().String("id", "", "Rotation ID (required)")
	addOutputFlags(scheduleRotationsGetCmd)

	scheduleRotationsCreateCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleRotationsCreateCmd.Flags().String("name", "", "Rotation name")
	scheduleRotationsCreateCmd.Flags().String("type", "weekly", "Rotation type (weekly, daily, hourly)")
	scheduleRotationsCreateCmd.Flags().String("start-date", "", "Start date (ISO 8601)")
	scheduleRotationsCreateCmd.Flags().Int("length", 1, "Rotation length")
	scheduleRotationsCreateCmd.Flags().String("participants", "", "JSON array of participant objects")

	scheduleRotationsUpdateCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleRotationsUpdateCmd.Flags().String("id", "", "Rotation ID (required)")
	scheduleRotationsUpdateCmd.Flags().String("name", "", "New name")
	scheduleRotationsUpdateCmd.Flags().String("type", "", "New type")
	scheduleRotationsUpdateCmd.Flags().String("start-date", "", "New start date (ISO 8601)")
	scheduleRotationsUpdateCmd.Flags().Int("length", 0, "New length")
	scheduleRotationsUpdateCmd.Flags().String("participants", "", "JSON array of participant objects")

	scheduleRotationsDeleteCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	scheduleRotationsDeleteCmd.Flags().String("id", "", "Rotation ID (required)")

	scheduleRotationsCmd.AddCommand(scheduleRotationsListCmd)
	scheduleRotationsCmd.AddCommand(scheduleRotationsGetCmd)
	scheduleRotationsCmd.AddCommand(scheduleRotationsCreateCmd)
	scheduleRotationsCmd.AddCommand(scheduleRotationsUpdateCmd)
	scheduleRotationsCmd.AddCommand(scheduleRotationsDeleteCmd)

	rootCmd.AddCommand(scheduleRotationsCmd)
}
