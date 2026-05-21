package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/RXWatcher/continuum-plugin-audiobooks/internal/store"
)

// TestABSStandaloneOptIn_RoundTrip exercises the three /me/abs-standalone
// routes end-to-end through the chi router with a real Postgres store.
// Validates that:
//   - GET returns the current mode + per-user enabled flag
//   - POST inserts the opt-in row (HasStandaloneOptIn → true)
//   - DELETE removes the row
//   - All three routes 401 without an authenticated identity
func TestABSStandaloneOptIn_RoundTrip(t *testing.T) {
	h, st := liveServer(t)

	// Auth gate: every route must 401 without identity headers.
	for _, method := range []string{"GET", "POST", "DELETE"} {
		w := do(h, req(method, "/api/v1/me/abs-standalone", nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s without auth = %d, want 401", method, w.Code)
		}
	}

	// Set the admin mode to "opt_in" so the GET response reflects it.
	if _, err := st.EnsureBackendConfig(context.Background(), []byte("32-bytes-of-jwt-secret-for-test-")); err != nil {
		t.Fatalf("ensure cfg: %v", err)
	}
	cfg, _ := st.GetBackendConfig(context.Background())
	cfg.StandaloneLoginMode = store.StandaloneLoginModeOptIn
	if err := st.UpdateBackendConfig(context.Background(), cfg); err != nil {
		t.Fatalf("update cfg: %v", err)
	}

	// Initial GET — no row, enabled=false.
	w := do(h, req("GET", "/api/v1/me/abs-standalone", asUser))
	if w.Code != http.StatusOK {
		t.Fatalf("GET initial = %d body=%s", w.Code, w.Body)
	}
	var got struct {
		Mode    string `json:"mode"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Mode != "opt_in" || got.Enabled {
		t.Errorf("initial state = %+v, want {opt_in, false}", got)
	}

	// POST — opt in.
	if w := do(h, req("POST", "/api/v1/me/abs-standalone", asUser)); w.Code != http.StatusOK {
		t.Fatalf("POST = %d body=%s", w.Code, w.Body)
	}

	// GET — enabled=true.
	w = do(h, req("GET", "/api/v1/me/abs-standalone", asUser))
	got = struct {
		Mode    string `json:"mode"`
		Enabled bool   `json:"enabled"`
	}{}
	_ = json.NewDecoder(w.Body).Decode(&got)
	if !got.Enabled {
		t.Errorf("after POST: enabled = %v, want true", got.Enabled)
	}

	// DELETE — opt out.
	if w := do(h, req("DELETE", "/api/v1/me/abs-standalone", asUser)); w.Code != http.StatusOK {
		t.Fatalf("DELETE = %d body=%s", w.Code, w.Body)
	}

	// GET — enabled=false again.
	w = do(h, req("GET", "/api/v1/me/abs-standalone", asUser))
	got = struct {
		Mode    string `json:"mode"`
		Enabled bool   `json:"enabled"`
	}{}
	_ = json.NewDecoder(w.Body).Decode(&got)
	if got.Enabled {
		t.Errorf("after DELETE: enabled = %v, want false", got.Enabled)
	}
}

// TestABSStandaloneOptIn_PerUserIsolation guards against an opt-in row
// for user A accidentally enabling login for user B. The opt-in is a
// per-user permission and must not leak across accounts.
func TestABSStandaloneOptIn_PerUserIsolation(t *testing.T) {
	h, st := liveServer(t)
	if _, err := st.EnsureBackendConfig(context.Background(), []byte("32-bytes-of-jwt-secret-for-test-")); err != nil {
		t.Fatalf("ensure cfg: %v", err)
	}

	asAlice := map[string]string{"X-Continuum-User-Id": "alice", "X-Continuum-User-Role": "user"}
	asBob := map[string]string{"X-Continuum-User-Id": "bob", "X-Continuum-User-Role": "user"}

	if w := do(h, req("POST", "/api/v1/me/abs-standalone", asAlice)); w.Code != http.StatusOK {
		t.Fatalf("Alice POST = %d", w.Code)
	}

	// Bob's GET must still say enabled=false; Alice's opt-in is hers alone.
	w := do(h, req("GET", "/api/v1/me/abs-standalone", asBob))
	var got struct {
		Enabled bool `json:"enabled"`
	}
	_ = json.NewDecoder(w.Body).Decode(&got)
	if got.Enabled {
		t.Errorf("Bob saw Alice's opt-in: %+v", got)
	}
}
