package main

import (
	"reflect"
	"testing"
)

func TestBootstrapArgs(t *testing.T) {
	got := BootstrapArgs("/home/ehsan/.agx/home")
	want := []string{
		"run", "--rm", "-it",
		"-v", "/home/ehsan/.agx/home:/home/agx",
		"agx:latest",
		"/usr/local/bin/agx-bootstrap",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestRunArgs_Headless(t *testing.T) {
	got := RunArgs(RunConfig{
		HomeDir:       "/h",
		WorkspacePath: "/w/foo",
		Interactive:   false,
	})
	want := []string{
		"run", "--rm", "--init",
		"-v", "/h:/home/agx",
		"-v", "/w/foo:/workspace",
		"-e", "AGX_MODE=headless",
		"-e", "AGX_PROMPT_FILE=/workspace/.agx/prompt.txt",
		"agx:latest",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestRunArgs_Interactive(t *testing.T) {
	got := RunArgs(RunConfig{
		HomeDir:       "/h",
		WorkspacePath: "/w/foo",
		Interactive:   true,
	})
	want := []string{
		"run", "--rm", "--init", "-it",
		"-v", "/h:/home/agx",
		"-v", "/w/foo:/workspace",
		"-e", "AGX_MODE=interactive",
		"-e", "AGX_PROMPT_FILE=/workspace/.agx/prompt.txt",
		"agx:latest",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}
