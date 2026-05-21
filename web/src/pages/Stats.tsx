import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Flame, Headphones, Target, Trophy } from 'lucide-react';
import { api } from '@/api/client';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import type { HeatmapDay } from '@/api/types';

// Stats dashboard. Single page that bundles:
//   - Reading streak counter
//   - Yearly goals (books + hours) with progress bars + on-pace flag
//   - 90-day listening heatmap (GitHub-style grid)
//   - Year-in-review (top books / narrators / authors)

export default function Stats() {
  const year = new Date().getUTCFullYear();
  const streak = useQuery({ queryKey: ['streak'], queryFn: () => api.getStreak() });
  const goalProgress = useQuery({
    queryKey: ['goal-progress', year],
    queryFn: () => api.getGoalProgress(year),
  });
  const heatmap = useQuery({
    queryKey: ['heatmap', 90],
    queryFn: () => api.getHeatmap(90),
  });
  const yearStats = useQuery({
    queryKey: ['year-stats', year],
    queryFn: () => api.getYearStats(year),
  });

  return (
    <div className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold tracking-tight">Your stats</h2>
        <p className="text-muted-foreground text-sm">
          Listening activity, streaks, and goal progress for {year}.
        </p>
      </header>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          icon={<Flame className="size-5" />}
          label="Current streak"
          value={
            streak.isLoading ? (
              <Skeleton className="h-7 w-14" />
            ) : (
              <>{streak.data?.current ?? 0} days</>
            )
          }
          hint={
            streak.data?.longest
              ? `Longest: ${streak.data.longest} days`
              : 'Start a session today to begin'
          }
        />
        <StatCard
          icon={<Headphones className="size-5" />}
          label="Hours this year"
          value={
            yearStats.isLoading ? (
              <Skeleton className="h-7 w-14" />
            ) : (
              <>{Math.round(yearStats.data?.total_hours ?? 0)}</>
            )
          }
        />
        <StatCard
          icon={<Trophy className="size-5" />}
          label="Books finished"
          value={
            yearStats.isLoading ? (
              <Skeleton className="h-7 w-14" />
            ) : (
              <>{yearStats.data?.books_finished ?? 0}</>
            )
          }
        />
        <StatCard
          icon={<Target className="size-5" />}
          label="Active days"
          value={
            yearStats.isLoading ? (
              <Skeleton className="h-7 w-14" />
            ) : (
              <>{yearStats.data?.distinct_days ?? 0}</>
            )
          }
        />
      </div>

      <GoalsCard year={year} progress={goalProgress.data?.goals ?? []} loading={goalProgress.isLoading} />

      <HeatmapCard days={heatmap.data?.days ?? []} loading={heatmap.isLoading} />

      <TopBooksCard
        top={yearStats.data?.top_books ?? []}
        loading={yearStats.isLoading}
      />

      <TopVoicesCard
        authors={yearStats.data?.top_authors ?? []}
        narrators={yearStats.data?.top_narrators ?? []}
        loading={yearStats.isLoading}
      />
    </div>
  );
}

function StatCard({
  icon,
  label,
  value,
  hint,
}: {
  icon: React.ReactNode;
  label: string;
  value: React.ReactNode;
  hint?: string;
}) {
  return (
    <Card className="bg-surface p-4">
      <div className="text-muted-foreground flex items-center gap-2 text-xs">
        {icon}
        {label}
      </div>
      <div className="mt-2 text-2xl font-semibold tabular-nums">{value}</div>
      {hint && <div className="text-muted-foreground mt-1 text-xs">{hint}</div>}
    </Card>
  );
}

