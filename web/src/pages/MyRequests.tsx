import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { api } from '@/api/client';
import { Button } from '@/components/ui/button';
import RequestForm from '@/components/RequestForm';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

const STATUS_LABELS: Record<string, string> = {
  pending: 'Awaiting admin review',
  submitted: 'Submitted',
  acknowledged: 'Acknowledged',
  queued: 'Queued',
  downloading: 'Downloading',
  imported: 'Fulfilled',
  failed: 'Failed',
  denied: 'Denied',
  cancelled: 'Cancelled',
};

export default function MyRequests() {
  const qc = useQueryClient();
  const q = useQuery({ queryKey: ['my-requests'], queryFn: () => api.listMyRequests() });

  const cancel = useMutation({
    mutationFn: (id: string) => api.cancelRequest(id),
    onSuccess: () => {
      toast.success('Cancelled');
      qc.invalidateQueries({ queryKey: ['my-requests'] });
    },
    onError: (e) => toast.error(`${e}`),
  });

  return (
    <div className="space-y-8">
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <h2 className="mb-3 text-2xl font-semibold">My requests</h2>
          <div className="bg-surface overflow-hidden rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Title</TableHead>
                  <TableHead>Author</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-24" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {(q.data?.items ?? []).map((r) => (
                  <TableRow key={r.id}>
                    <TableCell>{r.title}</TableCell>
                    <TableCell className="text-muted-foreground">{r.author || ''}</TableCell>
                    <TableCell>
                      <span className="bg-background rounded border px-2 py-0.5 text-xs">
                        {STATUS_LABELS[r.status] ?? r.status}
                      </span>
                    </TableCell>
                    <TableCell>
                      {['pending', 'submitted', 'acknowledged', 'queued', 'downloading'].includes(r.status) && (
                        <Button size="sm" variant="ghost" onClick={() => cancel.mutate(r.id)}>
                          Cancel
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
                {q.data && q.data.items.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={4} className="text-muted-foreground text-center py-8">
                      No requests yet.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </div>
        <div>
          <RequestForm />
        </div>
      </div>
    </div>
  );
}
