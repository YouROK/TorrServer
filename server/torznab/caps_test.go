package torznab

import (
	"testing"
)

const sampleCapsXML = `<?xml version="1.0" encoding="UTF-8"?>
<caps>
  <server version="1.0" title="Test Indexer"/>
  <limits max="100" default="50"/>
  <categories>
    <category id="2000" name="Movies">
      <subcat id="2030" name="Movies/Foreign"/>
      <subcat id="2040" name="Movies/HD"/>
    </category>
    <category id="5000" name="TV">
      <subcat id="5040" name="TV/HD"/>
    </category>
  </categories>
</caps>`

func TestParseCapsBytes(t *testing.T) {
	caps, err := ParseCapsBytes([]byte(sampleCapsXML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if caps.Limits.Max != 100 {
		t.Fatalf("limits.max = %d, want 100", caps.Limits.Max)
	}
	if caps.Limits.Default != 50 {
		t.Fatalf("limits.default = %d, want 50", caps.Limits.Default)
	}
	if len(caps.Categories) != 2 {
		t.Fatalf("len(categories) = %d, want 2", len(caps.Categories))
	}

	movies := caps.Categories[0]
	if movies.ID != "2000" || movies.Name != "Movies" {
		t.Fatalf("movies category = %+v, want id=2000 name=Movies", movies)
	}
	if len(movies.Subcats) != 2 {
		t.Fatalf("len(movies subcats) = %d, want 2", len(movies.Subcats))
	}
	if movies.Subcats[0].ID != "2030" || movies.Subcats[0].Name != "Movies/Foreign" {
		t.Fatalf("movies subcat[0] = %+v", movies.Subcats[0])
	}

	tv := caps.Categories[1]
	if tv.ID != "5000" || tv.Name != "TV" {
		t.Fatalf("tv category = %+v, want id=5000 name=TV", tv)
	}
	if len(tv.Subcats) != 1 || tv.Subcats[0].ID != "5040" {
		t.Fatalf("tv subcats = %+v", tv.Subcats)
	}
}
