package cmd

import (
	"fmt"
	"net/url"

	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deploymentsCmd)
	deploymentsCmd.AddCommand(deploymentsListCmd)
	deploymentsCmd.AddCommand(deploymentsGetCmd)
	deploymentsCmd.AddCommand(deploymentsCreateCmd)
	deploymentsCmd.AddCommand(deploymentsUpdateCmd)
	deploymentsCmd.AddCommand(deploymentsSearchCmd)

	addOutputFlags(deploymentsListCmd)
	addOutputFlags(deploymentsGetCmd)
	addOutputFlags(deploymentsCreateCmd)
	addOutputFlags(deploymentsUpdateCmd)
	addOutputFlags(deploymentsSearchCmd)

	// create flags
	deploymentsCreateCmd.Flags().String("name", "", "Deployment name (required)")
	deploymentsCreateCmd.Flags().String("description", "", "Deployment description")
	deploymentsCreateCmd.Flags().String("service-id", "", "Service ID")
	deploymentsCreateCmd.Flags().String("environment", "", "Deployment environment")
	_ = deploymentsCreateCmd.MarkFlagRequired("name")

	// update flags
	deploymentsUpdateCmd.Flags().String("name", "", "Deployment name")
	deploymentsUpdateCmd.Flags().String("description", "", "Deployment description")
	deploymentsUpdateCmd.Flags().String("environment", "", "Deployment environment")

	// list flags (list delegates to search endpoint, service is required)
	deploymentsListCmd.Flags().String("service", "", "Service ID to list deployments for (required)")
	_ = deploymentsListCmd.MarkFlagRequired("service")
	deploymentsListCmd.Flags().String("environment", "", "Filter by environment")

	// search flags
	deploymentsSearchCmd.Flags().String("service", "", "Service ID to search deployments for (required)")
	_ = deploymentsSearchCmd.MarkFlagRequired("service")
	deploymentsSearchCmd.Flags().String("environment", "", "Filter by environment")
}

var deploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Manage OpsGenie deployments",
	Long:  "Create, list, and manage deployments for change tracking.",
}

var deploymentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployments for a service",
	Long:  "List deployments for a service using the OpsGenie search endpoint. --service is required.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		params := url.Values{}
		service, _ := cmd.Flags().GetString("service")
		params.Set("serviceIds", service)
		if environment, _ := cmd.Flags().GetString("environment"); environment != "" {
			params.Set("environment", environment)
		}

		path := "/v2/deployments/search?" + params.Encode()

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get(path, &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "NAME", "ENVIRONMENT", "STATUS", "CREATED_AT"}
		rows := make([][]string, 0, len(resp.Data))
		for _, d := range resp.Data {
			rows = append(rows, []string{
				stringVal(d, "id"),
				stringVal(d, "name"),
				stringVal(d, "environment"),
				stringVal(d, "status"),
				stringVal(d, "createdAt"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var deploymentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a deployment by ID",
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
		if err := client.Get("/v2/deployments/"+args[0], &resp); err != nil {
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
			{"Environment", stringVal(resp.Data, "environment")},
			{"Status", stringVal(resp.Data, "status")},
			{"CreatedAt", stringVal(resp.Data, "createdAt")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var deploymentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a deployment",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		serviceID, _ := cmd.Flags().GetString("service-id")
		environment, _ := cmd.Flags().GetString("environment")

		body := map[string]interface{}{
			"name":        name,
			"description": description,
			"environment": environment,
		}
		if serviceID != "" {
			body["serviceId"] = serviceID
		}

		var result map[string]interface{}
		if err := client.Post("/v2/deployments", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Deployment %q created", name), opts)
		return output.RenderJSON(result, opts)
	},
}

var deploymentsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a deployment",
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
		if cmd.Flags().Changed("environment") {
			v, _ := cmd.Flags().GetString("environment")
			body["environment"] = v
		}

		var result map[string]interface{}
		if err := client.Patch("/v2/deployments/"+args[0]+"/update", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Deployment %q updated", args[0]), opts)
		return output.RenderJSON(result, opts)
	},
}

var deploymentsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search deployments",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		params := url.Values{}
		service, _ := cmd.Flags().GetString("service")
		params.Set("serviceIds", service)
		if environment, _ := cmd.Flags().GetString("environment"); environment != "" {
			params.Set("environment", environment)
		}

		path := "/v2/deployments/search?" + params.Encode()

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get(path, &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "NAME", "ENVIRONMENT", "STATUS", "CREATED_AT"}
		rows := make([][]string, 0, len(resp.Data))
		for _, d := range resp.Data {
			rows = append(rows, []string{
				stringVal(d, "id"),
				stringVal(d, "name"),
				stringVal(d, "environment"),
				stringVal(d, "status"),
				stringVal(d, "createdAt"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}
