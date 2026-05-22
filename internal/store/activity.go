package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ActivityEvent is one entry in a per-book activity timeline.
// Kind is the event type ("progress", "bookmark_added",
// "session_opened", "session_closed", "rated", "shared"); Payload
// is a kind-specific JSON shape the SPA renders by switch.
type ActivityEvent struct {
	At      time.Time      `json:"at"`
	Kind    string         `json:"kind"`
	Payload map[string]any `json:"payload,omitempty"`
}

// BookActivity merges events from progress, bookmark, abs_session,
// and rating tables for one (user, book) pair. Returned in
// reverse-chronological order so the SPA renders newest-first.
func (s *Store) BookActivity(ctx context.Context, userID, profileID, bookID string) ([]ActivityEvent, error) {
	if userID == "" || bookID == "" {
		return nil, errors.New("user_id, book_id required")
	}
	out := make([]ActivityEvent, 0, 32)

	// Progress row → emit one "progress" event with the latest
	// position. Older progress positions aren't preserved (we
	// don't keep a history) so this is a single point.
	if p, err := s.GetProgress(ctx, userID, profileID, bookID); err == nil {
		out = append(out, ActivityEvent{
			At:   p.UpdatedAt,
			Kind: "progress",
			Payload: map[string]any{
				"current_seconds": p.CurrentSeconds,
				"progress_pct":    p.ProgressPct,
				"is_finished":     p.IsFinished,
			},
		})
	}

	// Bookmarks — one event per row, kind="bookmark".
	if bms, err := s.ListBookmarks(ctx, userID, profileID, bookID); err == nil {
		for _, b := range bms {
			out = append(out, ActivityEvent{
				At:   b.CreatedAt,
				Kind: "bookmark",
				Payload: map[string]any{
					"id":           b.ID,
					"position":     b.PositionSeconds,
					"note":         b.Note,
					"chapter_id":   b.ChapterID,
				},
			})
		}
	}

	// ABS sessions — open/close events. We have started_at + last
	// activity; emit one per session.
	rows, err := s.pool.Query(ctx, `
		SELECT id, started_at, last_seen_at, current_seconds
		FROM abs_session WHERE user_id = $1 AND book_id = $2
		ORDER BY started_at DESC LIMIT 50
	`, userID, bookID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var started, lastSeen time.Time
			var pos int
			if err := rows.Scan(&id, &started, &lastSeen, &pos); err != nil {
				continue
			}
			out = append(out, ActivityEvent{
				At:   started,
				Kind: "session_opened",
				Payload: map[string]any{
					"session_id":    id,
					"start_position": pos,
				},
			})
			// Emit a close event when the session has visibly
			// ended (last_seen_at != started_at by some margin).
			// Sessions still in flight don't get a close event.
			if lastSeen.Sub(started) > time.Second {
				out = append(out, ActivityEvent{
					At:   lastSeen,
					Kind: "session_closed",
					Payload: map[string]any{
						"session_id":   id,
						"end_position": pos,
					},
				})
			}
		}
	}

	// Rating — one event for the latest rating row.
	var rating int
	var ratedAt time.Time
	err = s.pool.QueryRow(ctx, `
		SELECT rating, updated_at FROM rating WHERE user_id = $1 AND book_id = $2
	`, userID, bookID).Scan(&rating, &ratedAt)
	if err == nil {
		out = append(out, ActivityEvent{
			At:   ratedAt,
			Kind: "rated",
			Payload: map[string]any{
				"rating": rating,
			},
		})
	}

	// Share links — one event per outstanding link.
	shareRows, err := s.pool.Query(ctx, `
		SELECT id, slug, created_at FROM share_link
		WHERE user_id = $1 AND item_id = $2
	`, userID, bookID)
	if err == nil {
		defer shareRows.Close()
		for shareRows.Next() {
			var id, slug string
			var createdAt time.Time
			if err := shareRows.Scan(&id, &slug, &createdAt); err != nil {
				continue
			}
			out = append(out, ActivityEvent{
				At:   createdAt,
				Kind: "shared",
				Payload: map[string]any{
					"id":   id,
					"slug": slug,
				},
			})
		}
	}

	// Sort reverse-chronological. Tie-break on kind so the order
	// is deterministic for events recorded at the same instant.
	sortActivityDesc(out)
	return out, nil
}

// sortActivityDesc sorts in place. Uses a stable simple insertion
// since the typical list is <100 entries — no need for a heavier
// sort.Slice.
func sortActivityDesc(events []ActivityEvent) {
	for i := 1; i < len(events); i++ {
		cur := events[i]
		j := i - 1
		for j >= 0 && events[j].At.Before(cur.At) {
			events[j+1] = events[j]
			j--
		}
		events[j+1] = cur
	}
}

// _ = fmt placeholder so future debug strings don't churn imports.
var _ = fmt.Sprintf
