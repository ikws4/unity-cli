package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/ikws4/unity-cli/cmd"
)

var Version = "dev"

// Keep agent guidance in lockstep with the binary that exposes the commands.
//
//go:embed skills/*/SKILL.md
var embeddedContent embed.FS

func init() {
	cmd.Version = Version
	if skills, err := fs.Sub(embeddedContent, "skills"); err == nil {
		cmd.SetEmbeddedSkillContent(skills)
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
