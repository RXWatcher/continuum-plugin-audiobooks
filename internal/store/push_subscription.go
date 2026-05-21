package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// PushSubscription is one Web Push registration row. endpoint is
// the vendor-specific URL the browser provided; p256dh + auth are
// the ECDH public key + nonce needed to encrypt VAPID pushes for
// this endpoint.
type PushSubscription struct {
	ID         string
	UserID     string
	Endpoint   string
	P256dh     string
	Auth       string
	UserAgent  string
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

func (s *Store) UpsertPushSubscription(ctx context.Context, p PushSubscription) error {
	if p.ID == "" || p.UserID == "" || p.Endpoint == "" || p.P256dh == "" || p.Auth == "" {
		return errors.New("id, user_id, endpoint, p256dh, auth required")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO push_subscription (id, user_id, endpoint, p256dh, auth, user_agent)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,''))
		ON CONFLICT (endpoint) DO UPDATE SET
			user_id    = EXCLUDED.user_id,
			p256dh     = EXCLUDED.p256dh,
			auth       = EXCLUDED.auth,
			user_agent = EXCLUDED.user_agent
	`, p.ID, p.UserID, p.Endpoint, p.P256dh, p.Auth, p.UserAgent)
	if err != nil {
		return fmt.Errorf("upsert push_subscription: %w", err)
	}
	return nil
}

func (s *Store) ListPushSubscriptions(ctx context.Context, userID string) ([]PushSubscription, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, endpoint, p256dh, auth, COALESCE(user_agent,''),
		       created_at, last_used_at
		FROM push_subscription WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list push_subscription: %w", err)
	}
	defer rows.Close()
	var out []PushSubscription
	for rows.Next() {
		var p PushSubscription
		if err := rows.Scan(&p.ID, &p.UserID, &p.Endpoint, &p.P256dh, &p.Auth,
			&p.UserAgent, &p.CreatedAt, &p.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan push_subscription: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) DeletePushSubscription(ctx context.Context, id, userID string) error {
	if id == "" || userID == "" {
		return errors.New("id, user_id required")
	}
	_, err := s.pool.Exec(ctx, `
		DELETE FROM push_subscription WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return fmt.Errorf("delete push_subscription: %w", err)
	}
	return nil
}

// MarkPushSubscriptionUsed updates last_used_at to now. Called by
// the dispatcher on successful send so the SPA can show "last
// pinged on …".
func (s *Store) MarkPushSubscriptionUsed(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE push_subscription SET last_used_at = now() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("mark push_subscription used: %w", err)
	}
	return nil
}
