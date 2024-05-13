package spotify

import (
	"os"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

var (
	spotifyClientID     = os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
)

func TestSearchSpotifyArtist(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}

	plexArtist := &types.PlexMusicArtist{Name: "The Beatles"}

	ch := make(chan *types.SearchResults, 1)
	SearchSpotifyArtist(plexArtist, spotifyClientID, spotifyClientSecret, ch)

	got := <-ch
	if len(got.MusicSearchResults) != 1 {
		t.Errorf("SearchSpotifyArtist() returned %d results, expected 1", len(got.MusicSearchResults))
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
	plexArtist := &types.PlexMusicArtist{Name: "Angel Olsen"}

	ch := make(chan *types.SearchResults, 1)
	SearchSpotifyArtist(plexArtist, spotifyClientID, spotifyClientSecret, ch)

	got := <-ch
	t.Logf("SearchSpotifyArtist() = %+v", got)
}

func TestSearchSpotifyAlbums(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	type args struct {
		m *types.SearchResults
	}
	tests := []struct {
		name       string
		args       args
		albumCount int
		wantErr    bool
	}{
		{
			name:       "albums exist",
			args:       args{m: &types.SearchResults{MusicSearchResults: []types.MusicArtistSearchResult{{ID: "711MCceyCBcFnzjGY4Q7Un"}}}},
			albumCount: 21,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan *types.SearchResults, 1)
			SearchSpotifyAlbum(tt.args.m, spotifyClientID, spotifyClientSecret, ch)
			got := <-ch

			if len(got.MusicSearchResults[0].Albums) != tt.albumCount {
				t.Errorf("SearchSpotifyAlbums() = %v, want %v", len(got.MusicSearchResults[0].Albums), tt.albumCount)
			}
		})
	}
}

func TestSearchSpotifyAlbumsDebug(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	want := types.SearchResults{
		MusicSearchResults: []types.MusicArtistSearchResult{
			{
				Name: "",
				ID:   "16GcWuvvybAoaHr0NqT8Eh",
				URL:  "",
			},
		},
	}
	ch := make(chan *types.SearchResults, 1)
	SearchSpotifyAlbum(&want, spotifyClientID, spotifyClientSecret, ch)
	got := <-ch

	t.Logf("SearchSpotifyAlbum() = %v", got.MusicSearchResults)
}

func TestSpotifyLookupSimilarArtists(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	tests := []struct {
		name       string
		artistID   string
		wantErr    bool
		wantLength int
	}{
		{
			name:       "similar artists exist",
			artistID:   "711MCceyCBcFnzjGY4Q7Un",
			wantErr:    false,
			wantLength: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArtists, err := SearchSpotifySimilarArtist(tt.artistID, spotifyClientID, spotifyClientSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("SpotifyLookupSimilarArtists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotArtists.Artists) != tt.wantLength {
				t.Errorf("SpotifyLookupSimilarArtists() = %v, want %v", len(gotArtists.Artists), tt.wantLength)
			}
		})
	}
}

// func TestFindSimilarArtists(t *testing.T) {
// 	if spotifyClientID == "" || spotifyClientSecret == "" {
// 		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
// 	}
// 	ownedArtists := []types.MusicArtistSearchResult{
// 		{
// 			Name: "The Beatles",s
// 			ID:   "3WrFJ7ztbogyGnTHbHJFl2",
// 			URL:  "https://open.spotify.com/artist/3WrFJ7ztbogyGnTHbHJFl2",
// 		},
// 		{
// 			Name: "Queens of the Stone Age",
// 			ID:   "4pejUc4iciQfgdX6OKulQn",
// 			URL:  "https://open.spotify.com/artist/4pejUc4iciQfgdX6OKulQn",
// 		},
// 		{
// 			Name: "Eagles of Death Metal",
// 			ID:   "02uYdhMhCgdB49hZlYRm9o",
// 			URL:  "https://open.spotify.com/artist/02uYdhMhCgdB49hZlYRm9o",
// 		},
// 	}
