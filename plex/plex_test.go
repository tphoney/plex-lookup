package plex

import (
	"os"
	"testing"
	"time"

	types "github.com/tphoney/plex-lookup/types"
)

func TestFindMovieDetails(t *testing.T) {
	rawdata, err := os.ReadFile("testdata/movies.xml")
	if err != nil {
		t.Errorf("Error reading testdata/movies.xml: %s", err)
	}

	processed := extractMovies(string(rawdata))
	expected := []types.PlexMovie{
		{
			Title:     "Chaos Theory",
			Year:      "2007",
			DateAdded: time.Date(2023, time.January, 21, 15, 03, 10, 0, time.FixedZone("GMT", 0)),
		},
	}

	if len(processed) != 3 {
		t.Errorf("Expected 3 movies, but got %d", len(processed))
	}

	if processed[0].Title != expected[0].Title {
		t.Errorf("Expected title %s, but got %s", expected[0].Title, processed[0].Title)
	}

	if processed[0].Year != expected[0].Year {
		t.Errorf("Expected year %s, but got %s", expected[0].Year, processed[0].Year)
	}

	if processed[0].DateAdded.Compare(expected[0].DateAdded) != 0 {
		t.Errorf("Expected date %s, but got %s", expected[0].DateAdded, processed[0].DateAdded)
	}
}
