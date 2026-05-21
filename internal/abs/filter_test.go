package abs

import (
	"encoding/base64"
	"testing"
)

func TestParseFilter_Base64Authors(t *testing.T) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte("Brandon Sanderson"))
	f, ok := ParseFilter("authors." + encoded)
	if !ok {
		t.Fatalf("ParseFilter returned !ok")
	}
	if f.Kind != FilterAuthors {
		t.Errorf("Kind = %q, want authors", f.Kind)
	}
	if f.Value != "Brandon Sanderson" {
		t.Errorf("Value = %q, want decoded name", f.Value)
	}
}

func TestParseFilter_NoSeriesSentinel(t *testing.T) {
	f, ok := ParseFilter("series.no-series")
	if !ok || f.Value != SentinelNoSeries {
		t.Fatalf("no-series sentinel must pass through; got value=%q ok=%v", f.Value, ok)
	}
}

func TestParseFilter_ProgressUnencoded(t *testing.T) {
	for _, sub := range []string{"in-progress", "finished", "not-finished"} {
		f, ok := ParseFilter("progress." + sub)
		if !ok || f.Kind != FilterProgress || f.Value != sub {
			t.Errorf("progress.%s parsed wrong: kind=%s value=%q ok=%v",
				sub, f.Kind, f.Value, ok)
		}
	}
}

func TestParseFilter_StdBase64Fallback(t *testing.T) {
	// Some ABS clients use Std (padded) base64. We should accept both.
	encoded := base64.StdEncoding.EncodeToString([]byte("sci-fi"))
	f, ok := ParseFilter("genres." + encoded)
	if !ok || f.Value != "sci-fi" {
		t.Errorf("std-base64 not accepted: value=%q ok=%v", f.Value, ok)
	}
}

func TestParseFilter_EmptyAndMalformed(t *testing.T) {
	if _, ok := ParseFilter(""); ok {
		t.Error("empty input must not parse")
	}
	if _, ok := ParseFilter("no-dot"); ok {
		t.Error("missing dot must not parse")
	}
	if _, ok := ParseFilter("authors."); ok {
		t.Error("empty value must not parse")
	}
}

func TestFilter_MatchesAuthors(t *testing.T) {
	item := LibraryItem{
		Media: LibraryItemMedia{
			Metadata: Metadata{
				Authors: []AuthorObj{{ID: "a1", Name: "Brandon Sanderson"}},
			},
		},
	}
	if !(Filter{Kind: FilterAuthors, Value: "a1"}).Matches(item, false, false, false) {
		t.Error("must match by author id")
	}
	if !(Filter{Kind: FilterAuthors, Value: "Brandon Sanderson"}).Matches(item, false, false, false) {
		t.Error("must match by author name")
	}
	if (Filter{Kind: FilterAuthors, Value: "Other"}).Matches(item, false, false, false) {
		t.Error("must not match unrelated author")
	}
}

func TestFilter_MatchesNoSeries(t *testing.T) {
	withSeries := LibraryItem{Media: LibraryItemMedia{Metadata: Metadata{Series: []SeriesObj{{ID: "s1", Name: "Stormlight"}}}}}
	withoutSeries := LibraryItem{Media: LibraryItemMedia{Metadata: Metadata{Series: []SeriesObj{}}}}
	noSeriesFilter := Filter{Kind: FilterSeries, Value: SentinelNoSeries}
	if noSeriesFilter.Matches(withSeries, false, false, false) {
		t.Error("no-series must NOT match an item with a series")
	}
	if !noSeriesFilter.Matches(withoutSeries, false, false, false) {
		t.Error("no-series MUST match an item with no series")
	}
}

func TestFilter_MatchesProgress(t *testing.T) {
	item := LibraryItem{}
	cases := []struct {
		name        string
		value       string
		inProgress  bool
		finished    bool
		hasProgress bool
		want        bool
	}{
		{"in-progress with progress row", "in-progress", true, false, true, true},
		{"in-progress without progress row", "in-progress", false, false, false, false},
		{"finished with progress row", "finished", false, true, true, true},
		{"not-finished match", "not-finished", false, false, true, true},
		{"not-finished does not match finished", "not-finished", false, true, true, false},
		{"not-started has no row", "not-started", false, false, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := Filter{Kind: FilterProgress, Value: tc.value}
			got := f.Matches(item, tc.inProgress, tc.finished, tc.hasProgress)
			if got != tc.want {
				t.Errorf("Matches(%s) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}
