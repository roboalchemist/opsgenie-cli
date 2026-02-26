package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	embeddedSkillMD     string
	embeddedCommandsRef string
	embeddedSkillFS     embed.FS
)

// SetSkillData sets the embedded skill content from main package.
func SetSkillData(skillMD, commandsRef string, skillFS embed.FS) {
	embeddedSkillMD = skillMD
	embeddedCommandsRef = commandsRef
	embeddedSkillFS = skillFS
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage the embedded Claude Code skill",
}

var skillPrintCmd = &cobra.Command{
	Use:   "print",
	Short: "Print the embedded SKILL.md to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(embeddedSkillMD)
		return nil
	},
}

var skillAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Install skill to ~/.claude/skills/opsgenie-cli/",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}

		destDir := filepath.Join(home, ".claude", "skills", "opsgenie-cli")

		err = fs.WalkDir(embeddedSkillFS, "skill", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel("skill", path)
			dest := filepath.Join(destDir, rel)

			if d.IsDir() {
				return os.MkdirAll(dest, 0o755)
			}

			data, err := fs.ReadFile(embeddedSkillFS, path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			return os.WriteFile(dest, data, 0o644)
		})

		if err != nil {
			return fmt.Errorf("install skill: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Installed skill to %s\n", destDir)
		return nil
	},
}

func init() {
	skillCmd.AddCommand(skillPrintCmd)
	skillCmd.AddCommand(skillAddCmd)
	rootCmd.AddCommand(skillCmd)
}
