// These tests cover the local-first audit store so privileged actions remain
// traceable across restarts.
package audit

import "testing"

func TestMemoryStoreAppendAndListRecent(t *testing.T) {
	store := NewMemoryStore()
	if err := store.Append(Event{ActorSubject: "alice", ActorRole: "editor", Action: "trigger_pipeline", Resource: "personal_finance_pipeline", Outcome: "success"}); err != nil {
		t.Fatalf("append first event: %v", err)
	}
	if err := store.Append(Event{ActorSubject: "alice", ActorRole: "editor", Action: "save_dashboard", Resource: "finance_overview", Outcome: "success"}); err != nil {
		t.Fatalf("append second event: %v", err)
	}

	events, err := store.ListRecent(1)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}
	if len(events) != 1 || events[0].Action != "save_dashboard" {
		t.Fatalf("unexpected recent events: %#v", events)
	}
}
