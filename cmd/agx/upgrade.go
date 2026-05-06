package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Reinstall agent skills and instructions from the current image (run 'make build-image' first to pick up image changes)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "hint: to pick up image changes, run 'make build-image' first")
			c := exec.Command("docker", BootstrapArgs(paths.HomeDir, paths.KbDir)...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 125 {
					return fmt.Errorf("docker failed — is the image built? try 'make build-image'")
				}
				return err
			}
			return nil
		},
	}
}
