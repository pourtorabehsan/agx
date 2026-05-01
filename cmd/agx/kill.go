package main

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func newKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill NAME",
		Short: "Stop a running workspace container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "agx-" + args[0]
			out, err := exec.Command("docker", "kill", name).CombinedOutput()
			if err != nil {
				return fmt.Errorf("docker kill %s: %w\n%s", name, err, string(out))
			}
			return nil
		},
	}
}
