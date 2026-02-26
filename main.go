package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/roboalchemist/opsgenie-cli/cmd"
)

// version is set via ldflags at build time: -X main.version=x.y.z
var version = "dev"

//go:embed README.md
var readmeContents string

//go:embed skill/SKILL.md
var skillMD string

//go:embed skill/reference/commands.md
var commandsRef string

//go:embed skill
var skillFS embed.FS

func main() {
	cmd.SetVersion(version)
	cmd.SetReadmeContents(readmeContents)
	cmd.SetSkillData(skillMD, commandsRef, skillFS)
	if err := cmd.Execute(); err != nil {
		if cmd.IsJSON() {
			errJSON, _ := json.Marshal(map[string]interface{}{
				"code":        "ERROR",
				"message":     err.Error(),
				"recoverable": false,
			})
			fmt.Fprintln(os.Stderr, string(errJSON))
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
