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

// incidentsCmd is the parent command for all incident operations.
// NOTE: Incidents use /v1/incidents (not /v2).
var incidentsCmd = &cobra.Command{
	Use:   "incidents",
	Short: "Manage OpsGenie incidents",
}

// ─── incidents list ───────────────────────────────────────────────────────────

var (
	incidentsListLimit  int
	incidentsListOffset int
	incidentsListQuery  string
	incidentsListSort   string
	incidentsListOrder  string
)

var incidentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List incidents",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		params := url.Values{}
		if incidentsListLimit > 0 {
			params.Set("limit", strconv.Itoa(incidentsListLimit))
		}
		if incidentsListOffset > 0 {
			params.Set("offset", strconv.Itoa(incidentsListOffset))
		}
		if incidentsListQuery != "" {
			params.Set("query", incidentsListQuery)
		}
		if incidentsListSort != "" {
			params.Set("sort", incidentsListSort)
		}
		if incidentsListOrder != "" {
			params.Set("order", incidentsListOrder)
		}

		var incidents []api.IncidentResponse
		if err := client.ListAll("/v1/incidents", params, &incidents); err != nil {
			return err
		}

		headers := []string{"ID", "Message", "Status", "Priority", "Owner", "CreatedAt"}
		rows := make([][]string, len(incidents))
		for i, inc := range incidents {
			rows[i] = []string{
				inc.ID,
				inc.Message,
				inc.Status,
				inc.Priority,
				inc.Owner,
				inc.CreatedAt,
			}
		}
		return output.RenderTable(headers, rows, incidents, opts)
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsListCmd)
	addOutputFlags(incidentsListCmd)
	incidentsListCmd.Flags().IntVar(&incidentsListLimit, "limit", 0, "Maximum number of incidents to return (0 = all)")
	incidentsListCmd.Flags().IntVar(&incidentsListOffset, "offset", 0, "Start offset for pagination")
	incidentsListCmd.Flags().StringVar(&incidentsListQuery, "query", "", "Search query (OpsGenie query syntax)")
	incidentsListCmd.Flags().StringVar(&incidentsListSort, "sort", "", "Sort field (e.g. createdAt, updatedAt)")
	incidentsListCmd.Flags().StringVar(&incidentsListOrder, "order", "", "Sort order: asc or desc")
}

// ─── incidents get ────────────────────────────────────────────────────────────

var incidentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an incident by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		opts := getOutputOpts()

		var envelope api.APIResponse[api.IncidentResponse]
		if err := client.Get("/v1/incidents/"+args[0], &envelope); err != nil {
			return err
		}
		inc := envelope.Data

		headers := []string{"Field", "Value"}
		rows := [][]string{
			{"ID", inc.ID},
			{"TinyID", inc.TinyID},
			{"Message", inc.Message},
			{"Description", inc.Description},
			{"Status", inc.Status},
			{"Priority", inc.Priority},
			{"Owner", inc.Owner},
			{"Tags", strings.Join(inc.Tags, ", ")},
			{"CreatedAt", inc.CreatedAt},
			{"UpdatedAt", inc.UpdatedAt},
		}
		return output.RenderTable(headers, rows, inc, opts)
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsGetCmd)
	addOutputFlags(incidentsGetCmd)
}

// ─── incidents create ─────────────────────────────────────────────────────────

var (
	incidentCreateMessage     string
	incidentCreateDescription string
	incidentCreatePriority    string
	incidentCreateTags        string
	incidentCreateResponders  string
)

var incidentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new incident",
	RunE: func(cmd *cobra.Command, args []string) error {
		if incidentCreateMessage == "" {
			return fmt.Errorf("--message is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}

		body := map[string]interface{}{
			"message": incidentCreateMessage,
		}
		if incidentCreateDescription != "" {
			body["description"] = incidentCreateDescription
		}
		if incidentCreatePriority != "" {
			body["priority"] = incidentCreatePriority
		}
		if incidentCreateTags != "" {
			body["tags"] = splitAndTrim(incidentCreateTags)
		}
		if incidentCreateResponders != "" {
			body["responders"] = parseResponders(incidentCreateResponders)
		}

		var result map[string]interface{}
		if err := client.Post("/v1/incidents", body, &result); err != nil {
			return err
		}

		opts := GetOutputOptions()
		output.Success("Incident created", opts)
		if result != nil {
			return output.RenderJSON(result, opts)
		}
		return nil
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsCreateCmd)
	incidentsCreateCmd.Flags().StringVar(&incidentCreateMessage, "message", "", "Incident message (required)")
	incidentsCreateCmd.Flags().StringVar(&incidentCreateDescription, "description", "", "Incident description")
	incidentsCreateCmd.Flags().StringVar(&incidentCreatePriority, "priority", "", "Priority (P1-P5)")
	incidentsCreateCmd.Flags().StringVar(&incidentCreateTags, "tags", "", "Comma-separated tags")
	incidentsCreateCmd.Flags().StringVar(&incidentCreateResponders, "responders", "", "Comma-separated responders (e.g. team:myteam,user:user@example.com)")
}

// ─── incidents close ──────────────────────────────────────────────────────────

var incidentsCloseNote string

var incidentsCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close an incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{}
		if incidentsCloseNote != "" {
			body["note"] = incidentsCloseNote
		}
		if err := client.Post("/v1/incidents/"+args[0]+"/close", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Incident closed", opts)
		return nil
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsCloseCmd)
	incidentsCloseCmd.Flags().StringVar(&incidentsCloseNote, "note", "", "Note to add when closing")
}

// ─── incidents resolve ────────────────────────────────────────────────────────

var incidentsResolveNote string

var incidentsResolveCmd = &cobra.Command{
	Use:   "resolve <id>",
	Short: "Resolve an incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{}
		if incidentsResolveNote != "" {
			body["note"] = incidentsResolveNote
		}
		if err := client.Post("/v1/incidents/"+args[0]+"/resolve", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Incident resolved", opts)
		return nil
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsResolveCmd)
	incidentsResolveCmd.Flags().StringVar(&incidentsResolveNote, "note", "", "Note to add when resolving")
}

// ─── incidents reopen ─────────────────────────────────────────────────────────

var incidentsReopenNote string

var incidentsReopenCmd = &cobra.Command{
	Use:   "reopen <id>",
	Short: "Reopen a closed or resolved incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{}
		if incidentsReopenNote != "" {
			body["note"] = incidentsReopenNote
		}
		if err := client.Post("/v1/incidents/"+args[0]+"/reopen", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Incident reopened", opts)
		return nil
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsReopenCmd)
	incidentsReopenCmd.Flags().StringVar(&incidentsReopenNote, "note", "", "Note to add when reopening")
}

// ─── incidents delete ─────────────────────────────────────────────────────────

var incidentsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		if err := client.Delete("/v1/incidents/"+args[0], nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Incident deleted", opts)
		return nil
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsDeleteCmd)
}

// ─── incidents add-note ───────────────────────────────────────────────────────

var incidentsAddNoteNote string

var incidentsAddNoteCmd = &cobra.Command{
	Use:   "add-note <id>",
	Short: "Add a note to an incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if incidentsAddNoteNote == "" {
			return fmt.Errorf("--note is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"note": incidentsAddNoteNote,
		}
		if err := client.Post("/v1/incidents/"+args[0]+"/notes", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Note added", opts)
		return nil
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsAddNoteCmd)
	incidentsAddNoteCmd.Flags().StringVar(&incidentsAddNoteNote, "note", "", "Note text (required)")
}

// ─── incidents add-tags ───────────────────────────────────────────────────────

var incidentsAddTagsTags string

var incidentsAddTagsCmd = &cobra.Command{
	Use:   "add-tags <id>",
	Short: "Add tags to an incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if incidentsAddTagsTags == "" {
			return fmt.Errorf("--tags is required")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"tags": splitAndTrim(incidentsAddTagsTags),
		}
		if err := client.Post("/v1/incidents/"+args[0]+"/tags", body, nil); err != nil {
			return err
		}
		opts := GetOutputOptions()
		output.Success("Tags added", opts)
		return nil
	},
}

func init() {
	incidentsCmd.AddCommand(incidentsAddTagsCmd)
	incidentsAddTagsCmd.Flags().StringVar(&incidentsAddTagsTags, "tags", "", "Comma-separated tags to add (required)")
}

func init() {
	rootCmd.AddCommand(incidentsCmd)
}
