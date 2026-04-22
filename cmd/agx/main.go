package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	cmd, rest := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "bootstrap":
		err = cmdBootstrap(rest)
	case "run":
		err = cmdRun(rest)
	case "ls":
		err = cmdLs(rest)
	case "rm":
		err = cmdRm(rest)
	case "prune":
		err = cmdPrune(rest)
	case "-h", "--help", "help":
		usage(os.Stdout)
		return
	default:
		fmt.Fprintf(os.Stderr, "agx: unknown command %q\n", cmd)
		usage(os.Stderr)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "agx:", err)
		os.Exit(1)
	}
}

func usage(w *os.File) {
	fmt.Fprintln(w, `usage:
  agx bootstrap
  agx run [-f FILE] [--name NAME] [-i] [PROMPT]
  agx ls
  agx rm NAME [NAME...]
  agx prune`)
}


func cmdLs(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("ls takes no arguments")
	}
	paths, err := NewPaths()
	if err != nil {
		return err
	}
	names, err := ListWorkspaces(paths.Workspaces)
	if err != nil {
		return err
	}
	for _, n := range names {
		fmt.Println(n)
	}
	return nil
}

func cmdPrune(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("prune takes no arguments")
	}
	paths, err := NewPaths()
	if err != nil {
		return err
	}
	names, err := ListWorkspaces(paths.Workspaces)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Println("no workspaces to prune")
		return nil
	}
	for _, n := range names {
		fmt.Println(" ", n)
	}
	fmt.Printf("Remove %d workspace(s)? [y/N] ", len(names))
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if strings.TrimSpace(strings.ToLower(line)) != "y" {
		fmt.Println("aborted")
		return nil
	}
	for _, n := range names {
		if err := RemoveWorkspaces(paths.Workspaces, []string{n}); err != nil {
			return err
		}
		fmt.Println("Removed", n)
	}
	return nil
}

func cmdRm(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rm requires at least one workspace name")
	}
	paths, err := NewPaths()
	if err != nil {
		return err
	}
	return RemoveWorkspaces(paths.Workspaces, args)
}
