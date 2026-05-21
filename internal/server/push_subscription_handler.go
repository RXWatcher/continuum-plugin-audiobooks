package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"

	"github.com/ContinuumApp/continuum-plugin-audiobooks/internal/auth"
	"github.com/ContinuumApp/continuum-plugin-audiobooks/internal/store"
)

// Web Push subscription surface. The browser's Push API hands the
// SPA a {endpoint, keys: {p256dh, auth}} payload; the SPA POSTs
// it here so the notification dispatcher can later send VAPID-
// encrypted notifications to that endpoint.
//
// VAPID keys configured via env: VAPID_PUBLIC_KEY exposed to
// clients via /push/vapid-key; VAPID_PRIVATE_KEY stays server-
// side for the dispatcher's signing.

func (s *Server) mountPushSubscriptionRoutes(r chi.Router) {
	r.Get("/push/vapid-key", s.handleVAPIDPublicKey)
	r.Get("/me/push-subscriptions", s.handleListPushSubscriptions)
	r.Post("/me/push-subscriptions", s.handleCreatePushSubscription)
	r.Delete("/me/push-subscriptions/{id}", s.handleDeletePushSubscription)
}

// handleVAPIDPublicKey returns the operator-configured public key
// the browser needs to subscribe. No auth gate — the public key
// is, by design, public. Empty response when unconfigured so the
// SPA can show "push notifications not available."
func (s *Server) handleVAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"public_key": strings.TrimSpace(os.Getenv("VAPID_PUBLIC_KEY")),
	})
}

type pushSubscriptionBody struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func (s *Server) handleListPushSubscriptions(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.RequireUser(w, r)
	if !ok {
		return
	}
	rows, err := s.d.Store.ListPushSubscriptions(r.Context(), id.UserID)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	// Strip keys from the listing — they're sensitive and the SPA
	// never needs to re-read them.
	out := make([]map[string]any, 0, len(rows))
	for _, p := range rows {
		out = append(out, map[string]any{
			"id":           p.ID,
			"endpoint":     truncate(p.Endpoint, 80),
			"user_agent":   p.UserAgent,
			"created_at":   p.CreatedAt.UnixMilli(),
			"last_used_at": p.LastUsedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) handleCreatePushSubscription(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.RequireUser(w, r)
	if !ok {
		return
	}
	var body pushSubscriptionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Endpoint == "" || body.Keys.P256dh == "" || body.Keys.Auth == "" {
		writeError(w, http.StatusBadRequest, "endpoint + keys required")
		return
	}
	sub := store.PushSubscription{
		ID:        ulid.Make().String(),
		UserID:    id.UserID,
		Endpoint:  body.Endpoint,
		P256dh:    body.Keys.P256dh,
		Auth:      body.Keys.Auth,
		UserAgent: r.UserAgent(),
	}
	if err := s.d.Store.UpsertPushSubscription(r.Context(), sub); err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         sub.ID,
		"endpoint":   truncate(sub.Endpoint, 80),
		"created_at": sub.CreatedAt.UnixMilli(),
	})
}

func (s *Server) handleDeletePushSubscription(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.RequireUser(w, r)
	if !ok {
		return
	}
	if err := s.d.Store.DeletePushSubscription(r.Context(), chi.URLParam(r, "id"), id.UserID); err != nil {
		writeInternal(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
