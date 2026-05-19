# Audiobook Player QA Checklist

Use this for real-device/manual validation after the automated Playwright smoke
test passes.

## Desktop Browsers

- Resume starts from server progress, not a stale browser-local position.
- Multi-file playback transitions from one file to the next without stopping.
- Scrubbing across file boundaries selects the correct file and local time.
- Chapter clicks jump to the right whole-book position.
- Mini-player persists while navigating away from and back to the detail page.
- Media Session play, pause, and seek actions work where the browser exposes
  them.

## Mobile Browsers

- Full player controls fit without overlap at phone width.
- Mini-player does not block primary navigation or lower page actions.
- Lock-screen or notification controls appear where supported.
- Background playback behavior is noted for iOS Safari, Android Chrome, and
  installed PWA mode if available.

## Offline And Audio Processing

- Download completes for every file in a multi-file book.
- Downloaded source plays after disconnecting network access.
- Delete download removes the downloaded source and returns to streaming.
- Voice boost changes volume without clipping badly on normal narration.
- Silence trimming does not make normal speech unnaturally fast.
- EQ presets are audible and persist per book.

## Multi-Device And Sync

- A second active session appears in the detail page.
- Takeover closes the other session.
- Progress written by web appears in Audiobookshelf-compatible clients.
- Progress written by an Audiobookshelf-compatible client is used by web resume.
- Finished books are not moved back to unfinished by position-only sync.
