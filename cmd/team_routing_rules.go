package cmd

import (
	"encoding/json"
	"fmt"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var teamRoutingRulesCmd = &cobra.Command{
	Use:   "team-routing-rules",
	Short: "Manage team routing rules",
}

var teamRoutingRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List routing rules for a team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		teamID, _ := cmd.Flags().GetString("team")
		if teamID == "" {
			return fmt.Errorf("--team is required")
		}

		var resp struct {
			Data []api.TeamRoutingRuleResponse `json:"data"`
		}
		if err := client.Get("/v2/teams/"+teamID+"/routing-rules", &resp); err != nil {
			return err
		}

		if opts.Mode == output.ModeJSON {
			return output.RenderJSON(resp.Data, opts)
		}

		headers := []string{"ID", "Name", "Order", "Type"}
		rows := make([][]string, len(resp.Data))
		for i, r := range resp.Data {
			rows[i] = []string{r.ID, r.Name, fmt.Sprintf("%d", r.Order), r.Type}
		}
		return output.RenderTable(headers, rows, resp.Data, opts)
	},
}

var teamRoutingRulesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a routing rule by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		teamID, _ := cmd.Flags().GetString("team")
		ruleID, _ := cmd.Flags().GetString("id")
		if teamID == "" {
			return fmt.Errorf("--team is required")
		}
		if ruleID == "" {
			return fmt.Errorf("--id is required")
		}

		var resp struct {
			Data api.TeamRoutingRuleResponse `json:"data"`
		}
		if err := client.Get("/v2/teams/"+teamID+"/routing-rules/"+ruleID, &resp); err != nil {
			return err
		}

		return output.RenderJSON(resp.Data, opts)
	},
}

var teamRoutingRulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a routing rule for a team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		teamID, _ := cmd.Flags().GetString("team")
		name, _ := cmd.Flags().GetString("name")
		ruleType, _ := cmd.Flags().GetString("type")
		notifyJSON, _ := cmd.Flags().GetString("notify")

		if teamID == "" {
			return fmt.Errorf("--team is required")
		}

		body := map[string]interface{}{
			"name": name,
			"type": ruleType,
		}

		if notifyJSON != "" {
			var notify interface{}
			if err := json.Unmarshal([]byte(notifyJSON), &notify); err != nil {
				return fmt.Errorf("invalid --notify JSON: %w", err)
			}
			body["notify"] = notify
		}

		var resp struct {
			Data api.TeamRoutingRuleResponse `json:"data"`
		}
		if err := client.Post("/v2/teams/"+teamID+"/routing-rules", body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Routing rule created for team %s", teamID), opts)
		return output.RenderJSON(resp.Data, opts)
	},
}

var teamRoutingRulesUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a routing rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		teamID, _ := cmd.Flags().GetString("team")
		ruleID, _ := cmd.Flags().GetString("id")
		if teamID == "" {
			return fmt.Errorf("--team is required")
		}
		if ruleID == "" {
			return fmt.Errorf("--id is required")
		}

		body := map[string]interface{}{}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			body["name"] = name
		}
		if ruleType, _ := cmd.Flags().GetString("type"); ruleType != "" {
			body["type"] = ruleType
		}
		if notifyJSON, _ := cmd.Flags().GetString("notify"); notifyJSON != "" {
			var notify interface{}
			if err := json.Unmarshal([]byte(notifyJSON), &notify); err != nil {
				return fmt.Errorf("invalid --notify JSON: %w", err)
			}
			body["notify"] = notify
		}

		var resp struct {
			Data api.TeamRoutingRuleResponse `json:"data"`
		}
		if err := client.Patch("/v2/teams/"+teamID+"/routing-rules/"+ruleID, body, &resp); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Routing rule %s updated", ruleID), opts)
		return nil
	},
}

var teamRoutingRulesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a routing rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		teamID, _ := cmd.Flags().GetString("team")
		ruleID, _ := cmd.Flags().GetString("id")
		if teamID == "" {
			return fmt.Errorf("--team is required")
		}
		if ruleID == "" {
			return fmt.Errorf("--id is required")
		}

		var result json.RawMessage
		if err := client.Delete("/v2/teams/"+teamID+"/routing-rules/"+ruleID, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Routing rule %s deleted", ruleID), opts)
		return nil
	},
}

func init() {
	teamRoutingRulesListCmd.Flags().String("team", "", "Team ID or name (required)")
	addOutputFlags(teamRoutingRulesListCmd)

	teamRoutingRulesGetCmd.Flags().String("team", "", "Team ID or name (required)")
	teamRoutingRulesGetCmd.Flags().String("id", "", "Routing rule ID (required)")
	addOutputFlags(teamRoutingRulesGetCmd)

	teamRoutingRulesCreateCmd.Flags().String("team", "", "Team ID or name (required)")
	teamRoutingRulesCreateCmd.Flags().String("name", "", "Rule name")
	teamRoutingRulesCreateCmd.Flags().String("type", "", "Rule type (e.g. schedule, escalation)")
	teamRoutingRulesCreateCmd.Flags().String("notify", "", "JSON object for notify config")

	teamRoutingRulesUpdateCmd.Flags().String("team", "", "Team ID or name (required)")
	teamRoutingRulesUpdateCmd.Flags().String("id", "", "Routing rule ID (required)")
	teamRoutingRulesUpdateCmd.Flags().String("name", "", "New rule name")
	teamRoutingRulesUpdateCmd.Flags().String("type", "", "New rule type")
	teamRoutingRulesUpdateCmd.Flags().String("notify", "", "JSON object for notify config")

	teamRoutingRulesDeleteCmd.Flags().String("team", "", "Team ID or name (required)")
	teamRoutingRulesDeleteCmd.Flags().String("id", "", "Routing rule ID (required)")

	teamRoutingRulesCmd.AddCommand(teamRoutingRulesListCmd)
	teamRoutingRulesCmd.AddCommand(teamRoutingRulesGetCmd)
	teamRoutingRulesCmd.AddCommand(teamRoutingRulesCreateCmd)
	teamRoutingRulesCmd.AddCommand(teamRoutingRulesUpdateCmd)
	teamRoutingRulesCmd.AddCommand(teamRoutingRulesDeleteCmd)

	rootCmd.AddCommand(teamRoutingRulesCmd)
}
