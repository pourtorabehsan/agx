package main

import "path/filepath"

type RunConfig struct {
	HomeDir       string
	KbDir         string
	WorkspacePath string
	Agent         Agent
	Interactive   bool
	Attach        bool
	Resume        bool
	Model         string
	Memory        string
	CPUs          string
}

func BootstrapArgs(homeDir, kbDir string) []string {
	return []string{
		"run", "--rm", "-it",
		"-v", homeDir + ":/home/agx",
		"-v", kbDir + ":/kb",
		"agx:latest",
		"/usr/local/bin/agx-bootstrap",
	}
}

func RunArgs(c RunConfig) []string {
	mode := "headless"
	agent := c.Agent
	if agent == "" {
		agent = DefaultAgent()
	}
	args := []string{"run", "--rm", "--init", "--name", "agx-" + filepath.Base(c.WorkspacePath)}
	switch {
	case c.Attach:
		args = append(args, "-it")
		mode = "attach"
	case c.Resume:
		args = append(args, "-it")
		mode = "resume"
	case c.Interactive:
		args = append(args, "-it")
		mode = "interactive"
	}
	if c.Memory != "" {
		args = append(args, "--memory", c.Memory)
	}
	if c.CPUs != "" {
		args = append(args, "--cpus", c.CPUs)
	}
	args = append(args,
		"-v", c.HomeDir+":/home/agx",
		"-v", c.KbDir+":/kb",
		"-v", c.WorkspacePath+":/workspace",
		"-e", "AGX_MODE="+mode,
		"-e", "AGX_AGENT="+agent.String(),
		"-e", "AGX_PROMPT_FILE=/workspace/.agx/prompt.txt",
	)
	if c.Model != "" {
		args = append(args, "-e", "AGX_MODEL="+c.Model)
	}
	args = append(args, "agx:latest")
	return args
}
