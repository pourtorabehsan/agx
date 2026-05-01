package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newWaitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wait NAME",
		Short: "Block until a workspace container exits and propagate its exit code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "agx-" + args[0]
			out, err := exec.Command("docker", "wait", name).Output()
			if err != nil {
				return fmt.Errorf("docker wait %s: %w", name, err)
			}
			code, err := strconv.Atoi(strings.TrimSpace(string(out)))
			if err != nil {
				return fmt.Errorf("parse exit code %q: %w", strings.TrimSpace(string(out)), err)
			}
			os.Exit(code)
			return nil
		},
	}
}
