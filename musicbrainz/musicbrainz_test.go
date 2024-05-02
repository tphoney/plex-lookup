package musicbrainz

import "testing"

func TestSearchMusicBrainzArtist(t *testing.T) {
	type args struct {
		artistName string
	}
	tests := []struct {
		name       string
		args       args
		wantArtist MusicBrainzArtist
		wantErr    bool
	}{
		{
			name:       "artist exists",
			args:       args{artistName: "The Beatles"},
			wantArtist: MusicBrainzArtist{Name: "The Beatles", ID: "b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArtist, err := SearchMusicBrainzArtist(tt.args.artistName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchMusicBrainzArtist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotArtist != tt.wantArtist {
				t.Errorf("SearchMusicBrainzArtist() = %v, want %v", gotArtist, tt.wantArtist)
			}
		})
	}
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
			gotAlbums, err := SearchMusicBrainzAlbums(tt.args.artistID)
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
