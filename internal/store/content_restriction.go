package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ContentRestriction is the per-user list of blocked dimensions.
// Empty arrays mean "no filter on that dimension" — a brand-new user
// with no restriction row matches every item. Admin writes one row
// per restricted user; everyone else has no row and passes through.
type ContentRestriction struct {
	UserID           string    `json:"user_id"`
	BlockedGenres    []string  `json:"blocked_genres"`
	BlockedTags      []string  `json:"blocked_tags"`
	BlockedAuthors   []string  `json:"blocked_authors"`
	BlockedNarrators []string  `json:"blocked_narrators"`
	BlockedLibraries []int64   `json:"blocked_libraries"`
	ExplicitBlocked  bool      `json:"explicit_blocked"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GetContentRestriction returns the user's restriction row.
// Returns ErrNotFound when no restriction is set (the catalog
// handlers treat that as "allow everything").
func (s *Store) GetContentRestriction(ctx context.Context, userID string) (ContentRestriction, error) {
	if userID == "" {
		return ContentRestriction{}, errors.New("user_id required")
	}
	row := s.pool.QueryRow(ctx, `
		SELECT user_id, blocked_genres, blocked_tags, blocked_authors,
		       blocked_narrators, blocked_libraries, explicit_blocked,
		       created_at, updated_at
		FROM content_restriction WHERE user_id = $1
	`, userID)
	var r ContentRestriction
	if err := row.Scan(&r.UserID, &r.BlockedGenres, &r.BlockedTags,
		&r.BlockedAuthors, &r.BlockedNarrators, &r.BlockedLibraries,
		&r.ExplicitBlocked, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ContentRestriction{}, ErrNotFound
		}
		return ContentRestriction{}, fmt.Errorf("get content_restriction: %w", err)
	}
	return r, nil
}

// UpsertContentRestriction sets the restriction for one user.
// Empty arrays + ExplicitBlocked=false effectively clears the
// filter — equivalent to deleting the row but keeps the metadata
// timestamps for audit. Use DeleteContentRestriction to fully
// remove.
func (s *Store) UpsertContentRestriction(ctx context.Context, r ContentRestriction) error {
	if r.UserID == "" {
		return errors.New("user_id required")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO content_restriction (
			user_id, blocked_genres, blocked_tags, blocked_authors,
			blocked_narrators, blocked_libraries, explicit_blocked
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			blocked_genres    = EXCLUDED.blocked_genres,
			blocked_tags      = EXCLUDED.blocked_tags,
			blocked_authors   = EXCLUDED.blocked_authors,
			blocked_narrators = EXCLUDED.blocked_narrators,
			blocked_libraries = EXCLUDED.blocked_libraries,
			explicit_blocked  = EXCLUDED.explicit_blocked,
			updated_at        = now()
	`, r.UserID, r.BlockedGenres, r.BlockedTags, r.BlockedAuthors,
		r.BlockedNarrators, r.BlockedLibraries, r.ExplicitBlocked)
	if err != nil {
		return fmt.Errorf("upsert content_restriction: %w", err)
	}
	return nil
}

// DeleteContentRestriction removes the user's restriction row.
// Idempotent.
func (s *Store) DeleteContentRestriction(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user_id required")
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM content_restriction WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete content_restriction: %w", err)
	}
	return nil
}

// ListContentRestrictions returns every restriction row (admin
// surface). Ordered by user_id ascending so the admin UI renders
// consistently across reloads.
func (s *Store) ListContentRestrictions(ctx context.Context) ([]ContentRestriction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT user_id, blocked_genres, blocked_tags, blocked_authors,
		       blocked_narrators, blocked_libraries, explicit_blocked,
		       created_at, updated_at
		FROM content_restriction ORDER BY user_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list content_restriction: %w", err)
	}
	defer rows.Close()
	var out []ContentRestriction
	for rows.Next() {
		var r ContentRestriction
		if err := rows.Scan(&r.UserID, &r.BlockedGenres, &r.BlockedTags,
			&r.BlockedAuthors, &r.BlockedNarrators, &r.BlockedLibraries,
			&r.ExplicitBlocked, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan content_restriction: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// AllowsItem reports whether the given item passes this user's
// restriction. Pure function — pass any subset of the audiobook
// metadata fields the handler has on hand. Empty restriction
// returns true (no filter set).
//
// The match is case-insensitive on string fields and an OR across
// each blocked list — if ANY blocked genre is in the item's genres,
// the item is blocked.
func (r ContentRestriction) AllowsItem(libraryID int64, genres, tags, authors, narrators []string, explicit bool) bool {
	if r.UserID == "" {
		return true
	}
	if r.ExplicitBlocked && explicit {
		return false
	}
	for _, id := range r.BlockedLibraries {
		if id == libraryID {
			return false
		}
	}
	if anyMatch(genres, r.BlockedGenres) {
		return false
	}
	if anyMatch(tags, r.BlockedTags) {
		return false
	}
	if anyMatch(authors, r.BlockedAuthors) {
		return false
	}
	if anyMatch(narrators, r.BlockedNarrators) {
		return false
	}
	return true
}

// anyMatch returns true when haystack contains any element of
// needles, case-insensitive.
func anyMatch(haystack, needles []string) bool {
	if len(haystack) == 0 || len(needles) == 0 {
		return false
	}
	hs := make(map[string]struct{}, len(haystack))
	for _, s := range haystack {
		hs[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	for _, n := range needles {
		if _, ok := hs[strings.ToLower(strings.TrimSpace(n))]; ok {
			return true
		}
	}
	return false
}
