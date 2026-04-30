package main

type RunConfig struct {
	HomeDir       string
	KbDir         string
	WorkspacePath string
	Interactive   bool
	Resume        bool
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
	args := []string{"run", "--rm", "--init"}
	switch {
	case c.Resume:
		args = append(args, "-it")
		mode = "resume"
	case c.Interactive:
		args = append(args, "-it")
		mode = "interactive"
	}
	args = append(args,
		"-v", c.HomeDir+":/home/agx",
		"-v", c.KbDir+":/kb",
		"-v", c.WorkspacePath+":/workspace",
		"-e", "AGX_MODE="+mode,
		"-e", "AGX_PROMPT_FILE=/workspace/.agx/prompt.txt",
		"agx:latest",
	)
	return args
}
