package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage OpsGenie users",
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users (paginated)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var users []api.UserResponse
		if err := client.ListAll("/v2/users", url.Values{}, &users); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(users, opts)
		}

		headers := []string{"ID", "Username", "FullName", "Role", "Verified"}
		rows := make([][]string, len(users))
		for i, u := range users {
			verified := "false"
			if u.Verified {
				verified = "true"
			}
			rows[i] = []string{u.ID, u.Username, u.FullName, u.Role.Name, verified}
		}
		return output.RenderTable(headers, rows, users, opts)
	},
}

var usersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a user by ID or username",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var resp struct {
			Data api.UserResponse `json:"data"`
		}
		if err := client.Get("/v2/users/"+args[0], &resp); err != nil {
			return err
		}

		return output.RenderJSON(resp.Data, opts)
	},
}

var usersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		username, _ := cmd.Flags().GetString("username")
		fullName, _ := cmd.Flags().GetString("full-name")
		role, _ := cmd.Flags().GetString("role")

		if username == "" {
			return fmt.Errorf("--username is required")
		}

		body := map[string]interface{}{
			"username": username,
			"fullName": fullName,
			"role":     map[string]string{"name": role},
		}

		var resp struct {
			Data api.UserResponse `json:"data"`
		}
		if err := client.Post("/v2/users", body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("User %q created (id: %s)", resp.Data.Username, resp.Data.ID), opts)
		return output.RenderJSON(resp.Data, opts)
	},
}

var usersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a user by ID or username",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		body := map[string]interface{}{}
		if fullName, _ := cmd.Flags().GetString("full-name"); fullName != "" {
			body["fullName"] = fullName
		}
		if role, _ := cmd.Flags().GetString("role"); role != "" {
			body["role"] = map[string]string{"name": role}
		}

		var resp struct {
			Data api.UserResponse `json:"data"`
		}
		if err := client.Patch("/v2/users/"+args[0], body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("User %s updated", args[0]), opts)
		if resp.Data.ID != "" {
			return output.RenderJSON(resp.Data, opts)
		}
		return nil
	},
}

var usersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a user by ID or username",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var result json.RawMessage
		if err := client.Delete("/v2/users/"+args[0], &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("User %s deleted", args[0]), opts)
		return nil
	},
}

func init() {
	usersCreateCmd.Flags().String("username", "", "User email/username (required)")
	usersCreateCmd.Flags().String("full-name", "", "Full name")
	usersCreateCmd.Flags().String("role", "user", "Role name (e.g. admin, user, observer)")

	usersUpdateCmd.Flags().String("full-name", "", "New full name")
	usersUpdateCmd.Flags().String("role", "", "New role name")

	addOutputFlags(usersListCmd)
	addOutputFlags(usersGetCmd)

	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersGetCmd)
	usersCmd.AddCommand(usersCreateCmd)
	usersCmd.AddCommand(usersUpdateCmd)
	usersCmd.AddCommand(usersDeleteCmd)

	rootCmd.AddCommand(usersCmd)
}
