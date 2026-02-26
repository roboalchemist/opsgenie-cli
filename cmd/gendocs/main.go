package main

import (
	"fmt"
	"log"
	"os"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	rootCmd := cmd.GetRootCmd()

	// Prevent auto-generated timestamps from causing git churn
	rootCmd.DisableAutoGenTag = true

	format := "man"
	outDir := "./docs/man"
	if len(os.Args) > 1 {
		format = os.Args[1]
	}
	if len(os.Args) > 2 {
		outDir = os.Args[2]
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	switch format {
	case "man":
		header := &doc.GenManHeader{
			Title:   "OPSGENIE-CLI",
			Section: "1",
		}
		if err := doc.GenManTree(rootCmd, header, outDir); err != nil {
			log.Fatal(err)
		}
	case "markdown":
		if err := doc.GenMarkdownTree(rootCmd, outDir); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s (use 'man' or 'markdown')\n", format)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Generated %s docs in %s/\n", format, outDir)
}
