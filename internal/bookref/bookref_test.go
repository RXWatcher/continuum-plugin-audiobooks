package bookref

import "testing"

// TestEncode_AddsABSPrefix confirms outbound IDs match the real-ABS
// "li_..." item-id convention. ABS mobile clients pattern-match on the
// prefix in a few places (download paths, share intents); emitting the
// prefix means we don't surprise them.
func TestEncode_AddsABSPrefix(t *testing.T) {
	got := Encode(7, "book-42")
	if got[:3] != "li_" {
		t.Fatalf("Encode = %q, missing li_ prefix", got)
	}
}

// TestDecode_AcceptsPrefixedAndLegacy validates the backward-compatible
// strip behaviour: both "li_<libID>:<b64>" and bare "<libID>:<b64>" decode
// to the same (libID, backendID) pair. Operators with bookmarked URLs in
// the legacy form must keep working.
func TestDecode_AcceptsPrefixedAndLegacy(t *testing.T) {
	encoded := Encode(7, "book-42")
	idP, bidP, okP := Decode(encoded)
	if !okP || idP != 7 || bidP != "book-42" {
		t.Errorf("prefixed: id=%d bid=%q ok=%v", idP, bidP, okP)
	}

	// Hand-build the legacy form (no li_ prefix) to confirm it still works.
	legacy := encoded[3:] // strip "li_"
	idL, bidL, okL := Decode(legacy)
	if !okL || idL != 7 || bidL != "book-42" {
		t.Errorf("legacy: id=%d bid=%q ok=%v", idL, bidL, okL)
	}
}

// TestDecode_RawIDsWithoutLibraryPrefix — backend-native IDs (no library
// prefix at all) still come through unchanged, with libraryID 0 and
// encoded=false. Used when older backends emit IDs without library
// metadata.
func TestDecode_RawIDsWithoutLibraryPrefix(t *testing.T) {
	for _, raw := range []string{"book-42", "li_book-42"} {
		id, bid, encoded := Decode(raw)
		if id != 0 || encoded {
			t.Errorf("raw %q: id=%d encoded=%v, want 0/false", raw, id, encoded)
		}
		// The "li_" prefix should strip from the backend ID too — the
		// real-ABS convention is that "li_" is purely a presentation
		// marker, never part of the underlying id.
		if bid != "book-42" {
			t.Errorf("raw %q: bid=%q, want book-42", raw, bid)
		}
	}
}

// TestRoundTrip pins the full Encode → Decode → Encode equality so any
// future change to the encoding format breaks loudly.
func TestRoundTrip(t *testing.T) {
	cases := []struct {
		libID int64
		bid   string
	}{
		{1, "book-1"},
		{42, "complex/backend-id-with-slashes"},
		{99, "trailing-newline\n"},
	}
	for _, c := range cases {
		enc := Encode(c.libID, c.bid)
		gotID, gotBID, ok := Decode(enc)
		if !ok || gotID != c.libID || gotBID != c.bid {
			t.Errorf("round-trip (%d,%q) → %q → (%d,%q,%v)",
				c.libID, c.bid, enc, gotID, gotBID, ok)
		}
	}
}
