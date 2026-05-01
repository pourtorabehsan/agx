package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Meta struct {
	Prompt  string   `json:"prompt"`
	Repos   []string `json:"repos"`
	Started string   `json:"started"`
	Model   string   `json:"model,omitempty"`
}

func WriteMeta(wsPath string, m Meta) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(wsPath, ".agx", "meta.json"), data, 0o644)
}

func ReadMeta(wsPath string) (Meta, error) {
	data, err := os.ReadFile(filepath.Join(wsPath, ".agx", "meta.json"))
	if err != nil {
		return Meta{}, err
	}
	var m Meta
	if err := json.Unmarshal(data, &m); err != nil {
		return Meta{}, err
	}
	return m, nil
}
