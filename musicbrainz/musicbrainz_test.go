package musicbrainz

import (
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

const (
	musicBrainzURL = "https://musicbrainz.org/ws/2"
)

func TestSearchMusicBrainzArtist(t *testing.T) {
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
				SearchURL: "https://musicbrainz.org/artist/b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d",
				MusicSearchResults: []types.MusicArtistSearchResult{

					{
						Name:   "The Beatles",
						ID:     "b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d",
						Albums: make([]types.MusicAlbumSearchResult, 16),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "artist has special characters",
			args: &types.PlexMusicArtist{Name: "AC/DC"},
			wantArtist: types.SearchResults{
				SearchURL: "https://musicbrainz.org/artist/66c662b6-6e2f-4930-8610-912e24c63ed1",
				MusicSearchResults: []types.MusicArtistSearchResult{
					{
						Name:   "AC/DC",
						ID:     "66c662b6-6e2f-4930-8610-912e24c63ed1",
						Albums: make([]types.MusicAlbumSearchResult, 17),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArtist, err := SearchMusicBrainzArtist(tt.args, musicBrainzURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchMusicBrainzArtist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotArtist.MusicSearchResults[0].Name != tt.wantArtist.MusicSearchResults[0].Name {
				t.Errorf("SearchMusicBrainzArtist() Name = %v, want %v", gotArtist, tt.wantArtist)
			}
			if len(gotArtist.MusicSearchResults[0].Albums) != len(tt.wantArtist.MusicSearchResults[0].Albums) {
				t.Errorf("SearchMusicBrainzArtist() Albums size = %v, want %v",
					len(gotArtist.MusicSearchResults[0].Albums), len(tt.wantArtist.MusicSearchResults[0].Albums))
			}
		})
	}
}

// debug test for individual artists
func TestSearchMusicBrainzArtistDebug(t *testing.T) {
	artist := &types.PlexMusicArtist{Name: "Aaliyah"}
	artistSearchResult, err := SearchMusicBrainzArtist(artist, musicBrainzURL)
	if err != nil {
		t.Errorf("SearchMusicBrainzArtist() error = %v", err)
	}
	t.Logf("SearchMusicBrainzArtist() = %v", artistSearchResult)
}

func TestSearchMusicBrainzAlbums(t *testing.T) {
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
			name:       "artist exists",
			args:       args{artistID: "83d91898-7763-47d7-b03b-b92132375c47"},
			albumCount: 13,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAlbums, err := SearchMusicBrainzAlbums(tt.args.artistID, musicBrainzURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchMusicBrainzAlbums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotAlbums) != tt.albumCount {
				t.Errorf("SearchMusicBrainzAlbums() = %v, want %v", len(gotAlbums), tt.albumCount)
			}
		})
	}
}
