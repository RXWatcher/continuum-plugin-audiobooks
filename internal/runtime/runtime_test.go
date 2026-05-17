package runtime

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// The CDN signing secret and the DSN (which embeds the DB password) must
// never appear when a Config is logged or formatted.
func TestConfigRedaction(t *testing.T) {
	cfg := Config{
		DatabaseURL:          "postgres://user:sup3rsecret@db:5432/continuum",
		StandaloneHTTPListen: ":8080",
		CDNHostname:          "cdn.example.com",
		CDNSigningSecret:     "TOPSECRETSIGNINGKEY",
	}

	if s := cfg.String(); strings.Contains(s, "TOPSECRETSIGNINGKEY") || strings.Contains(s, "sup3rsecret") {
		t.Fatalf("String() leaked a secret: %s", s)
	}

	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("cfg", "config", cfg)
	out := buf.String()
	if strings.Contains(out, "TOPSECRETSIGNINGKEY") || strings.Contains(out, "sup3rsecret") {
		t.Fatalf("slog leaked a secret: %s", out)
	}
	// Non-secret fields should still be visible for debugging.
	if !strings.Contains(out, "cdn.example.com") {
		t.Fatalf("redaction also hid non-secret fields: %s", out)
	}

	// An empty secret stays empty (no spurious marker).
	if (Config{}).LogValue().String() == "" {
		t.Fatal("LogValue should still render group for an empty config")
	}
}
