package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ContinuumApp/continuum-plugin-audiobooks/internal/auth"
	"github.com/ContinuumApp/continuum-plugin-audiobooks/internal/hlc"
	"github.com/ContinuumApp/continuum-plugin-audiobooks/internal/store"
)

// Replica sync for audiobook bookmarks. Mirrors the ebook plugin's
// annotation sync — clients pull changes since a cursor + push
// batches of local changes. Row-level LWW with tombstones; later
// refinement to per-field LWW lives in the same change-log shape.

func (s *Server) mountSyncRoutes(r chi.Router) {
	r.Get("/me/sync/bookmarks", s.handlePullBookmarkChanges)
	r.Post("/me/sync/bookmarks", s.handlePushBookmarkChanges)
}

func (s *Server) handlePullBookmarkChanges(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.RequireUser(w, r)
	if !ok {
		return
	}
	since := r.URL.Query().Get("since")
	limit := 500
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	changes, err := s.d.Store.PullBookmarkChanges(r.Context(), id.UserID, since, limit)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	next := since
	if len(changes) > 0 {
		next = changes[len(changes)-1].HLC
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"changes":     changes,
		"next_cursor": next,
	})
}

type pushBookmarkBody struct {
	BookmarkID string          `json:"bookmark_id"`
	Op         string          `json:"op"`
	Payload    json.RawMessage `json:"payload"`
	HLC        string          `json:"hlc"`
}

func (s *Server) handlePushBookmarkChanges(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.RequireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Changes []pushBookmarkBody `json:"changes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	applied := 0
	var lastHLC string
	for _, ch := range body.Changes {
		if ch.BookmarkID == "" || ch.Op == "" {
			continue
		}
		var stamp hlc.Timestamp
		if ch.HLC != "" {
			parsed, err := hlc.Parse(ch.HLC)
			if err != nil {
				continue
			}
			stamp = parsed
			s.syncClock().Observe(stamp)
		} else {
			stamp = s.syncClock().Now()
		}
		if err := s.d.Store.AppendBookmarkChange(r.Context(), store.BookmarkChange{
			HLC:        stamp.String(),
			UserID:     id.UserID,
			BookmarkID: ch.BookmarkID,
			Op:         ch.Op,
			Payload:    ch.Payload,
			OriginNode: stamp.NodeID,
		}); err != nil {
			writeInternal(w, r, err)
			return
		}
		applied++
		lastHLC = stamp.String()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"applied": applied,
		"cursor":  lastHLC,
	})
}

// syncClock returns the process-shared HLC clock. Lazily
// constructed; one clock per replica with a stable nodeID.
func (s *Server) syncClock() *hlc.Clock {
	s.syncClockOnce.Do(func() {
		s.clockCached = hlc.New("audiobook-replica")
	})
	return s.clockCached
}

// _ = chi placeholder so the import stays if the route surface
// grows path params later.
var _ = chi.URLParam
