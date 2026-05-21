import { useEffect, useRef } from 'react';

type Props = {
  hasNextPage: boolean | undefined;
  isFetchingNextPage: boolean;
  fetchNextPage: () => void;
  label: string;
};

// InfiniteFooter combines an IntersectionObserver sentinel (auto-fetches when
// scrolled near the end) with a visible button (manual fallback for keyboard
// users and anyone whose IO is disabled — e.g. some accessibility tools).
// rootMargin: 600px starts fetching about a screen ahead of the user so they
// rarely see the "Loading…" state.
export default function InfiniteFooter({
  hasNextPage,
  isFetchingNextPage,
  fetchNextPage,
  label,
}: Props) {
  const sentinelRef = useRef<HTMLDivElement>(null);
  // Stash the latest fetchNextPage + isFetchingNextPage in refs so the
  // IntersectionObserver effect doesn't tear itself down on every render
  // when the parent passes fresh function identities (TanStack Query's
  // fetchNextPage isn't referentially stable). Without this, the IO was
  // rebuilt + observed on every keystroke that re-renders the parent.
  const fetchRef = useRef(fetchNextPage);
  const fetchingRef = useRef(isFetchingNextPage);
  useEffect(() => {
    fetchRef.current = fetchNextPage;
    fetchingRef.current = isFetchingNextPage;
  }, [fetchNextPage, isFetchingNextPage]);

  useEffect(() => {
    if (!hasNextPage) return;
    const node = sentinelRef.current;
    if (!node) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((e) => e.isIntersecting) && !fetchingRef.current) {
          fetchRef.current();
        }
      },
      { rootMargin: '600px' },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [hasNextPage]);

  if (!hasNextPage) return null;
  return (
    <div ref={sentinelRef} className="flex justify-center pt-4">
      <button
        type="button"
        onClick={() => fetchNextPage()}
        disabled={isFetchingNextPage}
        className="rounded-md border border-border bg-surface px-4 py-2 text-sm font-medium hover:bg-surface-hover disabled:opacity-50"
      >
        {isFetchingNextPage ? 'Loading…' : `Load more ${label}`}
      </button>
    </div>
  );
}
