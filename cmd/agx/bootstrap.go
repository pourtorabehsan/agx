package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

const kbIndexSeed = `# Knowledge Base Index

Entries are added automatically during sessions. Format:
- [title](path/file.md) — one-line description of what's in the file

## Categories

- decisions/ — architectural and technical choices with their reasoning
- conventions/ — repo conventions, naming rules, patterns to follow
- gotchas/ — non-obvious pitfalls, things that burned us before
- runbooks/ — step-by-step procedures for recurring operations
- repos/ — known repositories, their purpose, default branch, and relationships
- architecture/ — system design, service boundaries, data flows, and infrastructure
`

func newBootstrapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Set up credentials and config inside the container",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			for _, dir := range []string{paths.HomeDir, paths.Workspaces, paths.KbDir} {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return err
				}
			}
			indexPath := paths.KbDir + "/INDEX.md"
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				if err := os.WriteFile(indexPath, []byte(kbIndexSeed), 0o644); err != nil {
					return err
				}
			}
			c := exec.Command("docker", BootstrapArgs(paths.HomeDir, paths.KbDir)...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 125 {
					return fmt.Errorf("docker failed — is the image built? try 'make build'")
				}
				return err
			}
			return nil
		},
	}
}
