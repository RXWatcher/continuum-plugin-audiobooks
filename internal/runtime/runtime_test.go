package runtime

import (
	"bytes"
	"context"
	"encoding/base64"
	"log/slog"
	"strings"
	"testing"

	pluginv1 "github.com/ContinuumApp/continuum-plugin-sdk/pkg/pluginproto/continuum/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func cfgReq(t *testing.T, kv map[string]any) *pluginv1.ConfigureRequest {
	t.Helper()
	entries := make([]*pluginv1.ConfigEntry, 0, len(kv))
	for k, v := range kv {
		s, err := structpb.NewStruct(map[string]any{"value": v})
		if err != nil {
			t.Fatalf("structpb: %v", err)
		}
		entries = append(entries, &pluginv1.ConfigEntry{Key: k, Value: s})
	}
	return &pluginv1.ConfigureRequest{Config: entries}
}

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

func TestConfigure_ValidatesCDNPair(t *testing.T) {
	s := New(nil, func(Config) error { return nil })
	_, err := s.Configure(context.Background(), cfgReq(t, map[string]any{
		"database_url": "postgres://x",
		"cdn_hostname": "cdn.example.com",
	}))
	if err == nil {
		t.Fatal("expected missing cdn_signing_secret error")
	}
}

func TestConfigure_RejectsInvalidCDNHostname(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	s := New(nil, func(Config) error { return nil })
	_, err := s.Configure(context.Background(), cfgReq(t, map[string]any{
		"database_url":       "postgres://x",
		"cdn_hostname":       "https://cdn.example.com/path",
		"cdn_signing_secret": secret,
	}))
	if err == nil {
		t.Fatal("expected invalid cdn_hostname error")
	}
}

func TestDecodeCDNSigningSecret_Requires32Bytes(t *testing.T) {
	raw := base64.StdEncoding.EncodeToString([]byte("short"))
	if _, err := DecodeCDNSigningSecret(raw); err == nil {
		t.Fatal("expected length error")
	}
}
