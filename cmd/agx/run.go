package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/term"
)

func cmdRun(rawArgs []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var (
		file        string
		name        string
		interactive bool
	)
	fs.StringVar(&file, "f", "", "read prompt from file (short)")
	fs.StringVar(&file, "file", "", "read prompt from file")
	fs.StringVar(&name, "name", "", "workspace name")
	fs.BoolVar(&interactive, "i", false, "attach a TTY (short)")
	fs.BoolVar(&interactive, "interactive", false, "attach a TTY")
	if err := fs.Parse(rawArgs); err != nil {
		return err
	}

	var arg string
	switch fs.NArg() {
	case 0:
		// nothing
	case 1:
		arg = fs.Arg(0)
	default:
		return fmt.Errorf("unexpected extra arguments: %v", fs.Args()[1:])
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

	wsName := ResolveWorkspaceName(name, prompt, time.Now())
	wsPath, err := EnsureWorkspace(paths.Workspaces, wsName)
	if err != nil {
		return err
	}
	promptPath := filepath.Join(wsPath, ".agx", "prompt.txt")
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return err
	}

	cmd := exec.Command("docker", RunArgs(RunConfig{
		HomeDir:       paths.HomeDir,
		WorkspacePath: wsPath,
		Interactive:   interactive,
	})...)
	cmd.Stdin = os.Stdin

	if interactive {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		logPath := filepath.Join(wsPath, ".agx", "session.log")
		logFile, err := os.Create(logPath)
		if err != nil {
			return err
		}
		defer logFile.Close()
		cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	}

	fmt.Fprintln(os.Stderr, "==> workspace:", wsPath)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}
