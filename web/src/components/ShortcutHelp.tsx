import { createContext, ReactNode, useCallback, useContext, useEffect, useState } from 'react';
import { Keyboard, X } from 'lucide-react';

// Keyboard-shortcut help overlay. Opens on `?` (without modifier),
// Escape to close. Lists every global hotkey + per-page hotkey we
// surface, plus mobile gesture equivalents so touch users know what
// they can do without docs.

type Ctx = { open: () => void; close: () => void; isOpen: boolean };
const Ctx = createContext<Ctx | null>(null);

export function useShortcutHelp() {
  const c = useContext(Ctx);
  if (!c) throw new Error('useShortcutHelp must be used within ShortcutHelpProvider');
  return c;
}

export function ShortcutHelpProvider({ children }: { children: ReactNode }) {
  const [isOpen, setOpen] = useState(false);
  const open = useCallback(() => setOpen(true), []);
  const close = useCallback(() => setOpen(false), []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      // `?` triggers without modifiers and not while typing in an input.
      const target = e.target as HTMLElement | null;
      const inField =
        target && ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) ||
        (target?.isContentEditable ?? false);
      if (!inField && e.key === '?' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        setOpen((v) => !v);
      } else if (e.key === 'Escape') {
        setOpen(false);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  return (
    <Ctx.Provider value={{ open, close, isOpen }}>
      {children}
      {isOpen && <HelpOverlay close={close} />}
    </Ctx.Provider>
  );
}

const SHORTCUTS: { group: string; rows: { keys: string[]; label: string; mobile?: string }[] }[] = [
  {
    group: 'Global',
    rows: [
      { keys: ['⌘K', 'Ctrl-K'], label: 'Open command palette' },
      { keys: ['?'], label: 'Show this help', mobile: 'Three-finger tap' },
      { keys: ['/'], label: 'Focus search' },
      { keys: ['Esc'], label: 'Close overlays' },
    ],
  },
  {
    group: 'Player',
    rows: [
      { keys: ['Space'], label: 'Play / pause', mobile: 'Single tap centre' },
      { keys: ['→'], label: 'Skip forward', mobile: 'Swipe left' },
      { keys: ['←'], label: 'Skip backward', mobile: 'Swipe right' },
      { keys: ['↑'], label: 'Volume up' },
      { keys: ['↓'], label: 'Volume down' },
      { keys: ['B'], label: 'Add bookmark at current position' },
      { keys: ['M'], label: 'Mute / unmute' },
      { keys: [','], label: 'Slower playback (-0.1×)' },
      { keys: ['.'], label: 'Faster playback (+0.1×)' },
      { keys: ['S'], label: 'Sleep timer' },
    ],
  },
  {
    group: 'Library',
    rows: [
      { keys: ['G L'], label: 'Go to Library' },
      { keys: ['G A'], label: 'Go to Authors' },
      { keys: ['G S'], label: 'Go to Series' },
      { keys: ['G P'], label: 'Go to Podcasts' },
      { keys: ['G H'], label: 'Go to Home' },
    ],
  },
];

function HelpOverlay({ close }: { close: () => void }) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/60 backdrop-blur-sm pt-[8vh]"
      onClick={close}
    >
      <div
        className="bg-surface border-border w-full max-w-2xl overflow-hidden rounded-xl border shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="border-border flex items-center justify-between border-b px-4 py-3">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Keyboard className="size-4" /> Keyboard shortcuts
          </h2>
          <button
            type="button"
            onClick={close}
            className="hover:bg-surface-hover rounded p-1"
            aria-label="Close help"
          >
            <X className="size-4" />
          </button>
        </div>
        <div className="max-h-[70vh] overflow-y-auto p-4">
          {SHORTCUTS.map((g) => (
            <section key={g.group} className="mb-4 last:mb-0">
              <h3 className="text-muted-foreground mb-2 text-xs font-medium uppercase tracking-wide">
                {g.group}
              </h3>
              <div className="space-y-1">
                {g.rows.map((row) => (
                  <div key={row.label} className="flex items-center justify-between text-sm">
                    <span className="flex-1">{row.label}</span>
                    {row.mobile && (
                      <span className="text-muted-foreground mr-3 hidden text-xs sm:inline">
                        {row.mobile}
                      </span>
                    )}
                    <span className="flex shrink-0 gap-1">
                      {row.keys.map((k) => (
                        <kbd
                          key={k}
                          className="border-border bg-background min-w-7 rounded border px-2 py-0.5 text-center text-xs tabular-nums"
                        >
                          {k}
                        </kbd>
                      ))}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
        <div className="border-border text-muted-foreground border-t px-4 py-2 text-xs">
          Press <kbd className="border-border rounded border px-1">?</kbd> to toggle this
          help. Some shortcuts are page-specific and only fire when that view is focused.
        </div>
      </div>
    </div>
  );
}
