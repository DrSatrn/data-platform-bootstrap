// These tests lock in the scheduler's timezone-aware cron behavior. The
// scheduler intentionally uses a simple minute-by-minute matching strategy, so
// these cases guard against regressions in local-time interpretation.
package scheduler

import (
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestCurrentScheduleSlotHonorsTimezone(t *testing.T) {
	now := time.Date(2026, time.March, 14, 2, 15, 0, 0, time.UTC)
	slot, supported, err := currentScheduleSlot(now, orchestration.Schedule{
		Cron:     "0 12 * * *",
		Timezone: "Australia/Brisbane",
	})
	if err != nil {
		t.Fatalf("currentScheduleSlot returned error: %v", err)
	}
	if !supported {
		t.Fatalf("expected schedule to be supported")
	}

	expected := time.Date(2026, time.March, 14, 2, 0, 0, 0, time.UTC)
	if !slot.Equal(expected) {
		t.Fatalf("expected slot %s, got %s", expected.Format(time.RFC3339), slot.Format(time.RFC3339))
	}
}

func TestCurrentScheduleSlotSupportsDayOfWeek(t *testing.T) {
	now := time.Date(2026, time.March, 15, 1, 0, 0, 0, time.UTC)
	slot, supported, err := currentScheduleSlot(now, orchestration.Schedule{
		Cron:     "0 9 * * 0",
		Timezone: "Australia/Brisbane",
	})
	if err != nil {
		t.Fatalf("currentScheduleSlot returned error: %v", err)
	}
	if !supported {
		t.Fatalf("expected schedule to be supported")
	}

	expected := time.Date(2026, time.March, 14, 23, 0, 0, 0, time.UTC)
	if !slot.Equal(expected) {
		t.Fatalf("expected slot %s, got %s", expected.Format(time.RFC3339), slot.Format(time.RFC3339))
	}
}
