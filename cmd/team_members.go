package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var teamMembersCmd = &cobra.Command{
	Use:   "team-members",
	Short: "Manage team members",
}

var teamMembersAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a member to a team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		teamID, _ := cmd.Flags().GetString("team")
		userID, _ := cmd.Flags().GetString("user")
		role, _ := cmd.Flags().GetString("role")

		if teamID == "" {
			return fmt.Errorf("--team is required")
		}
		if userID == "" {
			return fmt.Errorf("--user is required")
		}

		body := map[string]interface{}{
			"user": map[string]string{"id": userID},
		}
		if role != "" {
			body["role"] = role
		}

		var result json.RawMessage
		if err := client.Post("/v2/teams/"+teamID+"/members", body, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("User %s added to team %s", userID, teamID), opts)
		return nil
	},
}

var teamMembersRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a member from a team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		teamID, _ := cmd.Flags().GetString("team")
		memberID, _ := cmd.Flags().GetString("member")

		if teamID == "" {
			return fmt.Errorf("--team is required")
		}
		if memberID == "" {
			return fmt.Errorf("--member is required")
		}

		var result json.RawMessage
		if err := client.Delete("/v2/teams/"+teamID+"/members/"+memberID, &result); err != nil {
			return err
		}

		output.Success(fmt.Sprintf("Member %s removed from team %s", memberID, teamID), opts)
		return nil
	},
}

func init() {
	teamMembersAddCmd.Flags().String("team", "", "Team ID or name (required)")
	teamMembersAddCmd.Flags().String("user", "", "User ID or username (required)")
	teamMembersAddCmd.Flags().String("role", "", "Member role (e.g. admin, user)")

	teamMembersRemoveCmd.Flags().String("team", "", "Team ID or name (required)")
	teamMembersRemoveCmd.Flags().String("member", "", "Member ID or username (required)")

	teamMembersCmd.AddCommand(teamMembersAddCmd)
	teamMembersCmd.AddCommand(teamMembersRemoveCmd)

	rootCmd.AddCommand(teamMembersCmd)
}
