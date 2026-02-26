package output

import (
	"encoding/json"
	"fmt"
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
}

// TableData holds structured table content.
type TableData struct {
	Headers []string
	Rows    [][]string
}

// RenderTable renders data in the appropriate output mode.
func RenderTable(td TableData, data interface{}, opts Options) error {
	switch opts.Mode {
	case ModeJSON:
		return renderJSON(os.Stdout, data)
	case ModePlaintext:
		return renderPlaintext(os.Stdout, td)
	default:
		return renderTable(os.Stdout, td, opts)
	}
}

// Render outputs raw data as JSON.
func Render(data interface{}, opts Options) error {
	return renderJSON(os.Stdout, data)
}

func renderJSON(w *os.File, data interface{}) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}

func renderPlaintext(w *os.File, td TableData) error {
	if len(td.Headers) > 0 {
		fmt.Fprintln(w, strings.Join(td.Headers, "\t"))
	}
	for _, row := range td.Rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	return nil
}

func renderTable(w *os.File, td TableData, opts Options) error {
	table := tablewriter.NewWriter(w)

	if !opts.NoColor && shouldColor() {
		colored := make([]string, len(td.Headers))
		for i, h := range td.Headers {
			colored[i] = color.New(color.FgCyan, color.Bold).Sprint(h)
		}
		table.SetHeader(colored)
	} else {
		table.SetHeader(td.Headers)
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

	for _, row := range td.Rows {
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

// Error outputs an error message respecting the output mode.
func Error(message string, opts Options) {
	switch opts.Mode {
	case ModeJSON:
		_ = renderJSON(os.Stderr, map[string]string{"error": message})
	default:
		fmt.Fprintf(os.Stderr, "%s %s\n", color.New(color.FgRed).Sprint("Error:"), message)
	}
}

// Success outputs a success message respecting the output mode.
func Success(message string, opts Options) {
	switch opts.Mode {
	case ModeJSON:
		_ = renderJSON(os.Stdout, map[string]string{"status": "success", "message": message})
	default:
		fmt.Printf("%s %s\n", color.New(color.FgGreen).Sprint("OK:"), message)
	}
}
