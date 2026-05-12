import { useEffect, useRef, useState } from 'react';
import { Pause, Play, SkipBack, SkipForward, Gauge } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { api } from '@/api/client';

const SPEEDS = [0.8, 1.0, 1.25, 1.5, 1.75, 2.0];

function fmt(t: number): string {
  if (!Number.isFinite(t) || t < 0) return '0:00';
  const h = Math.floor(t / 3600);
  const m = Math.floor((t % 3600) / 60);
  const s = Math.floor(t % 60);
  if (h) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  return `${m}:${String(s).padStart(2, '0')}`;
}

export default function AudioPlayer({
  bookId,
  fileIdx,
  durationSeconds,
  initialPosition,
}: {
  bookId: string;
  fileIdx: number;
  durationSeconds: number;
  initialPosition?: number;
}) {
  const ref = useRef<HTMLAudioElement | null>(null);
  const [playing, setPlaying] = useState(false);
  const [current, setCurrent] = useState(initialPosition ?? 0);
  const [speed, setSpeed] = useState(1.0);
  const lastWrite = useRef(0);

  // Resume position on first mount.
  useEffect(() => {
    if (!ref.current) return;
    if (initialPosition && initialPosition > 0) {
      ref.current.currentTime = initialPosition;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Throttled progress write (every 15s of wall clock).
  useEffect(() => {
    if (!playing) return;
    const id = setInterval(() => {
      const now = Date.now();
      if (now - lastWrite.current >= 15_000 && ref.current) {
        lastWrite.current = now;
        const ct = ref.current.currentTime;
        const pct = durationSeconds > 0 ? ct / durationSeconds : 0;
        api
          .upsertProgress(bookId, {
            current_seconds: Math.floor(ct),
            progress_pct: pct,
            is_finished: pct >= 0.95,
          })
          .catch(() => {});
      }
    }, 5_000);
    return () => clearInterval(id);
  }, [playing, bookId, durationSeconds]);

  const skip = (deltaSec: number) => {
    if (!ref.current) return;
    ref.current.currentTime = Math.max(0, ref.current.currentTime + deltaSec);
  };

  return (
    <div className="bg-surface rounded-2xl border p-4">
      <audio
        ref={ref}
        src={api.streamUrl(bookId, fileIdx)}
        preload="metadata"
        onTimeUpdate={(e) => setCurrent((e.target as HTMLAudioElement).currentTime)}
        onPlay={() => setPlaying(true)}
        onPause={() => setPlaying(false)}
        onEnded={() => setPlaying(false)}
      />
      <div className="flex items-center gap-4">
        <Button size="icon" variant="ghost" onClick={() => skip(-30)} aria-label="Back 30s">
          <SkipBack className="size-5" />
        </Button>
        <Button
          size="icon"
          onClick={() => {
            if (!ref.current) return;
            if (ref.current.paused) ref.current.play();
            else ref.current.pause();
          }}
          aria-label={playing ? 'Pause' : 'Play'}
          className="size-12 rounded-full"
        >
          {playing ? <Pause className="size-6" /> : <Play className="size-6" />}
        </Button>
        <Button size="icon" variant="ghost" onClick={() => skip(30)} aria-label="Forward 30s">
          <SkipForward className="size-5" />
        </Button>
        <div className="text-muted-foreground tabular-nums w-24 text-xs">
          {fmt(current)} / {fmt(durationSeconds)}
        </div>
        <div className="flex-1">
          <input
            type="range"
            min={0}
            max={durationSeconds}
            value={current}
            onChange={(e) => {
              const v = Number(e.target.value);
              setCurrent(v);
              if (ref.current) ref.current.currentTime = v;
            }}
            className="w-full"
          />
        </div>
        <div className="flex items-center gap-1">
          <Gauge className="text-muted-foreground size-4" />
          <select
            value={speed}
            onChange={(e) => {
              const v = Number(e.target.value);
              setSpeed(v);
              if (ref.current) ref.current.playbackRate = v;
            }}
            className="bg-background rounded border px-2 py-1 text-xs"
          >
            {SPEEDS.map((s) => (
              <option key={s} value={s}>
                {s}x
              </option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
}
