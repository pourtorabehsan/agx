package main

import (
	"fmt"
	"os"
	"os/exec"
)

func cmdBootstrap(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("bootstrap takes no arguments")
	}
	paths, err := NewPaths()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(paths.HomeDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(paths.Workspaces, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("docker", BootstrapArgs(paths.HomeDir)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 125 {
			return fmt.Errorf("docker failed — is the image built? try 'make build'")
		}
		return err
	}
	return nil
}
