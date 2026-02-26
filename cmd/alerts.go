package cmd

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/api"
	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

// alertsCmd is the parent command for all alert operations.
var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage OpsGenie alerts",
}

// ─── alerts list ─────────────────────────────────────────────────────────────

var (
	alertsListLimit  int
	alertsListOffset int
	alertsListQuery  string
	alertsListSort   string
)

var alertsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		params := url.Values{}
		if alertsListLimit > 0 {
			params.Set("limit", strconv.Itoa(alertsListLimit))
		}
		if alertsListOffset > 0 {
			params.Set("offset", strconv.Itoa(alertsListOffset))
		}
		if alertsListQuery != "" {
			params.Set("query", alertsListQuery)
		}
		if alertsListSort != "" {
			params.Set("sort", alertsListSort)
		}

		var alerts []api.AlertResponse
		if err := client.ListAll("/v2/alerts", params, &alerts); err != nil {
			return err
		}

		headers := []string{"ID", "Message", "Status", "Priority", "Acknowledged", "CreatedAt"}
		rows := make([][]string, len(alerts))
		for i, a := range alerts {
			rows[i] = []string{
				a.ID,
				a.Message,
				a.Status,
				a.Priority,
				strconv.FormatBool(a.Acknowledged),
				a.CreatedAt,
			}
		}
		return output.RenderTable(headers, rows, alerts, opts)
	},
}

func init() {
	alertsCmd.AddCommand(alertsListCmd)
	addOutputFlags(alertsListCmd)
	alertsListCmd.Flags().IntVar(&alertsListLimit, "limit", 0, "Maximum number of alerts to return (0 = all)")
	alertsListCmd.Flags().IntVar(&alertsListOffset, "offset", 0, "Start offset for pagination")
	alertsListCmd.Flags().StringVar(&alertsListQuery, "query", "", "Search query (OpsGenie query syntax)")
	alertsListCmd.Flags().StringVar(&alertsListSort, "sort", "", "Sort field (e.g. createdAt, updatedAt)")
}

// ─── alerts get ──────────────────────────────────────────────────────────────

var alertsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an alert by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var envelope api.APIResponse[api.AlertResponse]
		if err := client.Get("/v2/alerts/"+args[0]+"?identifierType=id", &envelope); err != nil {
			return err
		}
		a := envelope.Data

		headers := []string{"Field", "Value"}
		rows := [][]string{
			{"ID", a.ID},
			{"TinyID", a.TinyID},
			{"Alias", a.Alias},
			{"Message", a.Message},
			{"Status", a.Status},
			{"Priority", a.Priority},
			{"Acknowledged", strconv.FormatBool(a.Acknowledged)},
			{"Snoozed", strconv.FormatBool(a.Snoozed)},
			{"IsSeen", strconv.FormatBool(a.IsSeen)},
			{"Source", a.Source},
			{"Owner", a.Owner},
			{"Tags", strings.Join(a.Tags, ", ")},
			{"Count", strconv.Itoa(a.Count)},
			{"CreatedAt", a.CreatedAt},
			{"UpdatedAt", a.UpdatedAt},
			{"ClosedAt", a.ClosedAt},
		}
		return output.RenderTable(headers, rows, a, opts)
	},
}

func init() {
	alertsCmd.AddCommand(alertsGetCmd)
	addOutputFlags(alertsGetCmd)
}

// ─── alerts create ───────────────────────────────────────────────────────────

var (
	alertCreateMessage     string
	alertCreateDescription string
	alertCreatePriority    string
	alertCreateTags        string
	alertCreateResponders  string
)

var alertsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new alert",
	RunE: func(cmd *cobra.Command, args []string) error {
		if alertCreateMessage == "" {
			return fmt.Errorf("--message is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}

		body := map[string]interface{}{
			"message": alertCreateMessage,
		}
		if alertCreateDescription != "" {
			body["description"] = alertCreateDescription
		}
		if alertCreatePriority != "" {
			body["priority"] = alertCreatePriority
		}
		if alertCreateTags != "" {
			tags := splitAndTrim(alertCreateTags)
			body["tags"] = tags
		}
		if alertCreateResponders != "" {
			body["responders"] = parseResponders(alertCreateResponders)
		}

		var result map[string]interface{}
		if err := client.Post("/v2/alerts", body, &result); err != nil {
			return err
		}

		opts := GetOutputOptions()
		output.Success("Alert created", opts)
		if result != nil {
			return output.RenderJSON(result, opts)
		}
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsCreateCmd)
	alertsCreateCmd.Flags().StringVar(&alertCreateMessage, "message", "", "Alert message (required)")
	alertsCreateCmd.Flags().StringVar(&alertCreateDescription, "description", "", "Alert description")
	alertsCreateCmd.Flags().StringVar(&alertCreatePriority, "priority", "", "Priority (P1-P5)")
	alertsCreateCmd.Flags().StringVar(&alertCreateTags, "tags", "", "Comma-separated tags")
	alertsCreateCmd.Flags().StringVar(&alertCreateResponders, "responders", "", "Comma-separated responders (e.g. team:myteam,user:user@example.com)")
}

// ─── alerts delete ───────────────────────────────────────────────────────────

var alertsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an alert",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		if err := client.Delete("/v2/alerts/"+args[0]+"?identifierType=id", nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Alert deleted", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsDeleteCmd)
}

// ─── alerts acknowledge ───────────────────────────────────────────────────────

