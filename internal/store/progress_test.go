package store_test

import (
	"testing"

	"github.com/RXWatcher/continuum-plugin-audiobooks/internal/store"
)

func TestUpsertProgressPersistsDuration(t *testing.T) {
	st, ctx := newStore(t)
	if err := st.UpsertProgress(ctx, store.Progress{
		UserID: "u1", BookID: "b1", CurrentSeconds: 30, DurationSeconds: 3600, ProgressPct: 0.0083,
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err := st.GetProgress(ctx, "u1", "b1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.DurationSeconds != 3600 {
		t.Errorf("DurationSeconds = %d, want 3600", got.DurationSeconds)
	}
}
