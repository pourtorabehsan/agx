package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

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
			if err := os.MkdirAll(paths.HomeDir, 0o755); err != nil {
				return err
			}
			if err := os.MkdirAll(paths.Workspaces, 0o755); err != nil {
				return err
			}
			c := exec.Command("docker", BootstrapArgs(paths.HomeDir)...)
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
