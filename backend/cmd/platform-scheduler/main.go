// platform-scheduler starts the scheduling loop that evaluates time-based
// releases and dependency readiness. The scheduler is isolated from the API so
// long-running control tasks cannot starve interactive traffic.
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

	if err := app.RunScheduler(ctx); err != nil {
		log.Fatalf("platform-scheduler exited with error: %v", err)
	}
}
