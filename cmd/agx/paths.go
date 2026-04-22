package main

import (
	"os"
	"path/filepath"
)

type Paths struct {
	AgxDir     string
	HomeDir    string
	Workspaces string
}

func NewPaths() (Paths, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}
	agx := filepath.Join(h, ".agx")
	return Paths{
		AgxDir:     agx,
		HomeDir:    filepath.Join(agx, "home"),
		Workspaces: filepath.Join(agx, "workspaces"),
	}, nil
}

func (p Paths) Workspace(name string) string {
	return filepath.Join(p.Workspaces, name)
}
