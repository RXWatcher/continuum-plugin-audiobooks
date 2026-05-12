import type { AudiobookSummary } from '@/api/types';
import AudiobookCard from './AudiobookCard';
import { Skeleton } from '@/components/ui/skeleton';

export default function AudiobookGrid({
  items,
  loading,
  empty,
}: {
  items: AudiobookSummary[];
  loading?: boolean;
  empty?: React.ReactNode;
}) {
  if (loading) {
    return (
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
        {Array.from({ length: 12 }).map((_, i) => (
          <Skeleton key={i} className="aspect-[2/3] w-full" />
        ))}
      </div>
    );
  }
  if (!items.length) return <div className="text-muted-foreground py-12 text-sm">{empty}</div>;
  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
      {items.map((b) => (
        <AudiobookCard key={b.id} book={b} />
      ))}
    </div>
  );
}
