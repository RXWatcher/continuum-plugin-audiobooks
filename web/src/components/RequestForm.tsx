import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { api } from '@/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function RequestForm({ defaultTitle = '' }: { defaultTitle?: string }) {
  const qc = useQueryClient();
  const [title, setTitle] = useState(defaultTitle);
  const [author, setAuthor] = useState('');
  const [isbn, setIsbn] = useState('');

  const m = useMutation({
    mutationFn: () => api.createRequest({ title, author: author || undefined, isbn: isbn || undefined }),
    onSuccess: (out) => {
      toast.success(`Requested. Status: ${out.status}`);
      setTitle('');
      setAuthor('');
      setIsbn('');
      qc.invalidateQueries({ queryKey: ['my-requests'] });
    },
    onError: (err) => toast.error(`Failed: ${err}`),
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!title.trim()) {
          toast.error('Title required');
          return;
        }
        m.mutate();
      }}
      className="bg-surface space-y-3 rounded-lg border p-4"
    >
      <h3 className="text-sm font-medium">Request an audiobook</h3>
      <Input placeholder="Title (required)" value={title} onChange={(e) => setTitle(e.target.value)} />
      <Input placeholder="Author (optional)" value={author} onChange={(e) => setAuthor(e.target.value)} />
      <Input placeholder="ISBN (optional)" value={isbn} onChange={(e) => setIsbn(e.target.value)} />
      <Button type="submit" disabled={m.isPending}>
        {m.isPending ? 'Submitting...' : 'Submit request'}
      </Button>
    </form>
  );
}
