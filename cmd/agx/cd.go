package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cd SLUG",
		Short: "Print the workspace path (use with a shell function to cd into it)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			wsPath := paths.Workspace(args[0])
			if _, err := os.Stat(wsPath); os.IsNotExist(err) {
				return fmt.Errorf("workspace %q not found — run 'agx ps' to list workspaces", args[0])
			}
			fmt.Println(wsPath)
			return nil
		},
	}
}