function GoalsCard({
  year,
  progress,
  loading,
}: {
  year: number;
  progress: { kind: string; target: number; actual: number; percent_complete: number; on_pace_for_target: boolean }[];
  loading: boolean;
}) {
  const qc = useQueryClient();
  const [booksTarget, setBooksTarget] = useState<number | ''>('');
  const [hoursTarget, setHoursTarget] = useState<number | ''>('');

  // Hydrate the inputs from existing goals so editing reflects the
  // current target. Reset only when the year/progress shape changes.
  useEffect(() => {
    const books = progress.find((g) => g.kind === 'books')?.target;
    const hours = progress.find((g) => g.kind === 'hours')?.target;
    setBooksTarget(books ?? '');
    setHoursTarget(hours ?? '');
  }, [year, progress.length]);

  const save = useMutation({
    mutationFn: async () => {
      if (booksTarget && Number(booksTarget) > 0) {
        await api.putGoal(year, 'books', Number(booksTarget));
      }
      if (hoursTarget && Number(hoursTarget) > 0) {
        await api.putGoal(year, 'hours', Number(hoursTarget));
      }
    },
    onSuccess: () => {
      toast.success('Goals saved');
      qc.invalidateQueries({ queryKey: ['goal-progress', year] });
    },
    onError: (err) => toast.error(`Save failed: ${err}`),
  });

  return (
    <Card className="bg-surface p-4">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h3 className="font-medium">Yearly goals</h3>
          <p className="text-muted-foreground text-xs">
            Set a target for {year}; progress updates as you listen.
          </p>
        </div>
      </div>

      {loading ? (
        <Skeleton className="h-24 w-full" />
      ) : (
        <div className="space-y-4">
          {progress.map((g) => (
            <GoalRow key={g.kind} goal={g} />
          ))}
          {progress.length === 0 && (
            <p className="text-muted-foreground text-sm">No goals set yet.</p>
          )}
        </div>
      )}

      <div className="bg-border my-4 h-px" />

      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <Label htmlFor="books-target">Books target ({year})</Label>
          <Input
            id="books-target"
            type="number"
            min={0}
            placeholder="e.g. 24"
            value={booksTarget}
            onChange={(e) => setBooksTarget(e.target.value ? Number(e.target.value) : '')}
          />
        </div>
        <div>
          <Label htmlFor="hours-target">Hours target ({year})</Label>
          <Input
            id="hours-target"
            type="number"
            min={0}
            placeholder="e.g. 120"
            value={hoursTarget}
            onChange={(e) => setHoursTarget(e.target.value ? Number(e.target.value) : '')}
          />
        </div>
      </div>
      <Button className="mt-3" onClick={() => save.mutate()} disabled={save.isPending}>
        Save goals
      </Button>
    </Card>
  );
}