var alertsAcknowledgeCmd = &cobra.Command{
	Use:   "acknowledge <id>",
	Short: "Acknowledge an alert",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		if err := client.Post("/v2/alerts/"+args[0]+"/acknowledge", map[string]interface{}{}, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Alert acknowledged", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsAcknowledgeCmd)
}

// ─── alerts close ─────────────────────────────────────────────────────────────

var alertsCloseNote string

var alertsCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close an alert",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{}
		if alertsCloseNote != "" {
			body["note"] = alertsCloseNote
		}
		if err := client.Post("/v2/alerts/"+args[0]+"/close", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Alert closed", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsCloseCmd)
	alertsCloseCmd.Flags().StringVar(&alertsCloseNote, "note", "", "Note to add when closing")
}

// ─── alerts snooze ───────────────────────────────────────────────────────────

var alertsSnoozeEndTime string

var alertsSnoozeCmd = &cobra.Command{
	Use:   "snooze <id>",
	Short: "Snooze an alert until a given time",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if alertsSnoozeEndTime == "" {
			return fmt.Errorf("--end-time is required (RFC3339, e.g. 2024-01-15T10:00:00Z)")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"endTime": alertsSnoozeEndTime,
		}
		if err := client.Post("/v2/alerts/"+args[0]+"/snooze", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Alert snoozed", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsSnoozeCmd)
	alertsSnoozeCmd.Flags().StringVar(&alertsSnoozeEndTime, "end-time", "", "Snooze until this time (RFC3339)")
}

// ─── alerts escalate ─────────────────────────────────────────────────────────

var alertsEscalateEscalation string

var alertsEscalateCmd = &cobra.Command{
	Use:   "escalate <id>",
	Short: "Escalate an alert to an escalation policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if alertsEscalateEscalation == "" {
			return fmt.Errorf("--escalation is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"escalation": map[string]interface{}{
				"name": alertsEscalateEscalation,
			},
		}
		if err := client.Post("/v2/alerts/"+args[0]+"/escalate", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Alert escalated", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsEscalateCmd)
	alertsEscalateCmd.Flags().StringVar(&alertsEscalateEscalation, "escalation", "", "Escalation policy name")
}

// ─── alerts assign ────────────────────────────────────────────────────────────

var alertsAssignOwner string

var alertsAssignCmd = &cobra.Command{
	Use:   "assign <id>",
	Short: "Assign an alert to an owner",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if alertsAssignOwner == "" {
			return fmt.Errorf("--owner is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"owner": map[string]interface{}{
				"username": alertsAssignOwner,
			},
		}
		if err := client.Post("/v2/alerts/"+args[0]+"/assign", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Alert assigned", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsAssignCmd)
	alertsAssignCmd.Flags().StringVar(&alertsAssignOwner, "owner", "", "Username of the new owner")
}

// ─── alerts add-note ─────────────────────────────────────────────────────────

var alertsAddNoteNote string

var alertsAddNoteCmd = &cobra.Command{
	Use:   "add-note <id>",
	Short: "Add a note to an alert",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if alertsAddNoteNote == "" {
			return fmt.Errorf("--note is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"note": alertsAddNoteNote,
		}
		if err := client.Post("/v2/alerts/"+args[0]+"/notes", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Note added", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsAddNoteCmd)
	alertsAddNoteCmd.Flags().StringVar(&alertsAddNoteNote, "note", "", "Note text (required)")
}

// ─── alerts add-tags ─────────────────────────────────────────────────────────

var alertsAddTagsTags string

var alertsAddTagsCmd = &cobra.Command{
	Use:   "add-tags <id>",
	Short: "Add tags to an alert",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if alertsAddTagsTags == "" {
			return fmt.Errorf("--tags is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"tags": splitAndTrim(alertsAddTagsTags),
		}
		if err := client.Post("/v2/alerts/"+args[0]+"/tags", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Tags added", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsAddTagsCmd)
	alertsAddTagsCmd.Flags().StringVar(&alertsAddTagsTags, "tags", "", "Comma-separated tags to add (required)")
}

// ─── alerts remove-tags ───────────────────────────────────────────────────────

var alertsRemoveTagsTags string

var alertsRemoveTagsCmd = &cobra.Command{
	Use:   "remove-tags <id>",
	Short: "Remove tags from an alert",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if alertsRemoveTagsTags == "" {
			return fmt.Errorf("--tags is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		tagList := splitAndTrim(alertsRemoveTagsTags)
		path := "/v2/alerts/" + args[0] + "/tags?identifierType=id&tags=" + url.QueryEscape(strings.Join(tagList, ","))
		if err := client.Delete(path, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Tags removed", opts)
		return nil
	},
}

func init() {
	alertsCmd.AddCommand(alertsRemoveTagsCmd)
	alertsRemoveTagsCmd.Flags().StringVar(&alertsRemoveTagsTags, "tags", "", "Comma-separated tags to remove (required)")
}

// ─── alerts count ─────────────────────────────────────────────────────────────

var alertsCountQuery string

var alertsCountCmd = &cobra.Command{
	Use:   "count",
	Short: "Count alerts matching a query",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		path := "/v2/alerts/count"
		if alertsCountQuery != "" {
			path += "?query=" + url.QueryEscape(alertsCountQuery)
		}

		var envelope api.APIResponse[api.AlertCountResponse]
		if err := client.Get(path, &envelope); err != nil {
			return err
		}

		headers := []string{"Count"}
		rows := [][]string{
			{strconv.Itoa(envelope.Data.Count)},
		}
		return output.RenderTable(headers, rows, envelope.Data, opts)
	},
}

func init() {
	alertsCmd.AddCommand(alertsCountCmd)
	addOutputFlags(alertsCountCmd)
	alertsCountCmd.Flags().StringVar(&alertsCountQuery, "query", "", "Search query to count matching alerts")
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// splitAndTrim splits a comma-separated string and trims whitespace from each element.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// parseResponders parses a comma-separated list of "type:name" pairs into Responder objects.
// Example: "team:ops,user:alice@example.com"
func parseResponders(s string) []map[string]string {
	parts := splitAndTrim(s)
	responders := make([]map[string]string, 0, len(parts))
	for _, p := range parts {
		idx := strings.Index(p, ":")
		if idx < 0 {
			// Default to team if no type prefix
			responders = append(responders, map[string]string{"name": p, "type": "team"})
			continue
		}
		rType := p[:idx]
		rName := p[idx+1:]
		responders = append(responders, map[string]string{"name": rName, "type": rType})
	}
	return responders
}

func init() {
	rootCmd.AddCommand(alertsCmd)
}
