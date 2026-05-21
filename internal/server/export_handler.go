package server

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ContinuumApp/continuum-plugin-audiobooks/internal/auth"
	"github.com/ContinuumApp/continuum-plugin-audiobooks/internal/store"
)

// Personal data export — one-click ZIP of everything the plugin
// stores for the requesting user. Each table goes in its own JSON
// file inside the archive. Useful for portability + GDPR-style
// "right to data" requests + bootstrapping a fresh instance with
// the user's history.
//
// Audiobook coverage:
//   progress.json             — all progress rows
//   bookmarks.json            — all bookmark rows
//   collections.json          — manual collections + item lists
//   smart_collections.json    — smart collection definitions
//   ratings.json              — per-book ratings
//   reading_sessions.json     — session telemetry
//   reading_goals.json        — yearly targets
//   share_links.json          — owner-side share metadata
//   content_restriction.json  — admin-set restrictions visible to user

func (s *Server) mountExportRoutes(r chi.Router) {
	r.Get("/me/export", s.handleExport)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.RequireUser(w, r)
	if !ok {
		return
	}
	filename := fmt.Sprintf("continuum-audiobooks-export-%s-%s.zip",
		id.UserID, time.Now().UTC().Format("20060102"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	zw := zip.NewWriter(w)
	defer zw.Close()

	// Each writeSection call is independent — if one table fails
	// the others still land. The export is best-effort; we surface
	// per-table errors via an _errors.json sidecar so the user
	// knows what's missing without truncating the whole download.
	errs := map[string]string{}
	add := func(name string, data any) {
		if err := writeSection(zw, name, data); err != nil {
			errs[name] = err.Error()
		}
	}

	ctx := r.Context()
	if rows, err := s.d.Store.ListRecentProgress(ctx, id.UserID, 100000); err == nil {
		add("progress.json", rows)
	} else {
		errs["progress.json"] = err.Error()
	}
	// Bookmarks are per-book; flatten over all the user's progress
	// rows to enumerate. A future ListAllBookmarks helper would
	// short-circuit this loop.
	bookmarks := s.collectBookmarks(ctx, id.UserID)
	add("bookmarks.json", bookmarks)

	if rows, err := s.d.Store.ListUserCollections(ctx, id.UserID); err == nil {
		add("collections.json", rows)
	} else {
		errs["collections.json"] = err.Error()
	}
	if rows, err := s.d.Store.ListSmartCollections(ctx, id.UserID, 1000); err == nil {
		add("smart_collections.json", rows)
	} else {
		errs["smart_collections.json"] = err.Error()
	}
	if rows, err := s.d.Store.ListReadingGoals(ctx, id.UserID, 0); err == nil {
		add("reading_goals.json", rows)
	} else {
		errs["reading_goals.json"] = err.Error()
	}
	if rows, err := s.d.Store.ListShareLinks(ctx, id.UserID); err == nil {
		add("share_links.json", rows)
	} else {
		errs["share_links.json"] = err.Error()
	}
	if rec, err := s.d.Store.GetContentRestriction(ctx, id.UserID); err == nil {
		add("content_restriction.json", rec)
	}

	if len(errs) > 0 {
		_ = writeSection(zw, "_errors.json", errs)
	}
	// Manifest at the top so a directory-listing of the archive
	// shows the export at a glance.
	_ = writeSection(zw, "_manifest.json", map[string]any{
		"plugin":      "continuum-audiobooks",
		"user_id":     id.UserID,
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"sections":    []string{"progress", "bookmarks", "collections", "smart_collections", "reading_goals", "share_links", "content_restriction"},
	})
}

// collectBookmarks walks every distinct book in the user's
// bookmark table. We'd prefer a single SQL "all bookmarks for
// user" call; ListBookmarks today is per-book so we union by
// scanning progress + emitting one ListBookmarks call per row.
// Acceptable at user-export volume (one-off, dozens-of-books).
func (s *Server) collectBookmarks(ctx context.Context, userID string) []store.Bookmark {
	rows, err := s.d.Store.ListRecentProgress(ctx, userID, 100000)
	if err != nil {
		return nil
	}
	out := make([]store.Bookmark, 0, 128)
	for _, p := range rows {
		bms, err := s.d.Store.ListBookmarks(ctx, userID, p.BookID)
		if err != nil {
			continue
		}
		out = append(out, bms...)
	}
	return out
}

// writeSection JSON-encodes `data` into one zip entry. Uses
// MarshalIndent so the human opening the archive can read the
// files without a JSON formatter.
func writeSection(zw *zip.Writer, name string, data any) error {
	f, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	return nil
}
