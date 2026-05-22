package abs_test

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/RXWatcher/continuum-plugin-audiobooks/internal/abs"
)

// The track route must be registered so a real subscriber's enclosure
// URL resolves to a handler rather than chi's 404.
func TestPublicFeedTrackRouteIsRegistered(t *testing.T) {
	h := &abs.Handler{}
	r := chi.NewRouter()
	h.MountPublicFeed(r)

	rctx := chi.NewRouteContext()
	if !r.Match(rctx, http.MethodGet, "/feed/abc/track/0.mp3") {
		t.Fatal("GET /feed/{slug}/track/{idx} is not routed")
	}
}
