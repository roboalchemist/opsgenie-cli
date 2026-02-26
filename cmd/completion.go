package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for opsgenie-cli.

To load completions:

Bash:
  $ source <(opsgenie-cli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ opsgenie-cli completion bash > /etc/bash_completion.d/opsgenie-cli
  # macOS:
  $ opsgenie-cli completion bash > $(brew --prefix)/etc/bash_completion.d/opsgenie-cli

Zsh:
  $ source <(opsgenie-cli completion zsh)

  # To load completions for each session, execute once:
  $ opsgenie-cli completion zsh > "${fpath[1]}/_opsgenie-cli"

Fish:
  $ opsgenie-cli completion fish | source

  # To load completions for each session, execute once:
  $ opsgenie-cli completion fish > ~/.config/fish/completions/opsgenie-cli.fish

PowerShell:
  PS> opsgenie-cli completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
