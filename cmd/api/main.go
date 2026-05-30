package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-mentorship-platform/backend/internal/platform/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	defer application.Close()

	if err := application.Run(ctx); err != nil {
		application.Logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}

	application.Logger.Info("application stopped")
}
