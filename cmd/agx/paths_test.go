package main

import (
	"path/filepath"
	"testing"
)

func TestNewPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p, err := NewPaths()
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}

	want := Paths{
		AgxDir:     filepath.Join(home, ".agx"),
		HomeDir:    filepath.Join(home, ".agx", "home"),
		Workspaces: filepath.Join(home, ".agx", "workspaces"),
		KbDir:      filepath.Join(home, ".agx", "kb"),
	}
	if p != want {
		t.Errorf("got %+v, want %+v", p, want)
	}
}

func TestPaths_Workspace(t *testing.T) {
	p := Paths{Workspaces: "/tmp/ws"}
	if got := p.Workspace("foo"); got != "/tmp/ws/foo" {
		t.Errorf("Workspace(foo) = %q", got)
	}
}
