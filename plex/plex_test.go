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
			Title:      "Chaos Theory",
			Year:       "2007",
			Resolution: "sd",
			DateAdded:  time.Date(2023, time.January, 21, 15, 03, 10, 0, time.FixedZone("GMT", 0)),
		},
	}

	if len(processed) != 3 {
		t.Fatalf("Expected 3 movies, but got %d", len(processed))
	}

	if len(expected) == 0 {
		t.Fatalf("Expected slice is empty")
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

	if processed[0].Resolution != expected[0].Resolution {
		t.Errorf("Expected resolution %s, but got %s", expected[0].Resolution, processed[0].Resolution)
	}
}

func TestGetPlexMovies(t *testing.T) {
	if plexIP == "" || plexMovieLibraryID == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	result := AllMovies(plexIP, plexMovieLibraryID, plexToken)

	if len(result) == 0 {
		t.Errorf("Expected at least one TV show, but got %d", len(result))
	}
}

func TestGetPlexTV(t *testing.T) {
	if plexIP == "" || plexTVLibraryID == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	result := AllTV(plexIP, plexToken, plexTVLibraryID)

	if len(result) == 0 {
		t.Fatalf("Expected at least one TV show, but got %d", len(result))
	}

	if len(result[0].Seasons) == 0 {
		t.Fatalf("Expected at least one season, but got %d", len(result[0].Seasons))
	}

	if len(result[0].Seasons[0].Episodes) == 0 {
		t.Fatalf("Expected at least one episode, but got %d", len(result[0].Seasons[0].Episodes))
	}
}

func TestDebugGetPlexTVSeasons(t *testing.T) {
	if plexIP == "" || plexTVLibraryID == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	result := getPlexTVSeasons(plexIP, plexToken, "5383")

	if len(result) == 0 {
		t.Fatalf("Expected at least one TV show, but got %d", len(result))
	}

	if len(result[0].Episodes) == 0 {
		t.Fatalf("Expected at least one episode, but got %d", len(result[0].Episodes))
	}
}

func TestGetPlexMusic(t *testing.T) {
	if plexIP == "" || plexMusicLibraryID == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	result := AllMusicArtists(plexIP, plexToken, plexMusicLibraryID)

	if len(result) == 0 {
		t.Fatalf("Expected at least one album, but got %d", len(result))
	}

	// first artist should have at least one album
	if len(result[0].Albums) == 0 {
		t.Fatalf("Expected at least one album, but got %d", len(result[0].Albums))
	}
}

func TestGetPlaylists(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	playlists, err := GetPlaylists(plexIP, plexToken, "2")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Check the number of playlists
	if len(playlists) == 0 {
		t.Errorf("Expected at least one playlist, but got %d", len(playlists))
	}
}

func TestGetArtistsFromPlaylist(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	items := GetArtistsFromPlaylist(plexIP, plexToken, "111897")
	// Check the number of items
	if len(items) == 0 {
		t.Errorf("Expected at least one item, but got %d", len(items))
	}
}

func TestGetMoviesFromPlaylist(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	items := GetMoviesFromPlaylist(plexIP, plexToken, "111907")
	// Check the number of items
	if len(items) == 0 {
		t.Errorf("Expected at least one item, but got %d", len(items))
	}
}

func TestGetTVFromPlaylist(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	items := GetTVFromPlaylist(plexIP, plexToken, "111908")
	// Check the number of items
	if len(items) == 0 {
		t.Errorf("Expected at least one item, but got %d", len(items))
	}
}
func Test_findLowestResolution(t *testing.T) {
	tests := []struct {
		name                 string
		resolutions          []string
		wantLowestResolution string
	}{
		{
			name:                 "SD is lowest",
			resolutions:          []string{types.PlexResolutionSD, types.PlexResolution240, types.PlexResolution720, types.PlexResolution1080},
			wantLowestResolution: types.PlexResolutionSD,
		},
		{
			name:                 "4k is lowest",
			resolutions:          []string{types.PlexResolution4K, types.PlexResolution4K},
			wantLowestResolution: types.PlexResolution4K,
		},
		{
			name:                 "720 is lowest",
			resolutions:          []string{types.PlexResolution720, types.PlexResolution1080, types.PlexResolution720, types.PlexResolution1080},
			wantLowestResolution: types.PlexResolution720,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLowestResolution := findLowestResolution(tt.resolutions); gotLowestResolution != tt.wantLowestResolution {
				t.Errorf("findLowestResolution() = %v, want %v", gotLowestResolution, tt.wantLowestResolution)
			}
		})
	}
}

func Test_parsePlexDate(t *testing.T) {
	tests := []struct {
		name           string
		plexDate       string
		wantParsedDate time.Time
	}{
		{
			name:           "validate int date",
			plexDate:       "1676229015",
			wantParsedDate: time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:           "validate int date",
			plexDate:       "0",
			wantParsedDate: time.Date(1970, time.January, 1, 0, 0, 16, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotParsedDate := parsePlexDate(tt.plexDate); gotParsedDate.Compare(tt.wantParsedDate) == 0 {
				t.Errorf("parsePlexDate()\n%v\nwant\n%v", gotParsedDate, tt.wantParsedDate)
			}
		})
	}
}
