package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolvePrompt_Arg(t *testing.T) {
	got, err := ResolvePrompt(PromptSources{Arg: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestResolvePrompt_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.md")
	if err := os.WriteFile(path, []byte("from file"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ResolvePrompt(PromptSources{File: path})
	if err != nil {
		t.Fatal(err)
	}
	if got != "from file" {
		t.Errorf("got %q", got)
	}
}

func TestResolvePrompt_Stdin(t *testing.T) {
	got, err := ResolvePrompt(PromptSources{
		Stdin:      strings.NewReader("via stdin"),
		StdinIsTTY: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "via stdin" {
		t.Errorf("got %q", got)
	}
}

func TestResolvePrompt_StdinIgnoredWhenTTY(t *testing.T) {
	_, err := ResolvePrompt(PromptSources{
		Stdin:      strings.NewReader("ignored"),
		StdinIsTTY: true,
	})
	if err == nil {
		t.Fatal("expected error when stdin is a TTY and no other source given")
	}
}

func TestResolvePrompt_ArgAndFileConflict(t *testing.T) {
	_, err := ResolvePrompt(PromptSources{Arg: "x", File: "/nope"})
	if err == nil {
		t.Fatal("expected mutual-exclusion error")
	}
}

func TestResolvePrompt_InteractiveAllowsEmpty(t *testing.T) {
	got, err := ResolvePrompt(PromptSources{Interactive: true, StdinIsTTY: true})
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestResolvePrompt_NoSourceNonInteractive(t *testing.T) {
	_, err := ResolvePrompt(PromptSources{StdinIsTTY: true})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveWorkspaceName(t *testing.T) {
	fixed := time.Date(2026, 4, 21, 14, 30, 22, 0, time.UTC)
	cases := []struct {
		flag, prompt, want string
	}{
		{"my-ws", "anything", "my-ws"},
		{"My Workspace!", "anything", "my-workspace"}, // flag is slugified
		{"", "Refactor AUTH!!", "refactor-auth-2026-04-21-143022-000000000"},
		{"", "", "2026-04-21-143022-000000000"},
		{"", "!!!", "2026-04-21-143022-000000000"}, // slug empties out
	}
	for _, c := range cases {
		got, err := ResolveWorkspaceName(c.flag, c.prompt, fixed)
		if err != nil {
			t.Errorf("ResolveWorkspaceName(%q,%q) unexpected error: %v", c.flag, c.prompt, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveWorkspaceName(%q,%q) = %q, want %q", c.flag, c.prompt, got, c.want)
		}
	}

	// All-special-char name should error.
	if _, err := ResolveWorkspaceName("!!!", "anything", fixed); err == nil {
		t.Error("expected error for unsluggable --name")
	}
}

func TestWorkspaceTimestampIncludesNanoseconds(t *testing.T) {
	fixed := time.Date(2026, 4, 21, 14, 30, 22, 123456789, time.UTC)
	if got, want := WorkspaceTimestamp(fixed), "2026-04-21-143022-123456789"; got != want {
		t.Fatalf("WorkspaceTimestamp() = %q, want %q", got, want)
	}
}
