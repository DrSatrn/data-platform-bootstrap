// platform-api starts the HTTP control-plane server for orchestration,
// metadata, analytics, reporting, and health endpoints. This file stays small
// so operational startup remains easy to follow.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/streanor/data-platform/backend/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.RunAPI(ctx); err != nil {
		log.Fatalf("platform-api exited with error: %v", err)
	}
}
