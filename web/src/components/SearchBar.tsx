import { Search } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router';

export default function SearchBar() {
  const navigate = useNavigate();
  const [q, setQ] = useState('');
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        const trimmed = q.trim();
        if (!trimmed) return;
        navigate(`/?q=${encodeURIComponent(trimmed)}`);
      }}
      className="bg-surface focus-within:ring-ring/30 flex items-center gap-2 rounded-lg border px-3 py-1.5 focus-within:ring-2"
    >
      <Search className="text-muted-foreground size-4" />
      <input
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder="Search audiobooks..."
        className="placeholder:text-muted-foreground w-full bg-transparent text-sm outline-none"
      />
    </form>
  );
}
