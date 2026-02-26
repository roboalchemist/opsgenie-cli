package cmd

import (
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(contactsCmd)
	contactsCmd.AddCommand(contactsListCmd)
	contactsCmd.AddCommand(contactsGetCmd)
	contactsCmd.AddCommand(contactsCreateCmd)
	contactsCmd.AddCommand(contactsUpdateCmd)
	contactsCmd.AddCommand(contactsDeleteCmd)
	contactsCmd.AddCommand(contactsEnableCmd)
	contactsCmd.AddCommand(contactsDisableCmd)

	addOutputFlags(contactsListCmd)
	addOutputFlags(contactsGetCmd)
	addOutputFlags(contactsCreateCmd)
	addOutputFlags(contactsUpdateCmd)

	// shared user flag
	for _, c := range []*cobra.Command{
		contactsListCmd, contactsGetCmd, contactsCreateCmd,
		contactsUpdateCmd, contactsDeleteCmd, contactsEnableCmd, contactsDisableCmd,
	} {
		c.Flags().String("user", "", "User ID or username (required)")
		_ = c.MarkFlagRequired("user")
	}

	// get, update, delete, enable, disable need contact ID
	for _, c := range []*cobra.Command{
		contactsGetCmd, contactsUpdateCmd, contactsDeleteCmd,
		contactsEnableCmd, contactsDisableCmd,
	} {
		c.Flags().String("contact-id", "", "Contact ID (required)")
		_ = c.MarkFlagRequired("contact-id")
	}

	// create flags
	contactsCreateCmd.Flags().String("method", "", "Contact method (email, sms, voice, mobile) (required)")
	contactsCreateCmd.Flags().String("to", "", "Contact destination (required)")
	_ = contactsCreateCmd.MarkFlagRequired("method")
	_ = contactsCreateCmd.MarkFlagRequired("to")

	// update flags
	contactsUpdateCmd.Flags().String("to", "", "Contact destination")
}

var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Manage OpsGenie user contacts",
	Long:  "Create, list, and manage contact methods for OpsGenie users.",
}

var contactsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List contacts for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")

		var resp struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v2/users/"+userID+"/contacts", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "METHOD", "TO", "ENABLED"}
		rows := make([][]string, 0, len(resp.Data))
		for _, c := range resp.Data {
			rows = append(rows, []string{
				stringVal(c, "id"),
				stringVal(c, "method"),
				stringVal(c, "to"),
				stringVal(c, "enabled"),
			})
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var contactsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")
		contactID, _ := cmd.Flags().GetString("contact-id")

		var resp struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := client.Get("/v2/users/"+userID+"/contacts/"+contactID, &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"FIELD", "VALUE"}
		rows := [][]string{
			{"ID", stringVal(resp.Data, "id")},
			{"Method", stringVal(resp.Data, "method")},
			{"To", stringVal(resp.Data, "to")},
			{"Enabled", stringVal(resp.Data, "enabled")},
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var contactsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a contact for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")
		method, _ := cmd.Flags().GetString("method")
		to, _ := cmd.Flags().GetString("to")

		body := map[string]interface{}{
			"method": method,
			"to":     to,
		}

		var result map[string]interface{}
		if err := client.Post("/v2/users/"+userID+"/contacts", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Contact (%s: %s) created for user %q", method, to, userID), opts)
		return output.RenderJSON(result, opts)
	},
}

var contactsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		userID, _ := cmd.Flags().GetString("user")
		contactID, _ := cmd.Flags().GetString("contact-id")

		body := map[string]interface{}{}
		if cmd.Flags().Changed("to") {
			v, _ := cmd.Flags().GetString("to")
			body["to"] = v
		}

		var result map[string]interface{}
		if err := client.Patch("/v2/users/"+userID+"/contacts/"+contactID, body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Contact %q updated for user %q", contactID, userID), opts)
		return output.RenderJSON(result, opts)
	},
}

var contactsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		userID, _ := cmd.Flags().GetString("user")
		contactID, _ := cmd.Flags().GetString("contact-id")

		if err := client.Delete("/v2/users/"+userID+"/contacts/"+contactID, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Contact %q deleted for user %q", contactID, userID), opts)
		return nil
	},
}

var contactsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		userID, _ := cmd.Flags().GetString("user")
		contactID, _ := cmd.Flags().GetString("contact-id")

		if err := client.Post("/v2/users/"+userID+"/contacts/"+contactID+"/enable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Contact %q enabled for user %q", contactID, userID), opts)
		return nil
	},
}

var contactsDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := GetOutputOptions()

		userID, _ := cmd.Flags().GetString("user")
		contactID, _ := cmd.Flags().GetString("contact-id")

		if err := client.Post("/v2/users/"+userID+"/contacts/"+contactID+"/disable", nil, nil); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Contact %q disabled for user %q", contactID, userID), opts)
		return nil
	},
}
