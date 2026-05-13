package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWsAgentAndModelFromMeta(t *testing.T) {
	wsPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wsPath, ".agx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteMeta(wsPath, Meta{
		Prompt:  "p",
		Started: "s",
		Agent:   AgentClaude,
		Model:   "claude-opus-4-5",
	}); err != nil {
		t.Fatal(err)
	}

	if got := wsAgent(wsPath); got != "claude" {
		t.Fatalf("wsAgent() = %q, want %q", got, "claude")
	}
	if got := wsModel(wsPath); got != "claude-opus-4-5" {
		t.Fatalf("wsModel() = %q, want %q", got, "claude-opus-4-5")
	}
}

func TestWsAgentAndModelDefaults(t *testing.T) {
	wsPath := t.TempDir()

	if got := wsAgent(wsPath); got != "codex" {
		t.Fatalf("wsAgent() = %q, want %q", got, "codex")
	}
	if got := wsModel(wsPath); got != "default" {
		t.Fatalf("wsModel() = %q, want %q", got, "default")
	}
}
