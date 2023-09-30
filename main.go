package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/charmbracelet/log"
	"github.com/golonzovsky/k8s-clean-managed-fields/internal/cleanup"
)

func main() {
	log.SetLevel(log.DebugLevel)
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

	if err := cleanup.DoRunCleanup(ctx, true); err != nil {
		if !errors.Is(err, context.Canceled) {
			errorLogger.Error(err.Error())
		}
	}
}
