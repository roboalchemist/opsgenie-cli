package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/roboalchemist/opsgenie-cli/pkg/api"
	"github.com/roboalchemist/opsgenie-cli/pkg/auth"
	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long:  "Store and test OpsGenie API credentials.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save an OpsGenie API key to the config file",
	Long: `Save an OpsGenie API key to ~/.opsgenie-cli-auth.json.

The API key can be provided via --api-key flag, OPSGENIE_API_KEY env var,
or entered interactively when prompted.`,
	Example: `  # Save API key interactively
  opsgenie-cli auth login

  # Save API key via flag
  opsgenie-cli auth login --api-key YOUR_API_KEY`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")

		if apiKey == "" {
			apiKey = os.Getenv("OPSGENIE_API_KEY")
		}

		if apiKey == "" {
			fmt.Fprint(os.Stderr, "Enter OpsGenie API key: ")
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read API key: %w", err)
			}
			apiKey = strings.TrimSpace(line)
		}

		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		if err := auth.SaveAPIKey(apiKey); err != nil {
			return fmt.Errorf("failed to save API key: %w", err)
		}

		fmt.Fprintf(os.Stderr, "API key saved to %s\n", auth.ConfigPath())
		return nil
	},
}

var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test current authentication credentials",
	Example: `  # Verify current credentials work
  opsgenie-cli auth test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := getOutputOpts()

		apiKey, err := auth.GetAPIKey()
		if err != nil {
			return err
		}

		client := api.NewClient(apiKey, flagRegion, flagDebug)
		var envelope api.APIResponse[api.AccountResponse]
		if err := client.Get("/v2/account", &envelope); err != nil {
			return fmt.Errorf("authentication test failed: %w", err)
		}

		source := "config file (" + auth.ConfigPath() + ")"
		if os.Getenv("OPSGENIE_API_KEY") != "" {
			source = "OPSGENIE_API_KEY env var"
		}

		headers := []string{"Field", "Value"}
		rows := [][]string{
			{"Account", envelope.Data.Name},
			{"Source", source},
			{"Status", "OK"},
		}
		return output.RenderTable(headers, rows, map[string]string{
			"account": envelope.Data.Name,
			"source":  source,
			"status":  "OK",
		}, opts)
	},
}

func init() {
	authLoginCmd.Flags().String("api-key", "", "OpsGenie API key to save")
	addOutputFlags(authTestCmd)

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authTestCmd)
	rootCmd.AddCommand(authCmd)
}
