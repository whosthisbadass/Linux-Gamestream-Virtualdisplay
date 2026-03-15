package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/lifecycle"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s <session-start|session-stop|monitor>\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	controller := lifecycle.NewController()
	command := flag.Arg(0)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var err error
	switch command {
	case "session-start":
		err = controller.SessionStart(ctx)
	case "session-stop":
		err = controller.SessionStop()
	case "monitor":
		err = controller.Monitor(ctx)
	default:
		flag.Usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
