package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// BookmarkChange is one entry in the bookmark sync log. Identical
// shape to the ebook plugin's AnnotationChange — same row-level
// LWW + tombstones strategy applied to bookmark mutations.
type BookmarkChange struct {
	HLC         string
	UserID      string
	BookmarkID  string
	Op          string // "upsert" | "delete"
	Payload     json.RawMessage
	OriginNode  string
	CreatedAt   time.Time
}

func (s *Store) AppendBookmarkChange(ctx context.Context, c BookmarkChange) error {
	if c.HLC == "" || c.UserID == "" || c.BookmarkID == "" || c.Op == "" {
		return errors.New("hlc, user_id, bookmark_id, op required")
	}
	if len(c.Payload) == 0 {
		c.Payload = json.RawMessage("{}")
	}
	if c.OriginNode == "" {
		c.OriginNode = "unknown"
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO bookmark_change (hlc, user_id, bookmark_id, op, payload, origin_node)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (hlc) DO NOTHING
	`, c.HLC, c.UserID, c.BookmarkID, c.Op, c.Payload, c.OriginNode)
	if err != nil {
		return fmt.Errorf("append bookmark_change: %w", err)
	}
	return nil
}

func (s *Store) PullBookmarkChanges(ctx context.Context, userID, since string, limit int) ([]BookmarkChange, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, `
		SELECT hlc, user_id, bookmark_id, op, payload, origin_node, created_at
		FROM bookmark_change
		WHERE user_id = $1 AND hlc > $2
		ORDER BY hlc
		LIMIT $3
	`, userID, since, limit)
	if err != nil {
		return nil, fmt.Errorf("pull bookmark_change: %w", err)
	}
	defer rows.Close()
	var out []BookmarkChange
	for rows.Next() {
		var c BookmarkChange
		if err := rows.Scan(&c.HLC, &c.UserID, &c.BookmarkID, &c.Op, &c.Payload,
			&c.OriginNode, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
