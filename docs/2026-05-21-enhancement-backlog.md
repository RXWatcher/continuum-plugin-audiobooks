# Audiobooks + ebooks enhancement backlog

Date: 2026-05-21
Owner: audiobooks + ebooks plugins (shared)
Status: planning — discoveries from /opt/grimmory, /opt/readest, and
/opt/audiobookshelf{,-app} surveys

This doc is a single accountable place to record every concrete
enhancement we found during the comparative review of grimmory,
readest, and the upstream Audiobookshelf server + mobile app. The
intent is: nothing gets lost, every item is sized, and the user can
pick freely from the list rather than re-discovering anything later.

Items are grouped by effort. Within each tier they're ranked by
expected user-visible leverage. File / source references kept short
so the doc stays scannable.

## Quick wins — one session each

### Audiobook player (audiobooks plugin)
- **Sleep timer with end-of-chapter mode + 30-second fade** —
  grimmory's `audiobook-player.component.ts:929-1042`. Biggest
  perceived UX win on the audiobook side; the single most-requested
  feature in self-hosted audiobook apps.
- **Bookmark UI with auto-titled marks** ("Chapter 3 — 12:34") —
  grimmory `audiobook-player.component.ts:1063-1110` +
  `BookMarkController`. Pair with sleep timer; both touch the same
  player surface.
- **Smart cover fallback chain** (stored audiobook cover → embedded
  ID3 cover → portal library cover) — grimmory
  `audiobook-player.component.ts:893-905`. Makes empty covers rare.
- **5-second progress save heartbeat + immediate save on
  pause/end + MediaSession setPositionState** — grimmory
  `audiobook-player.component.ts:813-852`. Confirm our current
  intervals match this cadence.

### Ebook reader (ebooks plugin)
- **Footnote-as-popover** — readest's
  `FootnotePopup.tsx` + `foliate-js/footnotes.js`. Intercept
  footnote links, render inline. No more lose-your-place jumps.
- **Reading ruler overlay** (sliding band highlights N lines) —
  readest `ReadingRuler.tsx`. Surface polish that pays back
  immediately on long sessions.
- **Image-zoom viewer + table-viewer popovers** — readest
  `ImageViewer.tsx`, `TableViewer.tsx`. Same intercept pattern as
  footnotes.
- **Magnifier loupe on touch selection** (mobile) — readest
  `MagnifierLoupe.tsx`. Tiny code, huge mobile UX uplift.
- **Markdown annotation export** — readest
  `ExportMarkdownDialog.tsx`. One button, one route, one template.
- **Shareable quote-image generator** — readest
  `foliate-js/quote-image.js`. Cheap social sharing surface.
- **E-ink mode + screen wake-lock + Discord rich presence** —
  readest `useEinkMode.ts`, `useScreenWakeLock.ts`,
  `useDiscordPresence.ts`. Three small polish hooks bundled.

### Cross-cutting (both plugins)
- **Reading streak counter** — grimmory `UserStatsController.java`
  `getReadingStreak`. One endpoint, one badge. Engagement driver.
- **Keyboard-shortcut help dialog** (`?` opens overlay) — grimmory
  `shortcuts-help.component.ts`. Mostly text.
- **Atmosphere mode** (ambient background + audio loop themed per
  book) — readest `AtmosphereOverlay.tsx`. Polish, but users notice
  it in the first 30 seconds.
- **OKLCH-based theme palette generator** from a single base
  colour — readest `themes.ts:42`. Lets users tune the theme.
- **Disk-cache pattern with mtime-tolerance + per-key
  ReentrantLock** for sequential extraction — grimmory
  `ChapterCacheService.java:26-80`. Applies to CBX / EPUB chapter
  reads and to podcast feed parses.
- **Audit log with GeoIP country-code resolution** — grimmory
  `AuditService.java` + `GeoIpService.java`. Defence-in-depth and
  operator visibility.

## Medium effort — one to two days each

