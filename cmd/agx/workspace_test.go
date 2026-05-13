package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestEnsureWorkspace(t *testing.T) {
	dir := t.TempDir()
	p, err := EnsureWorkspace(dir, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, "foo") {
		t.Errorf("wrong path: %q", p)
	}
	if _, err := os.Stat(filepath.Join(p, ".agx")); err != nil {
		t.Errorf(".agx subdir missing: %v", err)
	}
	// Second call on the same name is a no-op, not an error.
	if _, err := EnsureWorkspace(dir, "foo"); err != nil {
		t.Errorf("re-ensure: %v", err)
	}
}

func TestEnsureNewWorkspace(t *testing.T) {
	dir := t.TempDir()
	p, err := EnsureNewWorkspace(dir, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, "foo") {
		t.Errorf("wrong path: %q", p)
	}
	if _, err := os.Stat(filepath.Join(p, ".agx")); err != nil {
		t.Errorf(".agx subdir missing: %v", err)
	}
	if _, err := EnsureNewWorkspace(dir, "foo"); !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected os.ErrExist, got %v", err)
	}
}

func TestListWorkspaces_Missing(t *testing.T) {
	names, err := ListWorkspaces(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty, got %v", names)
	}
}

func TestListWorkspaces_Sorted(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"bravo", "alpha", "charlie"} {
		if err := os.Mkdir(filepath.Join(dir, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Dangling regular file — should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "not-a-ws.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ListWorkspaces(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"alpha", "bravo", "charlie"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRemoveWorkspaces_OK(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"a", "b"} {
		if err := os.MkdirAll(filepath.Join(dir, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := RemoveWorkspaces(dir, []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("dir not empty: %v", entries)
	}
}

func TestRemoveWorkspaces_Missing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "exists"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := RemoveWorkspaces(dir, []string{"exists", "ghost"})
	if err == nil {
		t.Fatal("expected error for missing workspace")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error should mention ghost: %v", err)
	}
	// "exists" should still be removed even though "ghost" failed.
	if _, err := os.Stat(filepath.Join(dir, "exists")); !os.IsNotExist(err) {
		t.Errorf("exists should have been removed")
	}
}
