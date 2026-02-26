package cmd

import (
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(servicesListCmd)
	servicesCmd.AddCommand(servicesGetCmd)
	servicesCmd.AddCommand(servicesCreateCmd)
	servicesCmd.AddCommand(servicesUpdateCmd)
	servicesCmd.AddCommand(servicesDeleteCmd)

	addOutputFlags(servicesListCmd)
	addOutputFlags(servicesGetCmd)
	addOutputFlags(servicesCreateCmd)
	addOutputFlags(servicesUpdateCmd)

	// create flags
	servicesCreateCmd.Flags().String("name", "", "Service name (required)")
	servicesCreateCmd.Flags().String("description", "", "Service description")
	servicesCreateCmd.Flags().String("team-id", "", "Team ID that owns this service")
	_ = servicesCreateCmd.MarkFlagRequired("name")

	// update flags
	servicesUpdateCmd.Flags().String("name", "", "Service name")
	servicesUpdateCmd.Flags().String("description", "", "Service description")
	servicesUpdateCmd.Flags().String("team-id", "", "Team ID that owns this service")
}

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage OpsGenie services",
	Long:  "Create, list, and manage services for incident management.",
}

var servicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v1/services", &resp); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "NAME", "DESCRIPTION", "TEAM_ID"}
		rows := make([][]string, 0, len(resp.Data))
		for _, s := range resp.Data {
			rows = append(rows, []string{
				stringVal(s, "id"),
				stringVal(s, "name"),
				stringVal(s, "description"),
				stringVal(s, "teamId"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var servicesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a service by ID",
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
		if err := client.Get("/v1/services/"+args[0], &resp); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"Name", stringVal(resp.Data, "name")},
			{"Description", stringVal(resp.Data, "description")},
			{"TeamID", stringVal(resp.Data, "teamId")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var servicesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		teamID, _ := cmd.Flags().GetString("team-id")

		body := map[string]interface{}{
			"name":        name,
			"description": description,
		}
		if teamID != "" {
			body["teamId"] = teamID
		}

		var result map[string]interface{}
		if err := client.Post("/v1/services", body, &result); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Service %q created", name), opts)
		return output.RenderJSON(result, opts)
	},
}

var servicesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a service",
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
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			body["description"] = v
		}
		if cmd.Flags().Changed("team-id") {
			v, _ := cmd.Flags().GetString("team-id")
			body["teamId"] = v
		}

		var result map[string]interface{}
		if err := client.Patch("/v1/services/"+args[0], body, &result); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Service %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var servicesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v1/services/"+args[0], nil); err != nil {
			output.Error(err.Error(), opts)
			return err
		}

		output.Success(fmt.Sprintf("Service %q deleted", args[0]), opts)
		return nil
	},
}
