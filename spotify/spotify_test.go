package spotify

import (
	"context"
	"os"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

var (
	spotifyClientID     = os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
)

func TestGetArtistsInParallel(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	token, err := SpotifyOAuthToken(context.Background(), spotifyClientID, spotifyClientSecret)
	if err != nil {
		t.Errorf("GetSpotifyToken() returned an error: %s", err)
	}

	plexArtists := []types.PlexMusicArtist{
		{Name: "The Beatles"},
		{Name: "The Rolling Stones"},
		{Name: "The Who"},
		{Name: "The Kinks"},
	}

	got := GetArtistsInParallel(plexArtists, token)

	if len(got) != len(plexArtists) {
		t.Errorf("GetSpotifyArtistsInParallel() returned %d results, expected %d", len(got), len(plexArtists))
	}
}

func TestSearchSpotifyArtist(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	token, err := SpotifyOAuthToken(context.Background(), spotifyClientID, spotifyClientSecret)
	if err != nil {
		t.Errorf("GetSpotifyToken() returned an error: %s", err)
	}

	plexArtist := &types.PlexMusicArtist{Name: "The Beatles"}
	got := searchSpotifyArtistValue(plexArtist, token)

	if len(got.MusicSearchResults) != 1 {
		t.Fatalf("SearchSpotifyArtist() returned %d results, expected 1", len(got.MusicSearchResults))
	}

	expectedArtist := types.MusicArtistSearchResult{
		Name: "The Beatles",
		ID:   "3WrFJ7ztbogyGnTHbHJFl2",
		URL:  "https://open.spotify.com/artist/3WrFJ7ztbogyGnTHbHJFl2",
	}

	if got.MusicSearchResults[0].Name != expectedArtist.Name {
		t.Errorf("SearchSpotifyArtist() returned %s, expected %s", got.MusicSearchResults[0].Name, expectedArtist.Name)
	}
	if got.MusicSearchResults[0].ID != expectedArtist.ID {
		t.Errorf("SearchSpotifyArtist() returned %s, expected %s", got.MusicSearchResults[0].ID, expectedArtist.ID)
	}
	if got.MusicSearchResults[0].URL != expectedArtist.URL {
		t.Errorf("SearchSpotifyArtist() returned %s, expected %s", got.MusicSearchResults[0].URL, expectedArtist.URL)
	}
}

// debug test for individual artists
func TestSearchSpotifyArtistDebug(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	token, err := SpotifyOAuthToken(context.Background(), spotifyClientID, spotifyClientSecret)
	if err != nil {
		t.Errorf("GetSpotifyToken() returned an error: %s", err)
	}

	plexArtist := &types.PlexMusicArtist{Name: "Angel Olsen"}

	got := searchSpotifyArtistValue(plexArtist, token)

	t.Logf("SearchSpotifyArtist() = %+v", got)
}

func TestSearchSpotifyAlbums(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	token, err := SpotifyOAuthToken(context.Background(), spotifyClientID, spotifyClientSecret)
	if err != nil {
		t.Errorf("GetSpotifyToken() returned an error: %s", err)
	}
	type args struct {
		m *types.MusicSearchResponse
	}
	tests := []struct {
		name       string
		args       args
		albumCount int
		wantErr    bool
	}{
		{
			name:       "albums exist",
			args:       args{m: &types.MusicSearchResponse{MusicSearchResults: []types.MusicArtistSearchResult{{ID: "711MCceyCBcFnzjGY4Q7Un"}}}},
			albumCount: 21,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searchSpotifyAlbumValue(tt.args.m, token)

			if len(got.MusicSearchResults) == 0 {
				t.Fatalf("SearchSpotifyAlbums() returned no music search results")
			}
			if len(got.MusicSearchResults[0].FoundAlbums) != tt.albumCount {
				t.Errorf("SearchSpotifyAlbums() = %v, want %v", len(got.MusicSearchResults[0].FoundAlbums), tt.albumCount)
			}
		})
	}
}

func TestSearchSpotifyAlbumsDebug(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	token, err := SpotifyOAuthToken(context.Background(), spotifyClientID, spotifyClientSecret)
	if err != nil {
		t.Errorf("GetSpotifyToken() returned an error: %s", err)
	}

	want := types.MusicSearchResponse{
		MusicSearchResults: []types.MusicArtistSearchResult{
			{
				Name: "",
				ID:   "16GcWuvvybAoaHr0NqT8Eh",
				URL:  "",
			},
		},
	}

	got := searchSpotifyAlbumValue(&want, token)

	t.Logf("SearchSpotifyAlbum() = %v", got.MusicSearchResults)
}
