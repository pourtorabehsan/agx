package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "agx",
		Short:         "Run sandboxed Claude Code sessions in Docker",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newBootstrapCmd(),
		newRunCmd(),
		newResumeCmd(),
		newAttachCmd(),
		newLogsCmd(),
		newPsCmd(),
		newKillCmd(),
		newWaitCmd(),
		newPhasesCmd(),
		newRmCmd(),
		newPruneCmd(),
		newUpgradeCmd(),
		newWsCmd(),
		newCdCmd(),
	)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "agx:", err)
		os.Exit(1)
	}
}

func newPhasesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "phases NAME",
		Short: "List available phase outputs for a workspace",
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
			phasesDir := filepath.Join(wsPath, ".agx", "phases")
			entries, err := os.ReadDir(phasesDir)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fmt.Println("no phases yet")
					return nil
				}
				return err
			}
			for _, e := range entries {
				if phase, ok := strings.CutSuffix(e.Name(), "-output.md"); ok {
					fmt.Println(phase)
				}
			}
			return nil
		},
	}
}

func newRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm NAME [NAME...]",
		Short: "Remove one or more workspaces",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			return RemoveWorkspaces(paths.Workspaces, args)
		},
	}
}

func newPruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Remove all workspaces (with confirmation)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			names, err := ListWorkspaces(paths.Workspaces)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				fmt.Println("no workspaces to prune")
				return nil
			}
			for _, n := range names {
				fmt.Println(" ", n)
			}
			fmt.Printf("Remove %d workspace(s)? [y/N] ", len(names))
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) != "y" {
				fmt.Println("aborted")
				return nil
			}
			for _, n := range names {
				if err := RemoveWorkspaces(paths.Workspaces, []string{n}); err != nil {
					return err
				}
				fmt.Println("Removed", n)
			}
			return nil
		},
	}
}
