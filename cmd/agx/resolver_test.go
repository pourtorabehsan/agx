package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	cases := []struct {
		flag, generatedID, want string
	}{
		{flag: "my-ws", generatedID: "000102030405", want: "my-ws"},
		{flag: "My Workspace!", generatedID: "000102030405", want: "my-workspace"}, // flag is slugified
		{generatedID: "000102030405", want: "000102030405"},
	}
	for _, c := range cases {
		got, err := ResolveWorkspaceName(c.flag, c.generatedID)
		if err != nil {
			t.Errorf("ResolveWorkspaceName(%q,%q) unexpected error: %v", c.flag, c.generatedID, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveWorkspaceName(%q,%q) = %q, want %q", c.flag, c.generatedID, got, c.want)
		}
	}

	// All-special-char name should error.
	if _, err := ResolveWorkspaceName("!!!", "000102030405"); err == nil {
		t.Error("expected error for unsluggable --name")
	}
	if _, err := ResolveWorkspaceName("", ""); err == nil {
		t.Error("expected error for empty generated ID")
	}
}

func TestNewWorkspaceID(t *testing.T) {
	got, err := NewWorkspaceID(strings.NewReader("\x00\x01\x02\x03\x04\x05"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "000102030405"; got != want {
		t.Fatalf("NewWorkspaceID() = %q, want %q", got, want)
	}
}