function GoalRow({
  goal,
}: {
  goal: { kind: string; target: number; actual: number; percent_complete: number; on_pace_for_target: boolean };
}) {
  const pct = Math.max(0, Math.min(100, goal.percent_complete));
  return (
    <div>
      <div className="mb-1 flex items-baseline justify-between text-sm">
        <span className="font-medium capitalize">{goal.kind}</span>
        <span className="text-muted-foreground tabular-nums text-xs">
          {goal.actual} / {goal.target}
          {goal.on_pace_for_target ? ' · on pace' : ' · behind pace'}
        </span>
      </div>
      <div className="bg-muted h-2 overflow-hidden rounded-full">
        <div
          className={`h-full ${goal.on_pace_for_target ? 'bg-primary' : 'bg-amber-500'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

function HeatmapCard({ days, loading }: { days: HeatmapDay[]; loading: boolean }) {
  // GitHub-style heatmap grid: 13 weeks × 7 days = last 90 days.
  // Each cell colour-coded by seconds listened — log scale so a
  // 5-minute peek doesn't look identical to a 4-hour binge.
  const cells = useMemo(() => buildHeatmapCells(days), [days]);
  const maxSeconds = useMemo(
    () => days.reduce((m, d) => Math.max(m, d.seconds), 0),
    [days],
  );
  return (
    <Card className="bg-surface p-4">
      <h3 className="mb-3 font-medium">Last 90 days</h3>
      {loading ? (
        <Skeleton className="h-24 w-full" />
      ) : days.length === 0 ? (
        <p className="text-muted-foreground text-sm">No listening sessions yet.</p>
      ) : (
        <div className="flex gap-1">
          {cells.map((week, wi) => (
            <div key={wi} className="flex flex-col gap-1">
              {week.map((cell, di) =>
                cell ? (
                  <div
                    key={di}
                    className="size-3 rounded-sm"
                    style={{ backgroundColor: heatColor(cell.seconds, maxSeconds) }}
                    title={`${cell.date}: ${formatHM(cell.seconds)}`}
                  />
                ) : (
                  <div key={di} className="size-3" />
                ),
              )}
            </div>
          ))}
        </div>
      )}
    </Card>
  );
}

function TopBooksCard({
  top,
  loading,
}: {
  top: { book_id: string; title?: string; authors?: string[]; seconds_listened: number }[];
  loading: boolean;
}) {
  return (
    <Card className="bg-surface p-4">
      <h3 className="mb-3 font-medium">Most listened this year</h3>
      {loading ? (
        <Skeleton className="h-32 w-full" />
      ) : top.length === 0 ? (
        <p className="text-muted-foreground text-sm">No listening sessions yet.</p>
      ) : (
        <ol className="space-y-2">
          {top.slice(0, 8).map((b, i) => (
            <li key={b.book_id} className="flex items-baseline justify-between text-sm">
              <span className="truncate">
                <span className="text-muted-foreground mr-2 tabular-nums">{i + 1}.</span>
                <span className="font-medium">{b.title ?? b.book_id}</span>
                {b.authors?.length ? (
                  <span className="text-muted-foreground ml-2">
                    by {b.authors.join(', ')}
                  </span>
                ) : null}
              </span>
              <span className="text-muted-foreground tabular-nums text-xs">
                {formatHM(b.seconds_listened)}
              </span>
            </li>
          ))}
        </ol>
      )}
    </Card>
  );
}

function TopVoicesCard({
  authors,
  narrators,
  loading,
}: {
  authors: { name: string; seconds: number }[];
  narrators: { name: string; seconds: number }[];
  loading: boolean;
}) {
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Card className="bg-surface p-4">
        <h3 className="mb-3 font-medium">Top authors</h3>
        {loading ? <Skeleton className="h-24 w-full" /> : <NameList items={authors} empty="No data yet" />}
      </Card>
      <Card className="bg-surface p-4">
        <h3 className="mb-3 font-medium">Top narrators</h3>
        {loading ? <Skeleton className="h-24 w-full" /> : <NameList items={narrators} empty="No data yet" />}
      </Card>
    </div>
  );
}

function NameList({ items, empty }: { items: { name: string; seconds: number }[]; empty: string }) {
  if (!items.length) return <p className="text-muted-foreground text-sm">{empty}</p>;
  return (
    <ol className="space-y-1">
      {items.slice(0, 6).map((x, i) => (
        <li key={x.name + i} className="flex justify-between text-sm">
          <span className="truncate">{x.name}</span>
          <span className="text-muted-foreground tabular-nums text-xs">{formatHM(x.seconds)}</span>
        </li>
      ))}
    </ol>
  );
}

// buildHeatmapCells slots the day list into a 13×7 grid keyed by
// weekday so the first column starts on the correct day-of-week.
// Missing dates appear as gaps; the input is already filtered to
// non-zero days by the server.
function buildHeatmapCells(days: HeatmapDay[]): (HeatmapDay | null)[][] {
  const map = new Map<string, HeatmapDay>();
  for (const d of days) map.set(d.date, d);

  const today = new Date();
  const cells: (HeatmapDay | null)[][] = [];
  const start = new Date(today);
  start.setUTCDate(start.getUTCDate() - 89);
  // Snap to Sunday so columns are weeks.
  const startDay = start.getUTCDay();
  start.setUTCDate(start.getUTCDate() - startDay);

  const totalWeeks = 13 + (startDay > 0 ? 1 : 0);
  for (let w = 0; w < totalWeeks; w++) {
    const week: (HeatmapDay | null)[] = [];
    for (let d = 0; d < 7; d++) {
      const cur = new Date(start);
      cur.setUTCDate(start.getUTCDate() + w * 7 + d);
      const key = cur.toISOString().slice(0, 10);
      if (cur > today) {
        week.push(null);
      } else {
        week.push(map.get(key) ?? { date: key, sessions: 0, seconds: 0 });
      }
    }
    cells.push(week);
  }
  return cells;
}

function heatColor(seconds: number, max: number): string {
  if (!seconds) return 'rgba(255,255,255,0.04)';
  if (!max) return 'rgba(255,255,255,0.04)';
  // Log scale — 5 minutes vs 4 hours visibly differ but the heavy
  // days don't drown out the light ones.
  const intensity = Math.log10(1 + seconds) / Math.log10(1 + max);
  const alpha = 0.1 + intensity * 0.85;
  return `hsl(220, 80%, 55% / ${alpha})`;
}

function formatHM(seconds: number): string {
  if (!seconds) return '0m';
  const h = Math.floor(seconds / 3600);
  const m = Math.round((seconds % 3600) / 60);
  if (h && m) return `${h}h ${m}m`;
  if (h) return `${h}h`;
  return `${m}m`;
}
