package plex

import (
	"os"
	"testing"
	"time"

	types "github.com/tphoney/plex-lookup/types"
)

var (
	plexIP             = os.Getenv("PLEX_IP")
	plexToken          = os.Getenv("PLEX_TOKEN")
	plexMovieLibraryID = os.Getenv("PLEX_MOVIE_LIBRARY_ID")
	plexTVLibraryID    = os.Getenv("PLEX_TV_LIBRARY_ID")
	plexMusicLibraryID = os.Getenv("PLEX_MUSIC_LIBRARY_ID")
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

func TestGetPlexMovies(t *testing.T) {
	if plexIP == "" || plexMovieLibraryID == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	result := GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, "", nil)

	if len(result) == 0 {
		t.Errorf("Expected at least one TV show, but got %d", len(result))
	}
}

func TestGetPlexTV(t *testing.T) {
	if plexIP == "" || plexTVLibraryID == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	result := GetPlexTV(plexIP, plexTVLibraryID, plexToken, []string{})

	if len(result) == 0 {
		t.Errorf("Expected at least one TV show, but got %d", len(result))
	}

	if len(result[0].Seasons) == 0 {
		t.Errorf("Expected at least one season, but got %d", len(result[0].Seasons))
	}

	if len(result[0].Seasons[0].Episodes) == 0 {
		t.Errorf("Expected at least one episode, but got %d", len(result[0].Seasons[0].Episodes))
	}
}

func TestGetPlexMusic(t *testing.T) {
	if plexIP == "" || plexMusicLibraryID == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	result := GetPlexMusicArtists(plexIP, plexMusicLibraryID, plexToken)

	if len(result) == 0 {
		t.Errorf("Expected at least one album, but got %d", len(result))
	}

	if len(result) == 0 {
		t.Errorf("Expected at least one artist, but got %d", len(result))
	}
	// first artist should have at least one album
	if len(result[0].Albums) == 0 {
		t.Errorf("Expected at least one album, but got %d", len(result[0].Albums))
	}
}
