import { useEffect, useRef, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { api } from '@/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { BackendConfig } from '@/api/types';

type SecretState =
  | { status: 'idle' }
  | { status: 'generating' }
  | { status: 'ready'; secret: string }
  | { status: 'error'; message: string };

const MODES = ['proxy', 'cache', 'direct'] as const;

export default function AdminSettings() {
  const qc = useQueryClient();
  const q = useQuery({ queryKey: ['backend-config'], queryFn: () => api.getBackendConfig() });
  const [form, setForm] = useState<BackendConfig | null>(null);
  const [secretState, setSecretState] = useState<SecretState>({ status: 'idle' });
  const [copied, setCopied] = useState(false);
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (q.data && !form) setForm(q.data);
  }, [q.data, form]);

  const save = useMutation({
    mutationFn: (body: Partial<BackendConfig> & { rotate_abs_secret?: boolean }) =>
      api.updateBackendConfig(body),
    onSuccess: () => {
      toast.success('Saved');
      qc.invalidateQueries({ queryKey: ['backend-config'] });
    },
    onError: (e) => toast.error(`${e}`),
  });

  if (!form) return <div className="text-muted-foreground">Loading...</div>;

  const update = <K extends keyof BackendConfig>(k: K, v: BackendConfig[K]) =>
    setForm({ ...form, [k]: v });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate(form);
      }}
      className="space-y-6"
    >
      <h2 className="text-2xl font-semibold">Admin: Settings</h2>

      <section className="bg-surface space-y-4 rounded-lg border p-6">
        <h3 className="text-sm font-medium uppercase tracking-wide text-muted-foreground">Routing</h3>
        <Field label="Backend plugin id">
          <Input
            value={form.target_backend_plugin_id}
            onChange={(e) => update('target_backend_plugin_id', e.target.value)}
            placeholder="continuum.bookwarehouse-audio"
          />
        </Field>
        <Field label="Auto-approve requests">
          <Switch
            checked={form.auto_approve_requests}
            onCheckedChange={(v) => update('auto_approve_requests', v)}
          />
        </Field>
      </section>

      <section className="bg-surface space-y-4 rounded-lg border p-6">
        <h3 className="text-sm font-medium uppercase tracking-wide text-muted-foreground">Streaming</h3>
        <Field label="Mode">
          <Select
            value={form.streaming_mode}
            onValueChange={(v) => update('streaming_mode', v as BackendConfig['streaming_mode'])}
          >
            <SelectTrigger className="w-48">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {MODES.map((m) => (
                <SelectItem key={m} value={m}>
                  {m}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
        {form.streaming_mode === 'cache' && (
          <>
            <Field label="Cache directory">
              <Input
                value={form.cache_dir}
                onChange={(e) => update('cache_dir', e.target.value)}
                placeholder="/var/lib/continuum/audiobooks-cache"
              />
            </Field>
            <Field label="Max size (GB)">
              <Input
                type="number"
                value={form.cache_max_size_gb}
                onChange={(e) => update('cache_max_size_gb', Number(e.target.value))}
              />
            </Field>
            <Field label="Download concurrency">
              <Input
                type="number"
                value={form.cache_download_concurrency}
                onChange={(e) => update('cache_download_concurrency', Number(e.target.value))}
              />
            </Field>
          </>
        )}
      </section>

      <section className="bg-surface space-y-4 rounded-lg border p-6">
        <h3 className="text-sm font-medium uppercase tracking-wide text-muted-foreground">ABS API</h3>
        <Field label="Access token TTL (hours)">
          <Input
            type="number"
            value={form.abs_access_token_ttl_hours}
            onChange={(e) => update('abs_access_token_ttl_hours', Number(e.target.value))}
          />
        </Field>
        <Field label="Refresh token TTL (days)">
          <Input
            type="number"
            value={form.abs_refresh_token_ttl_days}
            onChange={(e) => update('abs_refresh_token_ttl_days', Number(e.target.value))}
          />
        </Field>
        <Button
          type="button"
          variant="outline"
          onClick={() => save.mutate({ rotate_abs_secret: true })}
          disabled={save.isPending}
        >
          Rotate ABS signing secret
        </Button>
      </section>

      <section className="bg-surface space-y-4 rounded-lg border p-6">
        <h3 className="text-sm font-medium uppercase tracking-wide text-muted-foreground">
          Audiobooks CDN (presigned URL streaming)
        </h3>
        <p className="text-sm text-muted-foreground">
          Generate a shared HMAC signing secret for CDN-presigned streaming URLs. Once generated,
          paste the value into <strong>this plugin's</strong> <code>cdn_signing_secret</code> config
          field <strong>and</strong> the <strong>audiobooksdb plugin's</strong>{' '}
          <code>stream_signing_secret</code> config field. The secret is shown only once — save it
          before navigating away.
        </p>
        <div className="flex flex-wrap items-center gap-3">
          <Button
            type="button"
            variant="outline"
            disabled={secretState.status === 'generating'}
            onClick={async () => {
              setSecretState({ status: 'generating' });
              setCopied(false);
              try {
                const { secret } = await api.adminGenerateStreamingSecret();
                setSecretState({ status: 'ready', secret });
              } catch (e) {
                setSecretState({ status: 'error', message: `${e}` });
                toast.error(`Failed to generate secret: ${e}`);
              }
            }}
          >
            {secretState.status === 'generating' ? 'Generating…' : 'Generate streaming secret'}
          </Button>
        </div>
        {secretState.status === 'ready' && (
          <div className="space-y-2">
            <Label>Generated secret — copy and save now</Label>
            <div className="flex gap-2">
              <Input
                readOnly
                value={secretState.secret}
                className="font-mono text-xs"
                onFocus={(e) => e.currentTarget.select()}
              />
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  navigator.clipboard.writeText(secretState.secret).then(() => {
                    setCopied(true);
                    if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current);
                    copyTimeoutRef.current = setTimeout(() => setCopied(false), 2000);
                  });
                }}
              >
                {copied ? 'Copied' : 'Copy'}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              This value will not be shown again. Store it securely before leaving this page.
            </p>
          </div>
        )}
        {secretState.status === 'error' && (
          <p className="text-sm text-destructive">{secretState.message}</p>
        )}
      </section>

      <Button type="submit" disabled={save.isPending}>
        {save.isPending ? 'Saving...' : 'Save'}
      </Button>
    </form>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-3 items-center gap-3">
      <Label className="col-span-1">{label}</Label>
      <div className="col-span-2">{children}</div>
    </div>
  );
}
