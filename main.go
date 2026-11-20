package main

import (
	"errors"
	"fmt"
	"os"
)

const version = "0.1.0"

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
		fmt.Println("bugsnap " + version)
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
  bugsnap inspect BUNDLE.zip
  bugsnap replay BUNDLE.zip [--apply] [--run] [--dir DIR]
  bugsnap version

capture flags:
  --log PATH    log file to include (last --tail lines)
  --tail N      log lines to keep (default 200)
  --out DIR     output directory (default ".")
  --note TEXT   note to attach to the bundle

Examples:
  bugsnap capture -- go test ./...
  bugsnap capture --log app.log --note "checkout fails" -- ./run.sh
  bugsnap inspect bugsnap-20260101T120000Z.zip
  bugsnap replay bugsnap-20260101T120000Z.zip --apply --run
`)
}
