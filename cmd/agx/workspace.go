package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func EnsureWorkspace(workspacesDir, name string) (string, error) {
	p := filepath.Join(workspacesDir, name)
	if err := os.MkdirAll(filepath.Join(p, ".agx"), 0o755); err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	return p, nil
}

func EnsureNewWorkspace(workspacesDir, name string) (string, error) {
	if err := os.MkdirAll(workspacesDir, 0o755); err != nil {
		return "", fmt.Errorf("create workspaces dir: %w", err)
	}
	p := filepath.Join(workspacesDir, name)
	if err := os.Mkdir(p, 0o755); err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	if err := os.Mkdir(filepath.Join(p, ".agx"), 0o755); err != nil {
		_ = os.RemoveAll(p)
		return "", fmt.Errorf("create workspace metadata dir: %w", err)
	}
	return p, nil
}

func ListWorkspaces(workspacesDir string) ([]string, error) {
	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func RemoveWorkspaces(workspacesDir string, names []string) error {
	var missing []string
	for _, n := range names {
		p := filepath.Join(workspacesDir, n)
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			missing = append(missing, n)
			continue
		}
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("remove %s: %w", n, err)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("workspace(s) not found: %v", missing)
	}
	return nil
}
