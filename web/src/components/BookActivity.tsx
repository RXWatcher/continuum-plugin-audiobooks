import { useQuery } from '@tanstack/react-query';
import {
  Bookmark,
  History,
  Pause,
  Play,
  Share2,
  Star,
  Volume2,
} from 'lucide-react';
import { api } from '@/api/client';
import { Card } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import type { ActivityEvent } from '@/api/types';

// Per-book activity timeline. Rendered on the book detail page;
// merges progress + bookmarks + sessions + ratings + shares into
// a single chronological feed.

const KIND_META: Record<
  string,
  { icon: React.ComponentType<{ className?: string }>; label: (p: any) => string }
> = {
  progress: {
    icon: Volume2,
    label: (p) =>
      p?.is_finished
        ? 'Finished'
        : `Updated progress (${Math.round((p?.progress_pct ?? 0) * 100)}%)`,
  },
  bookmark: {
    icon: Bookmark,
    label: (p) =>
      p?.note
        ? `Bookmark: ${truncate(p.note, 60)}`
        : `Bookmarked at ${formatHMS(p?.position ?? 0)}`,
  },
  session_opened: {
    icon: Play,
    label: (p) => `Started listening (${formatHMS(p?.start_position ?? 0)})`,
  },
  session_closed: {
    icon: Pause,
    label: (p) => `Paused at ${formatHMS(p?.end_position ?? 0)}`,
  },
  rated: {
    icon: Star,
    label: (p) => `Rated ${p?.rating ?? 0}/5`,
  },
  shared: {
    icon: Share2,
    label: () => 'Created a share link',
  },
};

export default function BookActivity({ bookId }: { bookId: string }) {
  const activity = useQuery({
    queryKey: ['book-activity', bookId],
    queryFn: () => api.getBookActivity(bookId),
    enabled: !!bookId,
  });

  return (
    <Card className="bg-surface p-4">
      <div className="mb-3 flex items-center gap-2">
        <History className="size-5" />
        <h3 className="font-medium">Activity</h3>
      </div>
      {activity.isLoading ? (
        <Skeleton className="h-24 w-full" />
      ) : (activity.data?.events ?? []).length === 0 ? (
        <p className="text-muted-foreground text-sm">
          No activity yet — your sessions, bookmarks, and progress will appear here.
        </p>
      ) : (
        <ol className="space-y-3">
          {activity.data!.events.slice(0, 50).map((ev, i) => (
            <EventRow key={`${ev.at}-${i}`} event={ev} />
          ))}
        </ol>
      )}
    </Card>
  );
}

function EventRow({ event }: { event: ActivityEvent }) {
  const meta = KIND_META[event.kind];
  const Icon = meta?.icon ?? History;
  const label = meta ? meta.label(event.payload) : event.kind;
  return (
    <li className="flex items-start gap-3">
      <div className="bg-background mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full">
        <Icon className="size-4" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm">{label}</div>
        <div className="text-muted-foreground text-xs tabular-nums">
          {formatWhen(event.at)}
        </div>
      </div>
    </li>
  );
}

function formatHMS(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return '0:00';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h)
    return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  return `${m}:${String(s).padStart(2, '0')}`;
}

function formatWhen(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const diffMs = Date.now() - d.getTime();
  const minute = 60_000;
  const hour = minute * 60;
  const day = hour * 24;
  if (diffMs < minute) return 'just now';
  if (diffMs < hour) return `${Math.floor(diffMs / minute)}m ago`;
  if (diffMs < day) return `${Math.floor(diffMs / hour)}h ago`;
  if (diffMs < day * 30) return `${Math.floor(diffMs / day)}d ago`;
  return d.toLocaleDateString();
}

function truncate(s: string, n: number): string {
  if (s.length <= n) return s;
  return s.slice(0, n - 1) + '…';
}
