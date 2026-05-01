package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var follow bool
	var conduct bool
	var phase string

	cmd := &cobra.Command{
		Use:   "logs SLUG",
		Short: "Show logs for a workspace session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			wsPath := paths.Workspace(slug)
			if _, err := os.Stat(wsPath); os.IsNotExist(err) {
				return fmt.Errorf("workspace %q not found — run 'agx ls' to list workspaces", slug)
			}

			// --conduct: print the structured conductor journal
			if conduct {
				return printFile(filepath.Join(wsPath, ".agx", "conduct.md"))
			}

			// --phase: print a specific phase output
			if phase != "" {
				return printFile(filepath.Join(wsPath, ".agx", "phases", phase+"-output.md"))
			}

			// default: session.log
			logPath := filepath.Join(wsPath, ".agx", "session.log")
			if follow {
				c := exec.Command("tail", "-f", logPath)
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				return c.Run()
			}
			return printFile(logPath)
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "tail -f session.log")
	cmd.Flags().BoolVar(&conduct, "conduct", false, "show conductor journal (conduct.md) instead of session log")
	cmd.Flags().StringVar(&phase, "phase", "", "show output for a specific phase (e.g. --phase clone)")
	return cmd
}

func printFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found", filepath.Base(path))
		}
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}
