package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SmartCollection is the row shape in the smart_collection table. The
// QueryDef is stored as JSONB; we keep it as json.RawMessage in Go so
// the smartcoll package's DSL types own the schema-shape side of the
// contract.
type SmartCollection struct {
	ID          string
	UserID      string
	ProfileID   string
	Name        string
	Description string
	Color       string
	IsPublic    bool
	IsPinned    bool
	QueryDef    json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertSmartCollection inserts or updates by id. The caller mints the
// id on creation (typically a ULID) and reuses it on updates.
func (s *Store) UpsertSmartCollection(ctx context.Context, c SmartCollection) error {
	if c.ID == "" || c.UserID == "" || c.Name == "" {
		return errors.New("id, user_id, name required")
	}
	if len(c.QueryDef) == 0 {
		c.QueryDef = json.RawMessage([]byte("{}"))
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO smart_collection (
			id, user_id, profile_id, name, description, color, is_public, is_pinned, query_def
		) VALUES ($1, $2, $3, $4, NULLIF($5,''), NULLIF($6,''), $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name        = EXCLUDED.name,
			description = EXCLUDED.description,
			color       = EXCLUDED.color,
			is_public   = EXCLUDED.is_public,
			is_pinned   = EXCLUDED.is_pinned,
			query_def   = EXCLUDED.query_def,
			updated_at  = now()
	`, c.ID, c.UserID, c.ProfileID, c.Name, c.Description, c.Color, c.IsPublic, c.IsPinned, c.QueryDef)
	if err != nil {
		return fmt.Errorf("upsert smart_collection: %w", err)
	}
	return nil
}

// GetSmartCollection reads by id. Returns ErrNotFound on miss.
func (s *Store) GetSmartCollection(ctx context.Context, id string) (SmartCollection, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, profile_id, name, COALESCE(description,''), COALESCE(color,''),
		       is_public, is_pinned, query_def, created_at, updated_at
		FROM smart_collection WHERE id = $1
	`, id)
	var c SmartCollection
	if err := row.Scan(&c.ID, &c.UserID, &c.ProfileID, &c.Name, &c.Description, &c.Color,
		&c.IsPublic, &c.IsPinned, &c.QueryDef, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SmartCollection{}, ErrNotFound
		}
		return SmartCollection{}, fmt.Errorf("get smart_collection: %w", err)
	}
	return c, nil
}

// ListSmartCollections returns all collections visible to the user in the
// given profile: the user's own (matching profile_id) + any is_public rows
// from other users. Pinned first, then alpha by name. limit caps the result;
// 0/negative → 500. profileID "" = primary profile.
func (s *Store) ListSmartCollections(ctx context.Context, userID, profileID string, limit int) ([]SmartCollection, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}
	if limit <= 0 {
		limit = 500
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, profile_id, name, COALESCE(description,''), COALESCE(color,''),
		       is_public, is_pinned, query_def, created_at, updated_at
		FROM smart_collection
		WHERE (user_id = $1 AND profile_id = $2) OR is_public = TRUE
		ORDER BY (user_id = $1 AND profile_id = $2) DESC, is_pinned DESC, LOWER(name)
		LIMIT $3
	`, userID, profileID, limit)
	if err != nil {
		return nil, fmt.Errorf("list smart_collections: %w", err)
	}
	defer rows.Close()
	var out []SmartCollection
	for rows.Next() {
		var c SmartCollection
		if err := rows.Scan(&c.ID, &c.UserID, &c.ProfileID, &c.Name, &c.Description, &c.Color,
			&c.IsPublic, &c.IsPinned, &c.QueryDef, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan smart_collection: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// DeleteSmartCollection removes by id; user_id and profile_id pin ownership
// so a caller can't delete another user's (or profile's) collection without
// an admin bypass at the handler level.
func (s *Store) DeleteSmartCollection(ctx context.Context, id, userID, profileID string) error {
	if id == "" || userID == "" {
		return errors.New("id, user_id required")
	}
	_, err := s.pool.Exec(ctx, `
		DELETE FROM smart_collection WHERE id = $1 AND user_id = $2 AND profile_id = $3
	`, id, userID, profileID)
	if err != nil {
		return fmt.Errorf("delete smart_collection: %w", err)
	}
	return nil
}
