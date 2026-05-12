import { useParams } from 'react-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Bookmark as BookmarkIcon } from 'lucide-react';
import { api } from '@/api/client';
import { Button } from '@/components/ui/button';
import AudioPlayer from '@/components/player/AudioPlayer';
import ChapterList from '@/components/ChapterList';
import BookmarkList from '@/components/BookmarkList';
import { Skeleton } from '@/components/ui/skeleton';

export default function Detail() {
  const { id = '' } = useParams();
  const qc = useQueryClient();
  const detail = useQuery({
    queryKey: ['audiobook', id],
    queryFn: () => api.getAudiobook(id),
    enabled: !!id,
  });

  const addBookmark = useMutation({
    mutationFn: (position: number) =>
      api.createBookmark(id, { position_seconds: Math.floor(position) }),
    onSuccess: () => {
      toast.success('Bookmark added');
      qc.invalidateQueries({ queryKey: ['audiobook', id] });
    },
    onError: (err) => toast.error(`Failed: ${err}`),
  });

  const removeBookmark = useMutation({
    mutationFn: (bmId: string) => api.deleteBookmark(id, bmId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['audiobook', id] }),
  });

  if (detail.isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-48 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }
  if (detail.error) {
    return <div className="text-destructive">Failed to load: {String(detail.error)}</div>;
  }
  if (!detail.data) return null;

  const a = detail.data.audiobook;
  const progress = detail.data.progress;
  const bookmarks = detail.data.bookmarks ?? [];

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-6 sm:flex-row">
        <div className="bg-muted aspect-[2/3] w-48 shrink-0 overflow-hidden rounded-lg">
          {a.cover_url ? (
            <img src={a.cover_url} alt={a.title} className="size-full object-cover" />
          ) : null}
        </div>
        <div className="flex-1 space-y-3">
          <h1 className="text-3xl font-semibold leading-tight">{a.title}</h1>
          {a.authors && (
            <div className="text-muted-foreground">{a.authors.join(', ')}</div>
          )}
          {a.narrators && a.narrators.length > 0 && (
            <div className="text-muted-foreground text-sm">
              Narrated by {a.narrators.join(', ')}
            </div>
          )}
          {a.description && <p className="text-muted-foreground text-sm">{a.description}</p>}
          <div className="text-muted-foreground flex flex-wrap gap-x-4 text-xs">
            {a.year ? <span>{a.year}</span> : null}
            {a.publisher ? <span>{a.publisher}</span> : null}
            {a.series ? (
              <span>
                {a.series}
                {a.series_position ? ` #${a.series_position}` : ''}
              </span>
            ) : null}
            {a.duration_seconds ? (
              <span>{Math.floor(a.duration_seconds / 3600)}h</span>
            ) : null}
          </div>
        </div>
      </header>

      {a.files.length > 0 && (
        <section className="space-y-2">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-medium uppercase tracking-wide text-muted-foreground">
              Player
            </h2>
            <Button
              size="sm"
              variant="outline"
              onClick={() => addBookmark.mutate(progress?.current_seconds ?? 0)}
            >
              <BookmarkIcon className="mr-2 size-4" /> Bookmark
            </Button>
          </div>
          <AudioPlayer
            bookId={a.id}
            fileIdx={a.files[0].index}
            durationSeconds={a.duration_seconds || a.files[0].duration_seconds}
            initialPosition={progress?.current_seconds}
          />
        </section>
      )}

      {a.chapters && a.chapters.length > 0 && (
        <section>
          <h2 className="mb-2 text-sm font-medium uppercase tracking-wide text-muted-foreground">
            Chapters
          </h2>
          <ChapterList chapters={a.chapters} />
        </section>
      )}

      <section>
        <h2 className="mb-2 text-sm font-medium uppercase tracking-wide text-muted-foreground">
          Bookmarks
        </h2>
        <BookmarkList bookmarks={bookmarks} onDelete={(bmId) => removeBookmark.mutate(bmId)} />
      </section>
    </div>
  );
}
