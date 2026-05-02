package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newWsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ws SLUG",
		Short: "Open a workspace in your editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			wsPath := paths.Workspace(args[0])
			if _, err := os.Stat(wsPath); os.IsNotExist(err) {
				return fmt.Errorf("workspace %q not found — run 'agx ps -a' to list workspaces", args[0])
			}

			editor := os.Getenv("AGX_EDITOR")
			if editor == "" {
				fmt.Fprintf(os.Stderr, "AGX_EDITOR is not set.\n")
				fmt.Fprintf(os.Stderr, "Set it by adding this line to your shell profile:\n\n")
				fmt.Fprintf(os.Stderr, "    echo 'export AGX_EDITOR=code' >> ~/.zshrc && source ~/.zshrc\n\n")
				fmt.Fprintf(os.Stderr, "Replace 'code' with your editor command (cursor, zed, subl, etc.).\n")
				return fmt.Errorf("AGX_EDITOR not set")
			}

			c := exec.Command(editor, wsPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}
