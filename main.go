package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"

	"github.com/charmbracelet/log"
	"github.com/ricardo/k8s-managed-field-cleanup/internal/cleanup"
)

func main() {
	var (
		dryRun   bool
		logLevel string
		k8sCtx   string
	)
	flag.BoolVar(&dryRun, "dry-run", false, "just list things to be cleaned")
	flag.StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error). Default info")
	flag.StringVar(&k8sCtx, "context", "", "kubernetes local context override. Default is current context")
	flag.Parse()

	log.SetLevel(log.ParseLevel(logLevel))
	errorLogger := log.New(os.Stderr)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() { signal.Stop(c) }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-c: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-c // second signal, hard exit
		errorLogger.Error("second interrupt, exiting")
		os.Exit(1)
	}()

	if err := cleanup.DoRunCleanup(ctx, k8sCtx, dryRun); err != nil {
		if !errors.Is(err, context.Canceled) {
			errorLogger.Error(err.Error())
		}
	}
}
