package abs

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMinify_FlattenAuthorsAndDropArrays(t *testing.T) {
	item := LibraryItem{
		ID:        "li-1",
		LibraryID: "L1",
		Media: LibraryItemMedia{
			Metadata: Metadata{
				Title:   "Way of Kings",
				Authors: []AuthorObj{{ID: "a1", Name: "Brandon Sanderson"}},
				Series:  []SeriesObj{{ID: "s1", Name: "Stormlight", Sequence: "1"}},
			},
		},
	}
	mini := Minify(item)
	if mini.Media.Metadata.AuthorName != "Brandon Sanderson" {
		t.Errorf("AuthorName = %q", mini.Media.Metadata.AuthorName)
	}
	if mini.Media.Metadata.AuthorNameLF != "Sanderson, Brandon" {
		t.Errorf("AuthorNameLF = %q, want Sanderson, Brandon", mini.Media.Metadata.AuthorNameLF)
	}
	if mini.Media.Metadata.SeriesName != "Stormlight" {
		t.Errorf("SeriesName = %q", mini.Media.Metadata.SeriesName)
	}
	if mini.Media.Metadata.SeriesSequence != "1" {
		t.Errorf("SeriesSequence = %q", mini.Media.Metadata.SeriesSequence)
	}

	// JSON must not carry the heavy arrays.
	body, err := json.Marshal(mini)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	if strings.Contains(s, `"authors":[`) {
		t.Errorf("minified payload must not include authors[]: %s", s)
	}
	if strings.Contains(s, `"series":[`) {
		t.Errorf("minified payload must not include series[]: %s", s)
	}
	if !strings.Contains(s, `"authorName":`) {
		t.Errorf("minified payload must include authorName: %s", s)
	}
}

func TestMinify_MultipleAuthors(t *testing.T) {
	item := LibraryItem{
		Media: LibraryItemMedia{Metadata: Metadata{
			Authors: []AuthorObj{
				{Name: "Brandon Sanderson"},
				{Name: "Robert Jordan"},
			},
		}},
	}
	mini := Minify(item)
	if mini.Media.Metadata.AuthorName != "Brandon Sanderson, Robert Jordan" {
		t.Errorf("AuthorName = %q", mini.Media.Metadata.AuthorName)
	}
	if mini.Media.Metadata.AuthorNameLF != "Sanderson, Brandon & Jordan, Robert" {
		t.Errorf("AuthorNameLF = %q", mini.Media.Metadata.AuthorNameLF)
	}
}

func TestMinify_SingleTokenName(t *testing.T) {
	item := LibraryItem{
		Media: LibraryItemMedia{Metadata: Metadata{
			Authors: []AuthorObj{{Name: "Homer"}},
		}},
	}
	mini := Minify(item)
	if mini.Media.Metadata.AuthorNameLF != "Homer" {
		t.Errorf("single-token name AuthorNameLF = %q, want Homer", mini.Media.Metadata.AuthorNameLF)
	}
}

func TestMinify_NoSeries(t *testing.T) {
	item := LibraryItem{Media: LibraryItemMedia{Metadata: Metadata{Authors: []AuthorObj{{Name: "X"}}}}}
	mini := Minify(item)
	if mini.Media.Metadata.SeriesName != "" || mini.Media.Metadata.SeriesSequence != "" {
		t.Errorf("expected blank series; got name=%q seq=%q",
			mini.Media.Metadata.SeriesName, mini.Media.Metadata.SeriesSequence)
	}
}
