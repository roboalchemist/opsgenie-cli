package cmd

import (
	"fmt"
	"strconv"

	"github.com/roboalchemist/opsgenie-cli/pkg/api"
	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

// accountCmd is the parent command for account operations.
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage OpsGenie account information",
}

// ─── account get ─────────────────────────────────────────────────────────────

var accountGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get account information",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var envelope api.APIResponse[api.AccountResponse]
		if err := client.Get("/v2/account", &envelope); err != nil {
			return err
		}
		acct := envelope.Data

		headers := []string{"Field", "Value"}
		rows := [][]string{
			{"Name", acct.Name},
			{"Plan", acct.Plan.Name},
			{"Plan MaxUsers", strconv.Itoa(acct.Plan.MaxUserCount)},
			{"Plan IsExpired", fmt.Sprintf("%v", acct.Plan.IsExpired)},
			{"UserCount", strconv.Itoa(acct.UserCount)},
			{"IsYearly", fmt.Sprintf("%v", acct.IsYearly)},
		}
		return output.RenderTable(headers, rows, acct, opts)
	},
}

func init() {
	accountCmd.AddCommand(accountGetCmd)
	addOutputFlags(accountGetCmd)
	rootCmd.AddCommand(accountCmd)
}
