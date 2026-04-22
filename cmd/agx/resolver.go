package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type PromptSources struct {
	Arg         string
	File        string
	Stdin       io.Reader
	StdinIsTTY  bool
	Interactive bool
}

func ResolvePrompt(s PromptSources) (string, error) {
	if s.File != "" && s.Arg != "" {
		return "", errors.New("--file and positional prompt are mutually exclusive")
	}
	if s.File != "" {
		b, err := os.ReadFile(s.File)
		if err != nil {
			return "", fmt.Errorf("read prompt file: %w", err)
		}
		return string(b), nil
	}
	if s.Arg != "" {
		return s.Arg, nil
	}
	if s.Stdin != nil && !s.StdinIsTTY {
		b, err := io.ReadAll(s.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		if len(b) > 0 {
			return string(b), nil
		}
	}
	if s.Interactive {
		return "", nil
	}
	return "", errors.New("no prompt provided (use positional arg, -f FILE, stdin, or -i)")
}

func ResolveWorkspaceName(flagName, prompt string, now time.Time) (string, error) {
	if flagName != "" {
		s := Slug(flagName)
		if s == "" {
			return "", fmt.Errorf("--name %q produces an empty workspace name after sanitization", flagName)
		}
		return s, nil
	}
	if s := Slug(prompt); s != "" {
		return s, nil
	}
	return now.Format("2006-01-02-150405"), nil
}
