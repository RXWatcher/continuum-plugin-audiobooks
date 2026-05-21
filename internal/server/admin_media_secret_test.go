package server_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
)

// TestAdminUpdateBackendConfig_RejectsShortMediaSecret pins the boundary
// validation on media_signing_secret: an admin must not be able to save a
// 1-character HMAC key. Empty string is allowed (operators may want to
// blank the secret during incident response).
func TestAdminUpdateBackendConfig_RejectsShortMediaSecret(t *testing.T) {
	h, st := liveServer(t)
	if _, err := st.EnsureBackendConfig(context.Background(), bytes.Repeat([]byte{1}, 32)); err != nil {
		t.Fatalf("seed cfg: %v", err)
	}

	for _, value := range []string{"x", "abcde", "fifteen-chars-1"} {
		body := `{"media_signing_secret":"` + value + `"}`
		w := do(h, jsonReq("PATCH", "/api/v1/admin/backend-config", asAdmin, body))
		if w.Code != http.StatusBadRequest {
			t.Errorf("len=%d: status = %d, want 400", len(value), w.Code)
		}
		if !strings.Contains(w.Body.String(), "at least 16") {
			t.Errorf("len=%d: body must surface the min-length message; got %q", len(value), w.Body.String())
		}
	}
}

// TestAdminUpdateBackendConfig_AcceptsValidMediaSecret confirms the
// happy-path threshold: a 16-character secret saves and the empty string
// clears it.
func TestAdminUpdateBackendConfig_AcceptsValidMediaSecret(t *testing.T) {
	h, st := liveServer(t)
	if _, err := st.EnsureBackendConfig(context.Background(), bytes.Repeat([]byte{1}, 32)); err != nil {
		t.Fatalf("seed cfg: %v", err)
	}

	long := `{"media_signing_secret":"sixteen-chars-ok"}`
	if w := do(h, jsonReq("PATCH", "/api/v1/admin/backend-config", asAdmin, long)); w.Code != http.StatusOK {
		t.Fatalf("16-char secret: status = %d body=%s", w.Code, w.Body.String())
	}

	clear := `{"media_signing_secret":""}`
	if w := do(h, jsonReq("PATCH", "/api/v1/admin/backend-config", asAdmin, clear)); w.Code != http.StatusOK {
		t.Fatalf("clear secret: status = %d body=%s", w.Code, w.Body.String())
	}
}
