package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume NAME",
		Short: "Resume an existing workspace in an interactive session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}

			wsPath := filepath.Join(paths.Workspaces, args[0])
			if _, err := os.Stat(wsPath); os.IsNotExist(err) {
				return fmt.Errorf("workspace %q not found — run 'agx ls' to list workspaces", args[0])
			}

			c := exec.Command("docker", RunArgs(RunConfig{
				HomeDir:       paths.HomeDir,
				KbDir:         paths.KbDir,
				WorkspacePath: wsPath,
				Resume:        true,
			})...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			fmt.Fprintln(os.Stderr, "==> workspace:", wsPath)
			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}
			return nil
		},
	}
}