### Cross-cutting
- **Command palette (Cmd-K)** — fuzzy search across books,
  podcasts, libraries, shelves, nav actions. Both grimmory and
  readest ship one; readest's is smaller code. Single biggest
  navigation upgrade once libraries grow.
- **Reading-session telemetry + streak/heatmap stats dashboard** —
  grimmory `ReadingSessionController.java`, `UserStatsController.java`,
  23 charts but a useful subset (clock + heatmap + streak +
  completion funnel) is medium. Unique surface neither plugin has.
- **Magic Shelves**: rule-builder JSON DSL + reorderable dashboard
  scrollers ("Continue Listening", "Continue Reading", "Recently
  Added", "Discover New", custom rules) — grimmory
  `magic-shelf.service.ts` + dashboard components. The three
  together (shelves + scrollers + Cmd-K) change the product feel.
- **Content restrictions per user** (CATEGORY/TAG/AGE_RATING/
  CONTENT_RATING, ALLOW/DENY) — grimmory
  `ContentRestrictionController.java`. Unlocks family / child mode.
- **Metadata enrichment on import** (OpenLibrary + Google Books) —
  readest `services/metadata/`. Provider abstraction is clean.
- **Hardcover.app sync** (per-user, async progress push, BYO API
  key) — grimmory `HardcoverSyncService.java` + readest
  `services/hardcover/`. Reading-log social integration.
- **Sidecar JSON files** (export/import metadata adjacent to book
  file) with sync-status detection — grimmory
  `SidecarController.java`. Metadata survives plugin reset.

### Audiobook + podcast
- **TTS controller with MediaSession integration** — readest
  `services/tts/` + `useTTSMediaSession.ts`. Multi-engine (Web
  Speech / Edge TTS / native), sentence-level highlight on text,
  OS lock-screen controls. On the audiobooks side this becomes
  "TTS audiobook" mode for books with no audio.
- **Listening-specific stats** (peak hours, monthly pace,
  completion funnel, longest audiobooks) — grimmory
  `UserStatsController.java:234-352`.

### Ebook
- **Highlight styles + colours + per-highlight notes + export to
  Readwise** — readest
  `types/book.ts:19` + `ReadwiseClient.ts`. Annotation system is
  the feature ebook power-users grade a reader on.
- **Foliate / .mrexpt annotation import** — readest
  `services/annotation/providers/`. Existing-user migration story.
- **Dictionary stack** (StarDict / dictd / Wiktionary / Wikipedia
  / web-search) — readest `services/dictionaries/providers/`.
- **In-text translation popover** with DeepL / Google / Azure /
  Yandex + sentence-cache + inline parallel translation render —
  readest `services/translators/` + `useTextTranslation.ts`.
- **RSVP / speed-reading overlay** + **paragraph-focus mode** —
  readest `rsvp/`, `paragraph/`. Reading-style polish modes.
- **Custom font upload + sync** — grimmory
  `CustomFontController.java` + readest `CustomFonts.tsx`.
- **Annotation aggregation Notebook view** — grimmory
  `NotebookController.java`. Filters across all annotations.
- **KOSync conflict-resolver UI** (show user which side wins) —
  readest `KOSyncResolver.tsx`.

## Heavy — architectural reshape, but each transformative

- **Replace ebook SPA renderer with `foliate-js`** (MIT) — readest
  `packages/foliate-js/`. Unlocks CFI-based stable locations,
  fixed-layout EPUB, MOBI/KF8/CBZ/PDF support, paginated +
  scrolled modes, overlay API. This is the foundation under most
  of the readest ebook items above; pick it up first if doing
  more than two of them.
- **CFI-based location addressing** for progress + annotations —
  readest `epubcfi.js`. Comes free with `foliate-js`.
- **HLC + field-level LWW CRDT replica sync** — readest
  `libs/crdt.README.md`, `services/sync/replicaSyncManager.ts`.
  Replaces ad-hoc kosync with conflict-free progress +
  annotations + bookmarks + settings + custom fonts + OPDS
  catalogues across web and mobile.
- **Vector-embedding-based similar-book recommendations** with
  entity-similarity fallback (title 1.5, series 2.0, authors 3.0,
  categories 3.5, rating 0.6) cached as `similar_books_json` —
  grimmory `BookSimilarityService.java`, `BookVectorService.java`,
  `BookRecommendationService.java`.
- **BookDrop** — watched folder, auto-metadata-enrich, queue for
  admin review with bulk-edit + filename-pattern extraction —
  grimmory `BookdropFileController.java`,
  `service/bookdrop/`. More polished than what's in any other
  self-hosted library product.
- **AI sidebar with per-book RAG** (chunk → embed → retrieve →
  chat; Ollama or proxied gateway) — readest
  `services/ai/ragService.ts`. Ebook-only feature: "what happened
  in chapter 3?" — speculative but unique.
- **Document-to-EPUB conversion on import** (DOCX via mammoth,
  RTF, HTML, TXT, Readability article, full-page clip with
  bundled images) — readest `services/send/conversion/`. Big
  ingestion-quality upgrade.

## Current Top 5 picks (will be re-ranked after the ABS reviews)

1. **Sleep timer + end-of-chapter + fade** (Quick, audiobooks).
   Smallest code, biggest user-visible win.
2. **Command palette** (Medium, both). Single navigation upgrade
   that scales with library size.
3. **Reading-session telemetry + streak + small stats page**
   (Medium, both). Unique engagement surface, data foundation
   for recommendations / magic shelves.
4. **`foliate-js` renderer swap + highlight system with Markdown
   export** (Heavy + Medium, ebooks). One reader rewrite gives
   CFI + FXL + MOBI/KF8/CBZ/PDF + proper annotations together.
5. **Content restrictions / family / child mode** (Medium, both).
   Unlocks an audience we currently have zero product for.

## ABS server + mobile-app spec gaps

### Compatibility blockers — the official ABS apps don't currently "just work"

These were flagged by walking the upstream server's route tree and
diffing against the actual mobile-client fetch calls. If we want the
official Audiobookshelf mobile / web clients to connect cleanly, these
have to land.

- **Path-prefix mismatch (Heavy)**: mobile app builds URLs as
  `${serverAddress}/login`, `${serverAddress}/auth/refresh`,
  `${serverAddress}/logout`, `${serverAddress}/api/authorize`,
  `${serverAddress}/status`, `${serverAddress}/ping`. Our routes are
  under `/abs/api/*`. Either re-mount the ABS surface at server root,
  or document that operators must enter `https://host/abs` as the
  server URL and the app appends `/login` → `https://host/abs/login`
  (would still need to also mount `/abs/login` etc. as aliases). The
  simplest fix on the standalone listener: serve ABS routes at root
  AND under `/abs/api/*` so both client expectations work.
- **Sort-param convention (Quick)**: real ABS uses `?sort=<field>&desc=1`,
  not `sortBy=...&sortDesc=true`. We currently emit / parse the latter.
  The SPA and mobile clients silently sort wrong.
- **Login response envelope (Medium)**: must return
  `{user{accessToken, refreshToken, id, username, type, mediaProgress[],
  bookmarks[], librariesAccessible[], permissions{update,delete,download,
  accessExplicitContent}, token, isOldToken}, userDefaultLibraryId,
  serverSettings{version,language}, ereaderDevices}`. `serverSettings.version`
  must be a 3-part semver ≥ `2.26.0` for the JWT path; missing version
  forces "old token" mode in the app.
- **`POST /auth/refresh` with `x-refresh-token` header (Medium)**:
  empty body, returns `{user:{accessToken, refreshToken}}` —
  NOT a flat token pair. Without this, the app falls back to
  re-login on every expiry.
- **`POST /api/authorize` (Quick)**: re-auth handshake on app launch.
  Returns the same shape as `/login` minus the new tokens. Without
  this the app can't recover an existing session and re-prompts on
  every launch.
- **Socket `init` event after `auth` succeeds (Quick)**: we emit
  `auth_authorized`; the app listens for `init`. Plus emit
  `auth_failed` with `{message}` on bad token.
- **`user_item_progress_updated` payload wrapper (Quick)**: must be
  `{data: <mediaProgress>}` — we currently emit the bare object.
  One-line fix.
- **`GET /api/libraries/:id?include=filterdata` envelope (Quick)**:
  must return `{library, filterdata, issues, numUserPlaylists}`,
  not just `library`. The mobile library-detail call branches on
  this shape.
- **Personalized shelves entity shapes (Medium)**: `type:'series'`
  and `type:'authors'` shelves return series/author payloads with
  embedded `books[]` / `numBooks`, NOT library items. `type:'episode'`
  entities are LibraryItem with a `recentEpisode` field set. `total`
  is the population total, not `entities.length`.
- **Item ID format (Quick, but uncertain)**: real ABS uses UUID v4
  for item IDs. Our `li_<libraryID>:<base64>` form may fail strict
  UUID parsers in some clients. Mobile app uses a regex
  `[a-z0-9-]{36}` in CORS-preflight matching — needs verification
  whether other code paths gate on it. May need to migrate to
  UUID-shaped IDs (probably a colon-free, lowercase shape that
  passes the regex but still encodes our library reference).

### Compatibility — endpoints the app calls that we don't implement

These are not strict blockers (the app degrades to "feature
unavailable" rather than crashes when they're missing), but every
absent route is an empty button or a silent failure.

User / Me:
- `GET /api/me/items-in-progress` — returns `{libraryItems:[…]}`. App
  uses this for the "Continue Listening" home shelf on mobile.
- `GET /api/me/listening-sessions`, `/api/me/listening-stats`,
  `/api/me/stats/year/:year` — year-in-review screen mobile shows
  every January.
- `GET /api/me/progress/:id/remove-from-continue-listening` (yes,
  GET not DELETE), `GET /api/me/series/:id/remove-from-continue-listening`,
  `GET /api/me/series/:id/readd-to-continue-listening`,
  `DELETE /api/me/progress/:id`, `PATCH /api/me/progress/batch/update`.
- `POST/PATCH/DELETE /api/me/item/:id/bookmark[/time]` — bookmarks
  CRUD; app has a "Bookmarks" tab that doesn't work without these.
- `POST /api/session/local`, `POST /api/session/local-all` — offline
  session sync. Without these the listening progress recorded
  offline-on-the-plane is silently lost.
- `GET /api/session/:id` — Android player refreshes session metadata
  via this.
- `POST /api/me/ereader-devices` — user's registered devices for
  "Send to E-reader".

Library:
- `PATCH /api/libraries/:id`, `DELETE /api/libraries/:id`,
  `POST /api/libraries`, `POST /api/libraries/order` — library CRUD
  + reorder (admin UI in mobile).
- `GET /api/libraries/:id/search?q=` — multi-bucket return
  `{book[], podcast[], series[], authors[], tags[]}`. The mobile
  search bar is dead without this.
- `GET /api/libraries/:id/recent-episodes?limit=&page=` — podcast
  Latest tab.
- `GET /api/libraries/:id/collections`, `/playlists`, `/stats`,
  `/narrators`, `/matchall`, `/recent-episodes`, `/opml`,
  `/podcast-titles`, `/download`, `/episode-downloads`.

Items:
- `GET /api/items/:id/file/:ino/download` (Bearer or `?token=`),
  `GET /api/items/:id/file/:ino?token=` (iOS streaming) — primary
  download + iOS-streaming paths. Range support required. Mime by
  file ext (mp3/m4b/mp4/epub/pdf/jpg).
- `POST /api/items/batch/{delete,update,get,quickmatch,scan}` —
  bulk actions.
- `GET /api/items/:id/metadata-object`, `POST /api/items/:id/chapters`,
  `GET /api/items/:id/ffprobe/:fileid`, `GET/DELETE /api/items/:id/file/:fileid`,
  `PATCH /api/items/:id/tracks`.
- `GET /api/items/:id/ebook[/:fileid]`, `PATCH /api/items/:id/ebook/:fileid/status` —
  ebook fetch + read-status update (audiobook plugin can ignore;
  ebook plugin needs this if we ever serve ebooks through the
  ABS-shaped surface).

Collections / Playlists / Series:
- Full CRUD families absent — `/api/collections*`, `/api/playlists*`,
  `/api/playlists/collection/:collectionId`,
  `PATCH /api/series/:id`. Mobile has UI for all of these.

Podcasts:
- `POST /api/podcasts` (add by feed URL) — what the mobile "Add
  Podcast" flow calls. Currently our admin endpoint is at
  `POST /api/v1/admin/podcasts`; need a mirror under `/api/podcasts`.
- `POST /api/podcasts/feed` (probe feed URL before add).
- `POST /api/podcasts/opml/{parse,create}`.
- `POST /api/podcasts/:id/{checknew,downloads,clear-queue,
  search-episode,download-episodes,match-episodes}`.
- `GET/PATCH/DELETE /api/podcasts/:id/episode/:episodeId`.

Sessions / Stats:
- `GET /api/sessions` (admin), `DELETE /api/sessions/:id`,
  `GET /api/sessions/open`, `POST /api/sessions/batch/delete`,
  `GET /api/stats/server`.

Authors / Search:
- `PATCH/DELETE /api/authors/:id`, `POST /api/authors/:id/match`,
  `POST/DELETE /api/authors/:id/image`.
- `/api/search/{covers,books,podcast,authors,chapters,providers}`.

Ereader / RSS / Tools / Custom-metadata / Share:
- `POST /api/emails/send-ebook-to-device` (the "Send to E-reader"
  workflow), `POST /api/emails/ereader-devices` (admin),
  `GET/PATCH /api/emails/settings`, `POST /api/emails/test`.
- `GET /api/feeds`, `POST /api/feeds/item/:itemId/open` (and
  `/collection/:id/open`, `/series/:id/open`),
  `POST /api/feeds/:id/close` — open feed = "publish this audiobook
  / series as a podcast RSS feed."
- `POST /api/tools/item/:id/encode-m4b` (+ DELETE cancel),
  `POST /api/tools/item/:id/embed-metadata`,
  `POST /api/tools/batch/embed-metadata`.
- `GET/POST/DELETE /api/custom-metadata-providers[/:id]` — provider
  contract documented at
  `/opt/audiobookshelf/custom-metadata-provider-specification.yaml`.
- `POST /api/share/mediaitem`, `DELETE /api/share/mediaitem/:id` +
  `/public/share/:slug/{,track/:index,cover,download,progress}`.

Misc / admin:
- `/api/upload`, `/api/tasks`, `PATCH /api/settings`, `/api/sorting-prefixes`,
  `/api/authorize`, `/api/tags(+rename/:tag)`, `/api/genres(+rename/:genre)`,
  `/api/validate-cron`, `/api/auth-settings`, `/api/watcher/update`,
  `/api/logger-data`, `/api/api-keys` (CRUD), `/api/notifications(+/test+/:id/test)`,
  `/api/notificationdata`, `/api/backups + /api/backups/path`,
  `/api/filesystem(+/pathexists)`, `/api/cache/{purge,items/purge}`.

Auth:
- `GET /auth/openid` + callback + `/auth/openid/config` +
  `/auth/openid/mobile-redirect` — full OIDC SSO flow.

### Socket.io event gaps

We emit `user_item_progress_updated` (need `{data}` wrapper) /
`user_session_updated` / `user_session_open` / `user_session_closed` /
`listener_count` / `item_added`. The mobile + web clients listen for:

- `init` — payload sent right after `auth` succeeds; treated as
  the "authenticated marker" by the app.
- `auth_failed` (payload `{message}`).
- `user_updated`, `user_added`, `user_stream_update`, `user_online`,
  `user_offline` (admin panel needs these).
- `item_updated`, `item_removed`, `items_added`, `items_updated`
  (plural; SPA uses for batch-refresh after scans).
- `series_added`, `series_updated`, `series_removed`,
  `author_added`, `author_updated`, `author_removed`.
- `library_added`, `library_updated`, `library_removed`.
- `playlist_added`, `playlist_updated`, `playlist_removed`,
  `collection_added`, `collection_updated`, `collection_removed`.
- `rss_feed_open`, `rss_feed_closed`.
- `task_started`, `task_finished`, `track_started`, `track_progress`,
  `track_finished`, `metadata_embed_queue_update`,
  `batch_quickmatch_complete`, `backup_applied`.
- `episode_download_queued`, `episode_download_started`,
  `episode_download_finished`.
- `custom_metadata_provider_added`, `_removed`.
- `ereader-devices-updated` (note hyphen, not underscore).
- `stream_open`, `stream_closed`, `stream_reset` (HLS).
- `admin_message`, `cancel_scan` inbound handler, cover-search
  streaming events (`search_covers` / `cover_search_result` /
  `_complete` / `_error` / `_cancelled`).

Auth payload subtlety: app emits `socket.emit('auth', token_string)` —
**bare JWT string, not an object**. Our current handler accepts both;
fine. But it expects the server to emit `init` on success, not
`auth_authorized`.

### Headers expected by the official clients

- `Authorization: Bearer <accessToken>` — standard.
- `x-refresh-token: <refreshToken>` — on `/auth/refresh` (header,
  not body) and on `/logout` (to delete that specific session).
- `x-return-tokens: true` — on `/login` to ask server to include
  the refresh token in the response.
- `Content-Type: application/json` whenever there's a body.
- No `User-Agent` or `X-Continuum-*` headers expected. No `Range`
  header sent by the JS; native (Capacitor / ExoPlayer / AVPlayer)
  does Range internally.

## Recomputed Top 10 — across all four sources

Re-ranking after the ABS reviews. Items move up when they're
strict compatibility blockers; the grimmory / readest items
remain but get pushed down behind the "works with the mobile
app" essentials.

1. **Re-mount ABS routes at server root + `/abs/api/*` aliases** —
   `/login`, `/auth/refresh`, `/logout`, `/status`, `/ping`,
   `/api/authorize`, `/api/me/*`, `/api/libraries/*`, `/api/items/*`.
   Standalone-listener only (host proxy stays at `/abs/api/*` for
   backwards compat). Heavy. Compatibility blocker.
2. **Login + authorize + refresh response shapes** to match
   `{user:{...}, userDefaultLibraryId, serverSettings{version,
   language}, ereaderDevices[]}` and the `x-refresh-token` /
   `x-return-tokens` header conventions. Medium. Blocker.
3. **Sort param convention fix** (`?sort=&desc=1`). Quick.
   Blocker (silent — clients sort wrong without it).
4. **Socket `init` event after auth + `{data}` wrapper on
   `user_item_progress_updated` + `auth_failed` on bad token**.
   Quick. Blocker.
5. **`GET /api/libraries/:id?include=filterdata` wrapper +
   personalized shelf entity shapes (series/authors/episode types)**.
   Medium. Blocker (mobile library + home tab broken without this).
6. **Bookmarks CRUD** (`POST/PATCH /api/me/item/:id/bookmark`,
   `DELETE /api/me/item/:id/bookmark/:time`). Quick. Polish that
   feels like a blocker — Bookmarks tab is empty otherwise.
7. **Downloads endpoint** (`GET /api/items/:id/file/:ino/download`
   + `…/file/:ino?token=` for iOS streaming) with Range support
   and per-file-ext mime types. Heavy. Blocker for offline.
8. **Items-in-progress + remove-from-continue-listening +
   listening-sessions/stats** (`GET /api/me/items-in-progress`,
   `GET /api/me/progress/:id/remove-from-continue-listening`,
   `DELETE /api/me/progress/:id`, `GET /api/me/listening-sessions`,
   `GET /api/me/listening-stats`). Medium. Blocker for the
   "Continue Listening" management surface.
9. **Library multi-bucket search** (`GET /api/libraries/:id/search?q=` →
   `{book[], podcast[], series[], authors[], tags[]}`). Quick.
   Blocker for the search bar.
10. **Sleep timer + end-of-chapter + 30-second fade in the
    audiobook player** (grimmory's UX feature). Quick. Highest
    leverage non-compatibility item.

Items pushed down (still worth doing, just not in top 10):

- Command palette (Medium, both plugins)
- Reading-session telemetry + streak / heatmap dashboard (Medium, both)
- Highlight system with Markdown export (Medium, ebooks)
- `foliate-js` renderer swap (Heavy, ebooks)
- Content restrictions / family mode (Medium, both)
- Magic Shelves + dashboard scrollers (Medium-Heavy, both)
- TTS controller with MediaSession (Medium, both)
- Vector recommendations (Heavy, both)
- BookDrop watched folder (Heavy, both)
- HLC + CRDT replica sync (Heavy, both)
- Plural Socket.io events (Medium, batch after scans)
- Collections + Playlists CRUD (Heavy, ABS compat)
- Send-to-ereader + RSS-feed-publish (Heavy, ABS compat)
- Custom metadata providers (Medium, ABS compat)

## Notes for execution

- The path-prefix change (#1) and ID-format change (#11 implicit
  in the ABS server review) are both schema-touching architectural
  edits. They should probably land together as one coordinated
  change, since `id` and URL prefix together define the public
  contract.
- Many items pair naturally:
  - sleep timer + bookmark UI + smart cover fallback ship as one
    audiobook-player polish commit.
  - `/api/libraries/:id` wrapper + personalized shelf shapes ship
    together (same surface).
  - Socket `init` + `auth_failed` + payload wrapper fix ship as
    one Socket.io-compatibility commit.
  - Items-in-progress + remove-from-CL + DELETE progress ship as
    one /me/* compatibility commit.

## Session progress — 2026-05-21

### Shipped this session

ABS Top-10 compatibility (audiobooks):
- Dual-mount routes (server root + /api + legacy /abs/api)
- Rich login envelope + x-refresh-token header + /api/authorize
- Sort param convention (audit — already correct)
- Socket init / auth_failed / {data} wrapper on user_item_progress_updated
- Library detail wrapper + personalized shelf shapes + podcast home tab
- Bookmarks CRUD on /me/item/{id}/bookmark
- Download endpoint /api/items/{id}/file/{ino} with Range + per-ext mime
- /me/items-in-progress + Continue-Listening management
- Library multi-bucket search
- Sleep-timer 30-second fade + ServerVersion 2.26

ABS Top-10 follow-ups:
- Plural Socket.io events (items_added, library_updated,
  episode_download_finished)
- Bookmark UI in SPA (auto-titles + inline rename, PATCH backend)
- Command palette (Cmd-K) in both audiobooks + ebooks

Rule-based Smart Collections + Embedding Recommendations:
- DSL + evaluator mirroring host's QueryDefinition shape, in
  audiobooks AND ebooks plugins
- pgvector schema + HNSW index + recommendation cache, both plugins
- OpenAI / Gemini / Ollama embedding client + audiobook + ebook
  text builders
- /api/items/{id}/similar (audiobooks) + /me/books/{id}/similar
  (ebooks) endpoints, cache-backed
- Smart Collection CRUD + evaluated /items endpoint, both plugins
- Migration 0018 in ebooks is pgvector-tolerant for dev/test envs

Other:
- Reading streak counter — /me/streak in both plugins
- Keyboard shortcut help overlay (?) in both plugins

### Not yet shipped — comprehensive remaining backlog

#### Quick (1 session each)

- **Atmosphere mode** — ambient background + audio loop themed per
  book. Both plugins.
- **Footnote popovers** (ebooks) — readest FootnotePopup pattern.
- **Reading ruler** (ebooks) — readest ReadingRuler.tsx.
- **Image-zoom viewer + table-viewer popovers** (ebooks).
- **Magnifier loupe on touch selection** (ebooks mobile).
- **E-ink mode + screen wake-lock + Discord rich presence** (ebooks).
- **Markdown annotation export** (ebooks).
- **Shareable quote-image generator** (ebooks).
- **Smart cover fallback chain** — embedded ID3 cover hop on the
  audiobooks side (requires ID3 parsing).

#### Medium (1–2 days each)

- **TTS controller + MediaSession** (both) — multi-engine, sentence-
  level highlight, OS lock-screen controls.
- **Reading-session telemetry + streak/heatmap dashboard** —
  streak counter shipped; heatmap + charts on a dedicated stats
  page is a follow-up.
- **Highlight system + Markdown / Readwise export** (ebooks).
- **Content restrictions / family / child mode** (both) —
  CATEGORY/TAG/AGE_RATING with ALLOW/DENY, admin-managed.
- **Dictionary + in-text translation popover** (ebooks).
- **Hardcover.app sync** (both) — BYO API key.
- **Sidecar JSON files** (both) — metadata adjacent to book file.
- **Custom font upload + sync** (ebooks).
- **Annotation aggregation Notebook view** (ebooks).
- **KOSync conflict-resolver UI** (ebooks).
- **Metadata enrichment on import** — OpenLibrary + Google Books.
- **Custom metadata providers** (ABS compat).

#### Heavy (architectural; each likely its own session)

- **`foliate-js` renderer swap** (ebooks) — unlocks CFI, FXL,
  MOBI/KF8/CBZ/PDF, paginated + scrolled, overlay API. Foundation
  for half the readest items above.
- **HLC + field-level LWW CRDT replica sync** — replaces ad-hoc
  kosync with conflict-free annotations + progress + bookmarks +
  settings across web + mobile.
- **BookDrop watched folder** — auto-metadata-enrich, admin
  review queue, bulk-edit, filename-pattern extraction.
- **Collections + Playlists CRUD** (ABS compat) — full domain
  models with matching socket events.
- **Send-to-ereader + RSS-feed-publish** (ABS compat) — SMTP +
  device-registration + feed generation.
- **AI sidebar with per-book RAG** (ebooks) — chunk → embed →
  retrieve → chat.
- **Document-to-EPUB conversion on import** — DOCX / RTF / HTML /
  Readability article.

### Notes for the next session

- pgvector is required in production for both plugins. Audiobooks
  uses testcontainers + pgvector/pgvector:pg17. Ebooks uses
  TEST_DATABASE_URL env var; migration 0018 is pgvector-tolerant
  for dev/test envs without it.
- All session work is on origin/main: audiobooks `e5d1cb7`,
  ebooks `2b5625b`.

## Source references

- /opt/grimmory README.md, `audiobook-player.component.ts`,
  `UserStatsController.java`, `MagicShelfController.java`,
  `BookSimilarityService.java`, `ContentRestrictionController.java`.
- /opt/readest README.md, `packages/foliate-js/`,
  `apps/readest-app/src/services/{sync,tts,ai,dictionaries,
  translators}`, `apps/readest-app/src/app/reader/components/`.
- /opt/audiobookshelf `server/routers/ApiRouter.js`,
  `server/SocketAuthority.js`, `server/controllers/LibraryController.js`,
  `custom-metadata-provider-specification.yaml`.
- /opt/audiobookshelf-app `pages/connect.vue`,
  `components/connect/ServerConnectForm.vue`, `store/{libraries,user}.js`,
  `plugins/{server,api}.js`, `players/AbsAudioPlayer.js`,
  `android/app/src/main/java/com/audiobookshelf/app/{server,player}/`,
  `ios/App/Shared/{player,services}/`.
