package main

import (
	"bufio"
	"fmt"
	"os"
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
		newLsCmd(),
		newRmCmd(),
		newPruneCmd(),
	)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "agx:", err)
		os.Exit(1)
	}
}

func newLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List workspaces",
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
			for _, n := range names {
				fmt.Println(n)
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
