// platform-worker starts the job execution loop responsible for running
// ingestion, transform, and quality tasks. Keeping worker startup explicit
// helps make failure and shutdown behavior easier to reason about.
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

	if err := app.RunWorker(ctx); err != nil {
		log.Fatalf("platform-worker exited with error: %v", err)
	}
}
