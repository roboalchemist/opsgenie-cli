package cmd

import (
	"fmt"
	"os"

	"gitea.roboalch.com/roboalchemist/opsgenie-cli/pkg/output"
	"github.com/spf13/cobra"
)

var appVersion = "dev"

// Global flag values
var (
	flagJSON      bool
	flagPlaintext bool
	flagNoColor   bool
	flagDebug     bool
	flagRegion    string
)

var rootCmd = &cobra.Command{
	Use:           "opsgenie-cli",
	Short:         "CLI for the OpsGenie REST API v2",
	Long:          "CLI for the OpsGenie REST API v2. Manage alerts, incidents, teams, schedules, and more.",
	Version:       appVersion,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.BoolVarP(&flagJSON, "json", "j", false, "JSON output")
	pf.BoolVarP(&flagPlaintext, "plaintext", "p", false, "Tab-separated output for piping")
	pf.BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	pf.BoolVar(&flagDebug, "debug", false, "Verbose logging to stderr")
	pf.StringVar(&flagRegion, "region", "us", "OpsGenie region (us or eu)")
}

// GetOutputOptions builds output.Options from global flags.
func GetOutputOptions() output.Options {
	opts := output.Options{
		NoColor: flagNoColor,
		Debug:   flagDebug,
	}
	switch {
	case flagJSON:
		opts.Mode = output.ModeJSON
	case flagPlaintext:
		opts.Mode = output.ModePlaintext
	default:
		opts.Mode = output.ModeTable
	}
	return opts
}

// GetRegion returns the configured region flag value.
func GetRegion() string {
	return flagRegion
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the application version on the root command.
func SetVersion(v string) {
	appVersion = v
	rootCmd.Version = v
}

// GetRootCmd returns the root command (used by cmd/gendocs).
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// DebugLog logs a debug message to stderr if --debug is set.
func DebugLog(format string, args ...interface{}) {
	if flagDebug {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}
