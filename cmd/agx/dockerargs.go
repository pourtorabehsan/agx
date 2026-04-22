package main

type RunConfig struct {
	HomeDir       string
	WorkspacePath string
	Interactive   bool
}

func BootstrapArgs(homeDir string) []string {
	return []string{
		"run", "--rm", "-it",
		"-v", homeDir + ":/home/agx",
		"agx:latest",
		"/usr/local/bin/agx-bootstrap",
	}
}

func RunArgs(c RunConfig) []string {
	mode := "headless"
	args := []string{"run", "--rm", "--init"}
	if c.Interactive {
		args = append(args, "-it")
		mode = "interactive"
	}
	args = append(args,
		"-v", c.HomeDir+":/home/agx",
		"-v", c.WorkspacePath+":/workspace",
		"-e", "AGX_MODE="+mode,
		"-e", "AGX_PROMPT_FILE=/workspace/.agx/prompt.txt",
		"agx:latest",
	)
	return args
}
