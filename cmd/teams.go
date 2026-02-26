package cmd

import (
	"encoding/json"
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage OpsGenie teams",
}

var teamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	Example: `  # List teams as a table
  opsgenie-cli teams list

  # List teams as JSON and filter with jq
  opsgenie-cli teams list --json | jq '.[].name'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []api.TeamResponse `json:"data"`
		}
		if err := client.Get("/v2/teams", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "Name", "Description"}
		rows := make([][]string, len(resp.Data))
		for i, t := range resp.Data {
			rows[i] = []string{t.ID, t.Name, t.Description}
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var teamsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a team by ID or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data api.TeamResponse `json:"data"`
		}
		if err := client.Get("/v2/teams/"+args[0], &resp); err != nil {
			return err
		}

		return output.RenderJSON(resp.Data, opts)
	},
}

var teamsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		body := map[string]interface{}{
			"name":        name,
			"description": description,
		}

		var resp struct {
			Data api.TeamResponse `json:"data"`
		}
		if err := client.Post("/v2/teams", body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Team %q created (id: %s)", resp.Data.Name, resp.Data.ID), opts)
		return output.RenderJSON(resp.Data, opts)
	},
}

var teamsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a team by ID or name",
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

		var resp struct {
			Data api.TeamResponse `json:"data"`
		}
		if err := client.Patch("/v2/teams/"+args[0], body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Team %s updated", args[0]), opts)
		if resp.Data.ID != "" {
			return output.RenderJSON(resp.Data, opts)
		}
		return nil
	},
}

var teamsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a team by ID or name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var result json.RawMessage
		if err := client.Delete("/v2/teams/"+args[0], &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Team %s deleted", args[0]), opts)
		return nil
	},
}

func init() {
	teamsCreateCmd.Flags().String("name", "", "Team name (required)")
	teamsCreateCmd.Flags().String("description", "", "Team description")

	teamsUpdateCmd.Flags().String("name", "", "New team name")
	teamsUpdateCmd.Flags().String("description", "", "New team description")

	addOutputFlags(teamsListCmd)
	addOutputFlags(teamsGetCmd)

	teamsCmd.AddCommand(teamsListCmd)
	teamsCmd.AddCommand(teamsGetCmd)
	teamsCmd.AddCommand(teamsCreateCmd)
	teamsCmd.AddCommand(teamsUpdateCmd)
	teamsCmd.AddCommand(teamsDeleteCmd)

	rootCmd.AddCommand(teamsCmd)
}
