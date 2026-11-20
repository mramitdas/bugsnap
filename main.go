package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

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
		fmt.Println("reprokit " + version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "reprokit:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`reprokit - capture a reproducible bug bundle

Usage:
  reprokit capture [flags] [-- COMMAND [ARGS...]]
  reprokit inspect BUNDLE.zip
  reprokit replay BUNDLE.zip [--apply] [--run] [--dir DIR]
  reprokit version

capture flags:
  --log PATH    log file to include (last --tail lines)
  --tail N      log lines to keep (default 200)
  --out DIR     output directory (default ".")
  --note TEXT   note to attach to the bundle

Examples:
  reprokit capture -- go test ./...
  reprokit capture --log app.log --note "checkout fails" -- ./run.sh
  reprokit inspect reprokit-20260101T120000Z.zip
  reprokit replay reprokit-20260101T120000Z.zip --apply --run
`)
}
