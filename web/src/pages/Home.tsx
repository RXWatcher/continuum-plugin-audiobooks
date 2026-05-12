import { useQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router';
import { api } from '@/api/client';
import AudiobookGrid from '@/components/AudiobookGrid';
import SearchBar from '@/components/SearchBar';

export default function Home() {
  const [params] = useSearchParams();
  const q = params.get('q') ?? '';

  if (q) {
    return <SearchResults q={q} />;
  }
  return <Shelves />;
}

function Shelves() {
  const progress = useQuery({
    queryKey: ['progress', 'recent'],
    queryFn: () => api.listMyProgress(12),
  });

  const recent = useQuery({
    queryKey: ['audiobooks', 'recent'],
    queryFn: () => api.listAudiobooks({ sort: 'added', order: 'desc', limit: 24 }),
  });

  return (
    <div className="space-y-10">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-2xl font-semibold tracking-tight">Audiobooks</h2>
        <div className="w-full max-w-md">
          <SearchBar />
        </div>
      </div>

      {progress.data?.items && progress.data.items.length > 0 && (
        <section>
          <h3 className="mb-3 text-sm font-medium uppercase tracking-wide text-muted-foreground">
            Continue listening
          </h3>
          <div className="text-muted-foreground bg-surface rounded-lg border p-4 text-sm">
            {progress.data.items.length} book{progress.data.items.length === 1 ? '' : 's'} in
            progress.
          </div>
        </section>
      )}

      <section>
        <h3 className="mb-3 text-sm font-medium uppercase tracking-wide text-muted-foreground">
          Recently added
        </h3>
        <AudiobookGrid
          items={recent.data?.items ?? []}
          loading={recent.isLoading}
          empty="No audiobooks yet. Ask an admin to configure a backend in /admin/settings."
        />
      </section>
    </div>
  );
}

function SearchResults({ q }: { q: string }) {
  const results = useQuery({
    queryKey: ['audiobooks', 'search', q],
    queryFn: () => api.searchAudiobooks(q),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <h2 className="text-xl font-semibold tracking-tight">Results for "{q}"</h2>
        <div className="w-full max-w-md">
          <SearchBar />
        </div>
      </div>
      <AudiobookGrid
        items={results.data?.items ?? []}
        loading={results.isLoading}
        empty={`No results for "${q}". Don't see it? Try requesting it from any book detail page.`}
      />
    </div>
  );
}
