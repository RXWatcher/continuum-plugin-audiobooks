package abs

import "testing"

// TestCollapseBySeries_PreservesNonSeriesItems verifies items with no
// series ref pass through unchanged. The flat fields (id, title) and
// CollapsedSeries == nil are the markers ABS clients use to decide
// whether to render a single-book or series-shelf entry.
func TestCollapseBySeries_PreservesNonSeriesItems(t *testing.T) {
	items := []LibraryItem{
		{ID: "li_1", Media: LibraryItemMedia{Metadata: Metadata{Title: "A"}}},
		{ID: "li_2", Media: LibraryItemMedia{Metadata: Metadata{Title: "B"}}},
	}
	got := CollapseBySeries(items)
	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
	for _, it := range got {
		if it.CollapsedSeries != nil {
			t.Errorf("non-series item should not carry CollapsedSeries: %+v", it)
		}
	}
}

// TestCollapseBySeries_FoldsSeriesEntries folds every book in a series
// into a single representative item with NumBooks + libraryItemIds. The
// representative is the first book seen (input-order stable).
func TestCollapseBySeries_FoldsSeriesEntries(t *testing.T) {
	series := SeriesObj{ID: "s-storm", Name: "The Stormlight Archive"}
	items := []LibraryItem{
		{
			ID: "li_a",
			Media: LibraryItemMedia{Metadata: Metadata{
				Title:  "Way of Kings",
				Series: []SeriesObj{series},
			}},
		},
		{
			ID: "li_b",
			Media: LibraryItemMedia{Metadata: Metadata{
				Title:  "Words of Radiance",
				Series: []SeriesObj{series},
			}},
		},
		{
			ID: "li_c",
			Media: LibraryItemMedia{Metadata: Metadata{
				Title:  "Oathbringer",
				Series: []SeriesObj{series},
			}},
		},
	}
	got := CollapseBySeries(items)
	if len(got) != 1 {
		t.Fatalf("got %d items, want 1 (collapsed)", len(got))
	}
	rep := got[0]
	if rep.ID != "li_a" {
		t.Errorf("representative id = %q, want li_a (first-in-order)", rep.ID)
	}
	if rep.CollapsedSeries == nil {
		t.Fatal("representative must carry CollapsedSeries")
	}
	cs := rep.CollapsedSeries
	if cs.ID != "s-storm" {
		t.Errorf("series id = %q", cs.ID)
	}
	if cs.NumBooks != 3 {
		t.Errorf("NumBooks = %d, want 3", cs.NumBooks)
	}
	if len(cs.LibraryItemIDs) != 3 {
		t.Errorf("LibraryItemIDs = %v, want 3", cs.LibraryItemIDs)
	}
	if cs.NameIgnorePrefix != "Stormlight Archive" {
		t.Errorf("NameIgnorePrefix = %q, want %q (article stripped)", cs.NameIgnorePrefix, "Stormlight Archive")
	}
}

// TestCollapseBySeries_KeyFallsBackToName confirms a backend that emits
// series without an ID (legacy data) still collapses correctly by name.
// Two books with series.Name == "X" but no ID should fold into one
// representative.
func TestCollapseBySeries_KeyFallsBackToName(t *testing.T) {
	items := []LibraryItem{
		{ID: "li_a", Media: LibraryItemMedia{Metadata: Metadata{
			Series: []SeriesObj{{Name: "Old Skool"}},
		}}},
		{ID: "li_b", Media: LibraryItemMedia{Metadata: Metadata{
			Series: []SeriesObj{{Name: "Old Skool"}},
		}}},
	}
	got := CollapseBySeries(items)
	if len(got) != 1 || got[0].CollapsedSeries.NumBooks != 2 {
		t.Fatalf("got %d items / numbooks=%d, want 1/2", len(got), got[0].CollapsedSeries.NumBooks)
	}
}

// TestStripLeadingArticle pins the sort-label semantics — these
// transformations match what real ABS does so client-side sort by
// `series.nameIgnorePrefix` agrees with what stock ABS clients expect.
func TestStripLeadingArticle(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"The Stormlight Archive", "Stormlight Archive"},
		{"A Memory of Light", "Memory of Light"},
		{"An Unwanted Quest", "Unwanted Quest"},
		{"Stormlight Archive", "Stormlight Archive"},
		{"Thunder", "Thunder"}, // not "the"
	}
	for _, c := range cases {
		if got := stripLeadingArticle(c.in); got != c.want {
			t.Errorf("stripLeadingArticle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
