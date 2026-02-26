package cmd

import (
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var onCallCmd = &cobra.Command{
	Use:   "on-call",
	Short: "Query on-call schedules",
}

var onCallGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current on-call participants for a schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		flat, _ := cmd.Flags().GetBool("flat")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}

		path := "/v2/schedules/" + scheduleID + "/on-calls"
		if flat {
			path += "?flat=true"
		}

		var resp struct {
			Data api.OnCallResponse `json:"data"`
		}
		if err := client.Get(path, &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"Schedule", "Start", "End", "Recipient"}
		var rows [][]string
		for _, p := range resp.Data.OnCallParticipants {
			rows = append(rows, []string{
				resp.Data.ScheduleRef.Name,
				p.OnCallStart,
				p.OnCallEnd,
				p.Name,
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var onCallNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Get next on-call participants for a schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		scheduleID, _ := cmd.Flags().GetString("schedule")
		flat, _ := cmd.Flags().GetBool("flat")
		if scheduleID == "" {
			return fmt.Errorf("--schedule is required")
		}

		path := "/v2/schedules/" + scheduleID + "/next-on-calls"
		if flat {
			path += "?flat=true"
		}

		var resp struct {
			Data api.OnCallResponse `json:"data"`
		}
		if err := client.Get(path, &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"Schedule", "Start", "End", "Recipient"}
		var rows [][]string
		for _, p := range resp.Data.OnCallParticipants {
			rows = append(rows, []string{
				resp.Data.ScheduleRef.Name,
				p.OnCallStart,
				p.OnCallEnd,
				p.Name,
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

func init() {
	onCallGetCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	onCallGetCmd.Flags().Bool("flat", false, "Return a flat list of on-call participants")
	addOutputFlags(onCallGetCmd)

	onCallNextCmd.Flags().String("schedule", "", "Schedule ID or name (required)")
	onCallNextCmd.Flags().Bool("flat", false, "Return a flat list of on-call participants")
	addOutputFlags(onCallNextCmd)

	onCallCmd.AddCommand(onCallGetCmd)
	onCallCmd.AddCommand(onCallNextCmd)

	rootCmd.AddCommand(onCallCmd)
}
