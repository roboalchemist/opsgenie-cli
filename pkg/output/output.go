package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// Mode represents the output rendering mode.
type Mode int

const (
	ModeTable     Mode = iota // Default: colored table
	ModePlaintext             // Tab-separated, no colors
	ModeJSON                  // Pretty-printed JSON
)

// Options controls output rendering behavior.
type Options struct {
	Mode    Mode
	NoColor bool
	Debug   bool
	Fields  []string // If set, filter JSON output to only these fields
	JQExpr  string   // If set, apply this jq expression to JSON output
}

// RenderTable renders data in the appropriate output mode.
// headers and rows are used for table/plaintext modes; rawData is used for JSON mode.
func RenderTable(headers []string, rows [][]string, rawData interface{}, opts Options) error {
	switch opts.Mode {
	case ModeJSON:
		return RenderJSON(rawData, opts)
	case ModePlaintext:
		return renderPlaintext(os.Stdout, headers, rows)
	default:
		return renderTable(os.Stdout, headers, rows, opts)
	}
}

// RenderJSON outputs data as JSON with optional fields filtering and jq evaluation.
func RenderJSON(data interface{}, opts Options) error {
	return renderJSONTo(os.Stdout, data, opts)
}

// renderJSONTo writes JSON output to w, applying fields filtering and jq expressions.
func renderJSONTo(w io.Writer, data interface{}, opts Options) error {
	// Apply fields filtering if specified
	if len(opts.Fields) > 0 {
		filtered, err := filterFields(data, opts.Fields)
		if err != nil {
			return fmt.Errorf("fields filter: %w", err)
		}
		data = filtered
	}

	// Apply jq expression if specified
	if opts.JQExpr != "" {
		result, err := applyJQ(data, opts.JQExpr)
		if err != nil {
			return fmt.Errorf("jq expression: %w", err)
		}
		data = result
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}

func renderPlaintext(w io.Writer, headers []string, rows [][]string) error {
	if len(headers) > 0 {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
	}
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	return nil
}

func renderTable(w io.Writer, headers []string, rows [][]string, opts Options) error {
	table := tablewriter.NewWriter(w)

	if !opts.NoColor && shouldColor() {
		colored := make([]string, len(headers))
		for i, h := range headers {
			colored[i] = color.New(color.FgCyan, color.Bold).Sprint(h)
		}
		table.SetHeader(colored)
	} else {
		table.SetHeader(headers)
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("   ")
	table.SetNoWhiteSpace(true)

	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
	return nil
}

func shouldColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// Error outputs an error message to stderr in red (unless NoColor is set).
func Error(msg string, opts ...Options) {
	noColor := false
	if len(opts) > 0 {
		noColor = opts[0].NoColor
	}
	if noColor || os.Getenv("NO_COLOR") != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", color.New(color.FgRed).Sprint("Error:"), msg)
	}
}

// Success outputs a success message to stderr in green (unless NoColor is set).
func Success(msg string, opts ...Options) {
	noColor := false
	if len(opts) > 0 {
		noColor = opts[0].NoColor
	}
	if noColor || os.Getenv("NO_COLOR") != "" {
		fmt.Fprintf(os.Stderr, "OK: %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", color.New(color.FgGreen).Sprint("OK:"), msg)
	}
}
