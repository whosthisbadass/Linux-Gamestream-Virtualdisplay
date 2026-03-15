package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/config"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/lifecycle"
)

const version = "0.2.0"

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "usage: %s <session-start|session-stop|monitor|status|doctor|validate-env|print-request|detect-client|show-config|show-rules|cleanup-stale|config-dump|version>\n", os.Args[0])
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
	case "status":
		err = controller.Status()
	case "doctor":
		err = controller.Doctor()
	case "validate-env":
		err = controller.ValidateEnv()
	case "print-request":
		err = controller.PrintRequest()
	case "detect-client":
		err = controller.DetectClient()
	case "show-config":
		err = controller.ShowConfig()
	case "show-rules":
		err = controller.ShowRules()
	case "cleanup-stale":
		err = controller.CleanupStale()
	case "config-dump":
		var cfg config.Config
		cfg, err = config.Load()
		if err == nil {
			var dump string
			dump, err = cfg.Dump()
			if err == nil {
				fmt.Println(dump)
			}
		}
	case "version":
		fmt.Println(version)
	default:
		flag.Usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
