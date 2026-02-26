package cmd

import (
	"fmt"
	"strconv"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(heartbeatsCmd)
	heartbeatsCmd.AddCommand(heartbeatsListCmd)
	heartbeatsCmd.AddCommand(heartbeatsGetCmd)
	heartbeatsCmd.AddCommand(heartbeatsCreateCmd)
	heartbeatsCmd.AddCommand(heartbeatsUpdateCmd)
	heartbeatsCmd.AddCommand(heartbeatsDeleteCmd)
	heartbeatsCmd.AddCommand(heartbeatsEnableCmd)
	heartbeatsCmd.AddCommand(heartbeatsDisableCmd)
	heartbeatsCmd.AddCommand(heartbeatsPingCmd)

	addOutputFlags(heartbeatsListCmd)
	addOutputFlags(heartbeatsGetCmd)
	addOutputFlags(heartbeatsCreateCmd)
	addOutputFlags(heartbeatsUpdateCmd)

	// create flags
	heartbeatsCreateCmd.Flags().String("name", "", "Heartbeat name (required)")
	heartbeatsCreateCmd.Flags().String("description", "", "Heartbeat description")
	heartbeatsCreateCmd.Flags().Int("interval", 10, "Ping interval")
	heartbeatsCreateCmd.Flags().String("interval-unit", "minutes", "Interval unit (minutes, hours, days)")
	heartbeatsCreateCmd.Flags().Bool("enabled", true, "Whether heartbeat is enabled")
	_ = heartbeatsCreateCmd.MarkFlagRequired("name")

	// update flags
	heartbeatsUpdateCmd.Flags().String("description", "", "Heartbeat description")
	heartbeatsUpdateCmd.Flags().Int("interval", 0, "Ping interval")
	heartbeatsUpdateCmd.Flags().String("interval-unit", "", "Interval unit (minutes, hours, days)")
	heartbeatsUpdateCmd.Flags().Bool("enabled", true, "Whether heartbeat is enabled")
}

var heartbeatsCmd = &cobra.Command{
	Use:   "heartbeats",
	Short: "Manage OpsGenie heartbeat monitors",
	Long:  "Create, list, and manage heartbeat monitors that alert when a service fails to ping.",
}

var heartbeatsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all heartbeats",
	Example: `  # List all heartbeats
  opsgenie-cli heartbeats list

  # List heartbeats as JSON
  opsgenie-cli heartbeats list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []api.HeartbeatResponse `json:"data"`
		}
		if err := client.Get("/v2/heartbeats", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"NAME", "ENABLED", "EXPIRED", "INTERVAL", "LAST_PING_AT"}
		rows := make([][]string, 0, len(resp.Data))
		for _, h := range resp.Data {
			rows = append(rows, []string{
				h.Name,
				strconv.FormatBool(h.Enabled),
				strconv.FormatBool(h.Expired),
				fmt.Sprintf("%d %s", h.Interval, h.IntervalUnit),
				h.LastPingAt,
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var heartbeatsGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a heartbeat by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data api.HeartbeatResponse `json:"data"`
		}
		if err := client.Get("/v2/heartbeats/"+args[0], &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"Name", resp.Data.Name},
			{"Description", resp.Data.Description},
			{"Enabled", strconv.FormatBool(resp.Data.Enabled)},
			{"Expired", strconv.FormatBool(resp.Data.Expired)},
			{"Interval", fmt.Sprintf("%d %s", resp.Data.Interval, resp.Data.IntervalUnit)},
			{"LastPingAt", resp.Data.LastPingAt},
			{"AlertMessage", resp.Data.AlertMessage},
			{"AlertPriority", resp.Data.AlertPriority},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var heartbeatsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new heartbeat",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		interval, _ := cmd.Flags().GetInt("interval")
		intervalUnit, _ := cmd.Flags().GetString("interval-unit")
		enabled, _ := cmd.Flags().GetBool("enabled")

		body := map[string]interface{}{
			"name":         name,
			"description":  description,
			"interval":     interval,
			"intervalUnit": intervalUnit,
			"enabled":      enabled,
		}

		var result map[string]interface{}
		if err := client.Post("/v2/heartbeats", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Heartbeat %q created", name), opts)
		return output.RenderJSON(result, opts)
	},
}

var heartbeatsUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a heartbeat",
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
		if cmd.Flags().Changed("interval") {
			v, _ := cmd.Flags().GetInt("interval")
			body["interval"] = v
		}
		if cmd.Flags().Changed("interval-unit") {
			v, _ := cmd.Flags().GetString("interval-unit")
			body["intervalUnit"] = v
		}
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			body["enabled"] = v
		}

		var result map[string]interface{}
		if err := client.Patch("/v2/heartbeats/"+args[0], body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Heartbeat %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var heartbeatsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a heartbeat",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v2/heartbeats/"+args[0], nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Heartbeat %q deleted", args[0]), opts)
		return nil
	},
}

var heartbeatsEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a heartbeat",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Post("/v2/heartbeats/"+args[0]+"/enable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Heartbeat %q enabled", args[0]), opts)
		return nil
	},
}

var heartbeatsDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a heartbeat",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Post("/v2/heartbeats/"+args[0]+"/disable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Heartbeat %q disabled", args[0]), opts)
		return nil
	},
}

var heartbeatsPingCmd = &cobra.Command{
	Use:   "ping <name>",
	Short: "Ping a heartbeat",
	Example: `  # Ping a heartbeat from a cron job
  opsgenie-cli heartbeats ping my-service-heartbeat

  # Ping silently (no output on success)
  opsgenie-cli heartbeats ping my-service-heartbeat --quiet`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Get("/v2/heartbeats/"+args[0]+"/ping", nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Heartbeat %q pinged", args[0]), opts)
		return nil
	},
}
