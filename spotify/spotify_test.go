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
	tests := []struct {
		name       string
		args       *types.PlexMusicArtist
		wantArtist types.SearchResults
		wantErr    bool
	}{
		{
			name: "artist exists",
			args: &types.PlexMusicArtist{Name: "The Beatles"},
			wantArtist: types.SearchResults{
				SearchURL: "https://open.spotify.com/artist/711MCceyCBcFnzjGY4Q7Un",
				MusicSearchResults: []types.MusicSearchResult{
					{
						Name:   "The Beatles",
						ID:     "b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d",
						Albums: make([]types.MusicSearchAlbumResult, 27),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "artist has special characters",
			args: &types.PlexMusicArtist{Name: "AC/DC"},
			wantArtist: types.SearchResults{
				SearchURL: "https://open.spotify.com/artist/711MCceyCBcFnzjGY4Q7Un",
				MusicSearchResults: []types.MusicSearchResult{
					{
						Name:   "AC/DC",
						ID:     "711MCceyCBcFnzjGY4Q7Un",
						Albums: make([]types.MusicSearchAlbumResult, 21),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArtist, err := SearchSpotifyArtist(tt.args, spotifyClientID, spotifyClientSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchSpotifyArtist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotArtist.MusicSearchResults[0].Name != tt.wantArtist.MusicSearchResults[0].Name {
				t.Errorf("SearchSpotifyArtist() Name = %v, want %v", gotArtist, tt.wantArtist)
			}
			if len(gotArtist.MusicSearchResults[0].Albums) != len(tt.wantArtist.MusicSearchResults[0].Albums) {
				t.Errorf("SearchSpotifyArtist() Albums size = %v, want %v",
					len(gotArtist.MusicSearchResults[0].Albums), len(tt.wantArtist.MusicSearchResults[0].Albums))
			}
		})
	}
}

// debug test for individual artists
func TestSearchSpotifyArtistDebug(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	artist := &types.PlexMusicArtist{Name: "Aaliyah"}
	artistSearchResult, err := SearchSpotifyArtist(artist, spotifyClientID, spotifyClientSecret)
	if err != nil {
		t.Errorf("SearchSpotifyArtist() error = %v", err)
	}
	t.Logf("SearchSpotifyArtist() = %v", artistSearchResult)
}

func TestSearchSpotifyAlbumsDebug(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	artistID := "0urTpYCsixqZwgNTkPJOJ4"
	artistSearchResult, err := SearchSpotifyAlbums(artistID, spotifyClientID, spotifyClientSecret)
	if err != nil {
		t.Errorf("SearchSpotifyAlbum() error = %v", err)
	}
	t.Logf("SearchSpotifyAlbum() = %v", artistSearchResult)
}

func TestSearchSpotifyAlbums(t *testing.T) {
	if spotifyClientID == "" || spotifyClientSecret == "" {
		t.Skip("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET not set")
	}
	type args struct {
		artistID string
	}
	tests := []struct {
		name       string
		args       args
		albumCount int
		wantErr    bool
	}{
		{
			name:       "albums exist",
			args:       args{artistID: "711MCceyCBcFnzjGY4Q7Un"},
			albumCount: 21,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAlbums, err := SearchSpotifyAlbums(tt.args.artistID, spotifyClientID, spotifyClientSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchSpotifyAlbums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotAlbums) != tt.albumCount {
				t.Errorf("SearchSpotifyAlbums() = %v, want %v", len(gotAlbums), tt.albumCount)
			}
		})
	}
}
