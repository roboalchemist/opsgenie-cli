package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/roboalchemist/opsgenie-cli/pkg/api"
	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var escalationsCmd = &cobra.Command{
	Use:   "escalations",
	Short: "Manage OpsGenie escalation policies",
}

var escalationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all escalation policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []api.EscalationResponse `json:"data"`
		}
		if err := client.Get("/v2/escalations", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "Name", "Description"}
		rows := make([][]string, len(resp.Data))
		for i, e := range resp.Data {
			rows[i] = []string{e.ID, e.Name, e.Description}
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var escalationsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an escalation policy by ID or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data api.EscalationResponse `json:"data"`
		}
		if err := client.Get("/v2/escalations/"+args[0], &resp); err != nil {
			return err
		}

		return output.RenderJSON(resp.Data, opts)
	},
}

var escalationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an escalation policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		rulesJSON, _ := cmd.Flags().GetString("rules")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body := map[string]interface{}{
			"name":        name,
			"description": description,
		}

		if rulesJSON != "" {
			var rules interface{}
			if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
				return fmt.Errorf("invalid --rules JSON: %w", err)
			}
			body["rules"] = rules
		}

		var resp struct {
			Data api.EscalationResponse `json:"data"`
		}
		if err := client.Post("/v2/escalations", body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Escalation %q created (id: %s)", resp.Data.Name, resp.Data.ID), opts)
		return output.RenderJSON(resp.Data, opts)
	},
}

var escalationsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an escalation policy by ID or name",
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
		if desc, _ := cmd.Flags().GetString("description"); desc != "" {
			body["description"] = desc
		}
		if rulesJSON, _ := cmd.Flags().GetString("rules"); rulesJSON != "" {
			var rules interface{}
			if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
				return fmt.Errorf("invalid --rules JSON: %w", err)
			}
			body["rules"] = rules
		}

		var resp struct {
			Data api.EscalationResponse `json:"data"`
		}
		if err := client.Put("/v2/escalations/"+args[0], body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Escalation %s updated", args[0]), opts)
		if resp.Data.ID != "" {
			return output.RenderJSON(resp.Data, opts)
		}
		return nil
	},
}

var escalationsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an escalation policy by ID or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var result json.RawMessage
		if err := client.Delete("/v2/escalations/"+args[0], &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Escalation %s deleted", args[0]), opts)
		return nil
	},
}

func init() {
	escalationsCreateCmd.Flags().String("name", "", "Escalation policy name (required)")
	escalationsCreateCmd.Flags().String("description", "", "Escalation policy description")
	escalationsCreateCmd.Flags().String("rules", "", "JSON array of escalation rules")

	escalationsUpdateCmd.Flags().String("name", "", "New name")
	escalationsUpdateCmd.Flags().String("description", "", "New description")
	escalationsUpdateCmd.Flags().String("rules", "", "JSON array of escalation rules")

	addOutputFlags(escalationsListCmd)
	addOutputFlags(escalationsGetCmd)

	escalationsCmd.AddCommand(escalationsListCmd)
	escalationsCmd.AddCommand(escalationsGetCmd)
	escalationsCmd.AddCommand(escalationsCreateCmd)
	escalationsCmd.AddCommand(escalationsUpdateCmd)
	escalationsCmd.AddCommand(escalationsDeleteCmd)

	rootCmd.AddCommand(escalationsCmd)
}
