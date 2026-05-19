package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func newPsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "List all workspaces",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := NewPaths()
			if err != nil {
				return err
			}
			running, err := runningWorkspaces()
			if err != nil {
				return fmt.Errorf("docker ps: %w", err)
			}
			names, err := ListWorkspaces(paths.Workspaces)
			if err != nil {
				return err
			}
			sort.Slice(names, func(i, j int) bool {
				ti := wsCreatedTime(paths.Workspace(names[i]))
				tj := wsCreatedTime(paths.Workspace(names[j]))
				return ti.After(tj)
			})

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCREATED\tSTATUS\tAGENT\tMODEL\tPROMPT")
			for _, name := range names {
				status := "stopped"
				if running[name] {
					status = "running"
				}
				wsPath := paths.Workspace(name)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					name,
					wsCreated(wsPath),
					status,
					wsAgent(wsPath),
					wsModel(wsPath),
					wsPrompt(wsPath),
				)
			}
			return w.Flush()
		},
	}
}

func runningWorkspaces() (map[string]bool, error) {
	out, err := exec.Command("docker", "ps", "--filter", "name=agx-", "--format", "{{.Names}}").Output()
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		// strip the "agx-" prefix that run.go adds
		m[strings.TrimPrefix(name, "agx-")] = true
	}
	return m, nil
}

func wsCreatedTime(wsPath string) time.Time {
	if meta, err := ReadMeta(wsPath); err == nil && meta.Started != "" {
		if t, err := time.Parse(time.RFC3339, meta.Started); err == nil {
			return t
		}
	}
	if info, err := os.Stat(wsPath); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}

func wsCreated(wsPath string) string {
	if t := wsCreatedTime(wsPath); !t.IsZero() {
		return t.Local().Format("Jan 02 15:04")
	}
	return "unknown"
}

func wsPrompt(wsPath string) string {
	if meta, err := ReadMeta(wsPath); err == nil && meta.Prompt != "" {
		return truncate(meta.Prompt, 40)
	}
	data, err := os.ReadFile(filepath.Join(wsPath, ".agx", "prompt.txt"))
	if err != nil {
		return ""
	}
	s := strings.TrimPrefix(string(data), "invoke conduct skill\n")
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return truncate(line, 40)
		}
	}
	return ""
}

func wsAgent(wsPath string) string {
	if meta, err := ReadMeta(wsPath); err == nil {
		return meta.Agent.String()
	}
	return DefaultAgent().String()
}

func wsModel(wsPath string) string {
	if meta, err := ReadMeta(wsPath); err == nil && meta.Model != "" {
		return meta.Model
	}
	return "default"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
