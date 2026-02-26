package cmd

import (
	"fmt"

	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(postmortemsCmd)
	postmortemsCmd.AddCommand(postmortemsGetCmd)
	postmortemsCmd.AddCommand(postmortemsCreateCmd)
	postmortemsCmd.AddCommand(postmortemsUpdateCmd)
	postmortemsCmd.AddCommand(postmortemsDeleteCmd)

	addOutputFlags(postmortemsGetCmd)
	addOutputFlags(postmortemsCreateCmd)
	addOutputFlags(postmortemsUpdateCmd)

	// create flags
	postmortemsCreateCmd.Flags().String("incident-id", "", "Incident ID to create postmortem for (required)")
	_ = postmortemsCreateCmd.MarkFlagRequired("incident-id")

	// update flags
	postmortemsUpdateCmd.Flags().String("title", "", "Postmortem title")
	postmortemsUpdateCmd.Flags().String("description", "", "Postmortem description")
}

var postmortemsCmd = &cobra.Command{
	Use:   "postmortems",
	Short: "Manage OpsGenie postmortems",
	Long:  "Create, get, update, and delete postmortems linked to incidents.",
}

var postmortemsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a postmortem by ID",
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
		if err := client.Get("/v2/postmortem/"+args[0], &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"Title", stringVal(resp.Data, "title")},
			{"Description", stringVal(resp.Data, "description")},
			{"IncidentID", stringVal(resp.Data, "incidentId")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var postmortemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a postmortem",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		incidentID, _ := cmd.Flags().GetString("incident-id")

		body := map[string]interface{}{
			"incidentId": incidentID,
		}

		var result map[string]interface{}
		if err := client.Post("/v2/postmortem", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Postmortem created for incident %q", incidentID), opts)
		return output.RenderJSON(result, opts)
	},
}

var postmortemsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a postmortem",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		body := map[string]interface{}{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			body["title"] = v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			body["description"] = v
		}

		var result map[string]interface{}
		if err := client.Put("/v2/postmortem/"+args[0], body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Postmortem %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var postmortemsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a postmortem",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		if err := client.Delete("/v2/postmortem/"+args[0], nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Postmortem %q deleted", args[0]), opts)
		return nil
	},
}
