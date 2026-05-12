import { Card } from '@/components/ui/card';
import { Smartphone, Apple } from 'lucide-react';

export default function Apps() {
  const absServerUrl = `${window.location.origin}${window.location.pathname.replace(/\/?$/, '')}`;
  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-semibold">Audiobookshelf mobile app</h2>
      <Card className="bg-surface space-y-4 p-6">
        <div className="flex items-center gap-3 text-sm">
          <Smartphone className="size-5" />
          <span>
            Listen on iOS or Android with the official Audiobookshelf app. Add this server URL in the app:
          </span>
        </div>
        <div className="bg-background overflow-auto rounded-md border p-3 font-mono text-xs">
          {absServerUrl}
        </div>
        <ol className="text-muted-foreground list-decimal space-y-1 pl-5 text-sm">
          <li>Install Audiobookshelf from your app store.</li>
          <li>Tap "Add a server", paste the URL above.</li>
          <li>Log in with your Continuum username + password.</li>
          <li>Your audiobooks library will sync automatically.</li>
        </ol>
        <div className="flex gap-4 pt-2 text-sm">
          <a
            href="https://apps.apple.com/us/app/audiobookshelf/id1610126326"
            target="_blank"
            rel="noreferrer"
            className="text-primary inline-flex items-center gap-1 underline"
          >
            <Apple className="size-4" /> iOS App Store
          </a>
          <a
            href="https://play.google.com/store/apps/details?id=com.audiobookshelf.app"
            target="_blank"
            rel="noreferrer"
            className="text-primary inline-flex items-center gap-1 underline"
          >
            Google Play
          </a>
        </div>
      </Card>
    </div>
  );
}
