package main

import (
	"fmt"
	"os"
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
  agx rm NAME [NAME...]`)
}

func cmdBootstrap(args []string) error { return fmt.Errorf("not implemented") }
func cmdRun(args []string) error       { return fmt.Errorf("not implemented") }
func cmdLs(args []string) error        { return fmt.Errorf("not implemented") }
func cmdRm(args []string) error        { return fmt.Errorf("not implemented") }
