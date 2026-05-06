package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newAttachCmd() *cobra.Command {
	var agentName string

	cmd := &cobra.Command{
		Use:   "attach NAME",
		Short: "Open an interactive agent session in an existing workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			wsPath := filepath.Join(paths.Workspaces, args[0])
			if _, err := os.Stat(wsPath); os.IsNotExist(err) {
				return fmt.Errorf("workspace %q not found — run 'agx ps -a' to list workspaces", args[0])
			}
			agent := DefaultAgent()
			if meta, err := ReadMeta(wsPath); err == nil {
				agent = meta.Agent
			}
			if cmd.Flags().Changed("agent") {
				agent, err = ParseAgent(agentName)
				if err != nil {
					return err
				}
			}
			c := exec.Command("docker", RunArgs(RunConfig{
				HomeDir:       paths.HomeDir,
				KbDir:         paths.KbDir,
				WorkspacePath: wsPath,
				Agent:         agent,
				Attach:        true,
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
	cmd.Flags().StringVar(&agentName, "agent", "", "agent override (claude or codex; defaults to workspace agent)")
	return cmd
}
