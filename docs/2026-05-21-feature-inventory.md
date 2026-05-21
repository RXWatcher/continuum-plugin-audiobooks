# Feature inventory — audiobooks + ebooks plugins

Date: 2026-05-21
Status: final after triage. Cuts have been deleted from `main` —
migrations dropped, store + handler + frontend code removed.

This is the authoritative list of what these two plugins do. Anything
not here either never existed or was explicitly cut.

---

## Audiobooks plugin

### ABS API compatibility

- Dual-mount routes (root + `/api/*` + `/abs/api/*`) so the official
  ABS apps connect without path tweaks.
- Full login envelope (`permissions`, `librariesAccessible`,
  `mediaProgress`, `bookmarks`, `serverSettings`, `ereaderDevices`).
- `x-refresh-token` header convention + `/api/authorize` re-auth.
- ServerVersion 2.26 (unlocks the mobile app's JWT path).
- Socket `init` / `auth_failed` events + `{data}` wrapper on
  `user_item_progress_updated`.
- Library detail wrapper + personalized shelf entity shapes.
- Podcast personalized shelves (recent-episodes, newest-podcasts,
  listen-again).
- Bookmarks CRUD on `/me/item/{id}/bookmark`.
- Download endpoint `/api/items/{id}/file/{ino}` with Range +
  per-ext MIME.
- `/me/items-in-progress` + Continue-Listening management.
- Library multi-bucket search (`{book, podcast, series, authors,
  tags}`).
- Plural Socket.io events (`items_added`, `library_updated`,
  `episode_download_finished`).
- Collections CRUD at upstream paths.
- Playlists CRUD with episode-scoped entries.
- Custom metadata providers (admin CRUD + proxied search).
- RSS-feed-publish — item / series / collection renderings.
- Share links with public audio bytes.

### Features

- Sleep-timer with 30-second fade.
- Smart Collections — rule DSL, evaluator, CRUD.
- Embedding-based similar-items (pgvector + HNSW; OpenAI / Gemini /
  Ollama).
- Reading streak counter.
- Reading-session telemetry + heatmap + year-in-review.
- Reading goals (books + hours).
- Per-book activity timeline.
- Notification preferences.
- Content restrictions / family mode.

### Frontend

- Command palette (Cmd-K).
- Keyboard shortcut help (?).
- Atmosphere mode overlay.

---

## Ebooks plugin

### Backend

- Smart Collections (DSL + evaluator + CRUD).
- Embedding-based similar-items.
- `foliate-js` vendored locally.
- Content restrictions / family mode.
- Custom metadata providers.
- Send-to-ereader (device registry + SMTP send).
- Readwise.io export.
- Hardcover.app sync.
- Metadata enrichment (OpenLibrary + Google Books).
- Dictionary lookup (Wiktionary).
- In-text translation (LibreTranslate-compatible).
- Custom font upload + serve.
- Reading streak counter.
- Reading goals (books).
- Year-in-review stats.
- Per-book activity timeline.
- Notification preferences.
- Share links.
- Scheduled cleanup tasks (expired share links + recommendation
  cache).

### Frontend

- Command palette (Cmd-K).
- Keyboard shortcut help (?).
- Atmosphere overlay component (built; not mounted in reader yet).
- Screen wake-lock hook (wired into reader).
- E-ink mode hook (CSS rules in place).
- TTS controller hook with MediaSession (built; not yet mounted as
  a reader button).
