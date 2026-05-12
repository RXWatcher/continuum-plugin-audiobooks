import { Link } from 'react-router';
import { Headphones } from 'lucide-react';
import type { AudiobookSummary } from '@/api/types';
import { Card } from '@/components/ui/card';

export default function AudiobookCard({ book }: { book: AudiobookSummary }) {
  return (
    <Link to={`/audiobook/${encodeURIComponent(book.id)}`} className="group block">
      <Card className="bg-surface hover:bg-surface-hover overflow-hidden border-0 p-0 transition-colors">
        <div className="bg-muted aspect-[2/3] w-full overflow-hidden">
          {book.cover_url ? (
            <img
              src={book.cover_url}
              alt={book.title}
              loading="lazy"
              className="size-full object-cover transition-transform group-hover:scale-105"
            />
          ) : (
            <div className="text-muted-foreground flex size-full items-center justify-center">
              <Headphones className="size-10" />
            </div>
          )}
        </div>
        <div className="p-3">
          <div className="line-clamp-2 text-sm font-medium leading-snug">{book.title}</div>
          {book.authors && book.authors.length > 0 && (
            <div className="text-muted-foreground mt-1 line-clamp-1 text-xs">
              {book.authors.join(', ')}
            </div>
          )}
        </div>
      </Card>
    </Link>
  );
}
