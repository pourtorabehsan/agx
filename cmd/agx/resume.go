package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func newResumeCmd() *cobra.Command {
	var agentName, model, memory, cpus string
	var detach, interactive bool

	cmd := &cobra.Command{
		Use:   "resume NAME",
		Short: "Re-run an existing workspace with its original prompt",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if detach && interactive {
				return errors.New("--detach and --interactive are mutually exclusive")
			}

			wsName := args[0]
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			wsPath := paths.Workspace(wsName)
			if _, err := os.Stat(wsPath); os.IsNotExist(err) {
				return fmt.Errorf("workspace %q not found — run 'agx ps -a' to list workspaces", wsName)
			}
			running, err := runningWorkspaces()
			if err != nil {
				return fmt.Errorf("docker ps: %w", err)
			}
			if running[wsName] {
				return fmt.Errorf("workspace %q is currently running — stop it first with 'agx kill %s'", wsName, wsName)
			}

			agent := DefaultAgent()
			meta, metaErr := ReadMeta(wsPath)
			if metaErr == nil {
				agent = meta.Agent
			}
			if cmd.Flags().Changed("agent") {
				agent, err = ParseAgent(agentName)
				if err != nil {
					return err
				}
			}

			// update started time in meta so ps -a reflects the new run
			if metaErr == nil {
				meta.Started = time.Now().UTC().Format(time.RFC3339)
				meta.Agent = agent
				if model != "" {
					meta.Model = model
				}
				_ = WriteMeta(wsPath, meta)
			}

			c := exec.Command("docker", RunArgs(RunConfig{
				HomeDir:       paths.HomeDir,
				KbDir:         paths.KbDir,
				WorkspacePath: wsPath,
				Agent:         agent,
				Interactive:   interactive,
				Model:         model,
				Memory:        memory,
				CPUs:          cpus,
			})...)

			logPath := filepath.Join(wsPath, ".agx", "session.log")
			logFile, err := os.Create(logPath)
			if err != nil {
				return err
			}
			defer logFile.Close()

			fmt.Fprintln(os.Stderr, "==> workspace:", wsPath)

			if detach {
				c.Stdout = logFile
				c.Stderr = logFile
				if err := c.Start(); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "==> detached: %s\n==> logs: %s\n", "agx-"+wsName, logPath)
				return nil
			}

			c.Stdin = os.Stdin
			if interactive {
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
			} else {
				c.Stdout = io.MultiWriter(os.Stdout, logFile)
				c.Stderr = io.MultiWriter(os.Stderr, logFile)
			}

			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "run in background; logs go to workspace session.log")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "attach a TTY")
	cmd.Flags().StringVar(&agentName, "agent", "", "agent override (claude or codex; defaults to workspace agent)")
	cmd.Flags().StringVar(&model, "model", "", "model override for the selected agent")
	cmd.Flags().StringVar(&memory, "memory", "", "container memory limit (e.g. 8g)")
	cmd.Flags().StringVar(&cpus, "cpus", "", "container CPU limit (e.g. 4)")
	return cmd
}
