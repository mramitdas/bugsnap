package main

import (
	"errors"
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type exitError struct{ code int }

func (e exitError) Error() string { return fmt.Sprintf("command exited with status %d", e.code) }

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "capture":
		err = cmdCapture(os.Args[2:])
	case "inspect":
		err = cmdInspect(os.Args[2:])
	case "replay":
		err = cmdReplay(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("bugsnap %s\ncommit: %s\nbuilt:  %s\n", version, commit, date)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		var ee exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		fmt.Fprintln(os.Stderr, "bugsnap:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`bugsnap - capture a reproducible bug bundle

Usage:
  bugsnap capture [flags] [-- COMMAND [ARGS...]]
  bugsnap inspect BUNDLE.zip [--json]
  bugsnap replay BUNDLE.zip [--apply] [--run] [--docker] [--json] [--dir DIR]
  bugsnap version

capture flags:
  --log PATH    log file to include (last --tail lines)
  --tail N      log lines to keep (default 200)
  --out DIR     output directory (default ".")
  --note TEXT   note to attach to the bundle
  --json        print machine-readable JSON

Examples:
  bugsnap capture -- go test ./...
  bugsnap capture --log app.log --note "checkout fails" -- ./run.sh
  bugsnap inspect bugsnap-20260101T120000Z.zip
  bugsnap replay bugsnap-20260101T120000Z.zip --apply --run
  bugsnap replay bugsnap-20260101T120000Z.zip --docker --image node:22
`)
}
