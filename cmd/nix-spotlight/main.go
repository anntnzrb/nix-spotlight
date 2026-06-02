package main

import (
	"flag"
	"fmt"
	"os"

	nixspotlight "github.com/anntnzrb/nix-spotlight"
)

// version is set via ldflags at build time (e.g. -X main.version=0.2.0).
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 1 && args[0] == "--version" {
		fmt.Printf("nix-spotlight %s\n", version)
		return 0
	}

	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printUsage(os.Stdout)
		return 0
	}

	if len(args) < 1 || args[0] != "sync" {
		printUsage(os.Stderr)
		return 2
	}

	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		printUsage(os.Stderr)
	}
	noDock := fs.Bool("no-dock", false, "skip dock syncing")

	if err := fs.Parse(syncFlagArgs(args[1:])); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	positional := fs.Args()
	if len(positional) != 2 {
		fs.Usage()
		return 2
	}

	fromDir, toDir := positional[0], positional[1]
	trampolines, err := nixspotlight.SyncTrampolines(fromDir, toDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return 1
	}

	if !*noDock {
		result := nixspotlight.SyncDock(trampolines, "")
		for _, errMsg := range result.Errors {
			fmt.Fprintf(os.Stderr, "warning: %s\n", errMsg)
		}
	}

	fmt.Printf("Synced %d apps to %s\n", len(trampolines), toDir)
	return 0
}

func printUsage(out *os.File) {
	fmt.Fprintln(out, "usage: nix-spotlight sync <from> <to> [--no-dock]")
}

func syncFlagArgs(args []string) []string {
	normalized := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--no-dock" {
			normalized = append(normalized, arg)
		} else {
			positionals = append(positionals, arg)
		}
	}
	return append(normalized, positionals...)
}
