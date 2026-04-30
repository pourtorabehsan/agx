package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newRunCmd() *cobra.Command {
	var file, name, repo string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "run [PROMPT]",
		Short: "Run a Claude Code session in a container",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var arg string
			if len(args) == 1 {
				arg = args[0]
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
			if prompt != "" {
				prompt = "invoke conduct skill\n" + prompt
			}
			if repo != "" {
				prompt = prompt + "\nTarget repo: " + repo
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
			promptPath := filepath.Join(wsPath, ".agx", "prompt.txt")
			if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
				return err
			}

			c := exec.Command("docker", RunArgs(RunConfig{
				HomeDir:       paths.HomeDir,
				KbDir:         paths.KbDir,
				WorkspacePath: wsPath,
				Interactive:   interactive,
			})...)
			c.Stdin = os.Stdin

			if interactive {
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
			} else {
				logPath := filepath.Join(wsPath, ".agx", "session.log")
				logFile, err := os.Create(logPath)
				if err != nil {
					return err
				}
				defer logFile.Close()
				c.Stdout = io.MultiWriter(os.Stdout, logFile)
				c.Stderr = io.MultiWriter(os.Stderr, logFile)
			}

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

	cmd.Flags().StringVarP(&file, "file", "f", "", "read prompt from file")
	cmd.Flags().StringVarP(&name, "name", "n", "", "workspace name")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "target repository (owner/name), appended to prompt")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "attach a TTY")
	return cmd
}
