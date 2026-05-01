package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newRunCmd() *cobra.Command {
	var file, name, model, memory, cpus string
	var repos []string
	var interactive, detach, noConductConduct bool

	cmd := &cobra.Command{
		Use:   "run [PROMPT]",
		Short: "Run a Claude Code session in a container",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var arg string
			if len(args) == 1 {
				arg = args[0]
			}

			if detach && interactive {
				return errors.New("--detach and --interactive are mutually exclusive")
			}

			stdinIsTTY := term.IsTerminal(int(os.Stdin.Fd()))
			prompt, err := ResolvePrompt(PromptSources{
				Arg:         arg,
				File:        file,
				Stdin:       os.Stdin,
				StdinIsTTY:  stdinIsTTY,
				Interactive: interactive,
			})
			if err != nil {
				return err
			}
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			if _, err := os.Stat(paths.HomeDir); os.IsNotExist(err) {
				return fmt.Errorf("%s missing — run 'agx bootstrap' first", paths.HomeDir)
			}

			wsName, err := ResolveWorkspaceName(name, prompt, time.Now())
			if err != nil {
				return err
			}
			wsPath, err := EnsureWorkspace(paths.Workspaces, wsName)
			if err != nil {
				return err
			}

			if err := WriteMeta(wsPath, Meta{
				Prompt:  firstLine(prompt, 80),
				Repos:   repos,
				Started: time.Now().UTC().Format(time.RFC3339),
				Model:   model,
			}); err != nil {
				return err
			}

			if prompt != "" && !noConductConduct {
				prompt = "invoke conduct skill\n" + prompt
			}
			for _, r := range repos {
				prompt += "\nTarget repo: " + r
			}

			promptPath := filepath.Join(wsPath, ".agx", "prompt.txt")
			if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
				return err
			}

			c := exec.Command("docker", RunArgs(RunConfig{
				HomeDir:       paths.HomeDir,
				KbDir:         paths.KbDir,
				WorkspacePath: wsPath,
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

	cmd.Flags().StringVarP(&file, "file", "f", "", "read prompt from file")
	cmd.Flags().StringVarP(&name, "name", "n", "", "workspace name")
	cmd.Flags().StringArrayVarP(&repos, "repo", "r", nil, "target repository (owner/name); repeatable")
	cmd.Flags().StringVar(&model, "model", "", "model override (e.g. claude-opus-4-7)")
	cmd.Flags().StringVar(&memory, "memory", "", "container memory limit (e.g. 8g)")
	cmd.Flags().StringVar(&cpus, "cpus", "", "container CPU limit (e.g. 4)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "attach a TTY")
	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "run in background; logs go to workspace session.log")
	cmd.Flags().BoolVar(&noConductConduct, "no-conduct", false, "skip the 'invoke conduct skill' prefix (use for simple/exploratory prompts)")
	return cmd
}

func firstLine(s string, max int) string {
	line := strings.SplitN(s, "\n", 2)[0]
	line = strings.TrimSpace(line)
	if len(line) > max {
		line = line[:max]
	}
	return line
}
