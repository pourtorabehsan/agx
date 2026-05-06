package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMetaDefaultsAgent(t *testing.T) {
	wsPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wsPath, ".agx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsPath, ".agx", "meta.json"), []byte(`{"prompt":"p","repos":null,"started":"s"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadMeta(wsPath)
	if err != nil {
		t.Fatal(err)
	}
	if got.Agent != AgentCodex {
		t.Fatalf("got %q, want %q", got.Agent, AgentCodex)
	}
}

func TestWriteMetaWritesAgent(t *testing.T) {
	wsPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wsPath, ".agx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteMeta(wsPath, Meta{
		Prompt:  "p",
		Started: "s",
		Agent:   AgentCodex,
	}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(wsPath, ".agx", "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"prompt":"p","repos":null,"started":"s","agent":"codex"}` {
		t.Fatalf("got %s", data)
	}
}

func TestWriteMetaDefaultsAgent(t *testing.T) {
	wsPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wsPath, ".agx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteMeta(wsPath, Meta{
		Prompt:  "p",
		Started: "s",
	}); err != nil {
		t.Fatal(err)
	}

	got, err := ReadMeta(wsPath)
	if err != nil {
		t.Fatal(err)
	}
	if got.Agent != AgentCodex {
		t.Fatalf("got %q, want %q", got.Agent, AgentCodex)
	}
}
