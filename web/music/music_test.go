package music

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

func Test_removeOwnedAlbums(t *testing.T) {
	tests := []struct {
		name                string
		args                []types.SearchResults
		wantFilteredResults []types.SearchResults
	}{
		{
			name: "Filter already owned albums",
			args: []types.SearchResults{
				{
					PlexMusicArtist: types.PlexMusicArtist{
						Name: "Test Artist",
						Albums: []types.PlexMusicAlbum{
							{
								Title: "Test Album",
								Year:  "2022",
							},
						},
					},
					MusicSearchResults: []types.MusicSearchResult{
						{
							Name: "Test Artist",
							Albums: []types.MusicSearchAlbumResult{
								{
									Title: "Test Album",
									Year:  "2022",
								},
								{
									Title: "Test Album 2",
									Year:  "2021",
								},
							},
						},
					},
				},
			},
			wantFilteredResults: []types.SearchResults{
				{
					PlexMusicArtist: types.PlexMusicArtist{
						Name: "Test Artist",
						Albums: []types.PlexMusicAlbum{
							{
								Title: "Test Album",
								Year:  "2022",
							},
						},
					},
					MusicSearchResults: []types.MusicSearchResult{
						{
							Name: "Test Artist",
							Albums: []types.MusicSearchAlbumResult{
								{
									Title: "Test Album 2",
									Year:  "2021",
								},
							},
						},
					},
				},
			},
		},
		// remove 2 albums
		{
			name: "Filter 2 owned albums",
			args: []types.SearchResults{
				{
					PlexMusicArtist: types.PlexMusicArtist{
						Name: "Another Artist",
						Albums: []types.PlexMusicAlbum{
							{
								Title: "Another Album",
								Year:  "2021",
							},
							{
								Title: "Another Album 2",
								Year:  "2022",
							},
						},
					},
					MusicSearchResults: []types.MusicSearchResult{
						{
							Name: "Another Artist",
							Albums: []types.MusicSearchAlbumResult{
								{
									Title: "Another Album",
									Year:  "2021",
								},
								{
									Title: "Another Album 2",
									Year:  "2022",
								},
								{
									Title: "Another Album 3",
									Year:  "2023",
								},
							},
						},
					},
				},
			},
			wantFilteredResults: []types.SearchResults{
				{
					PlexMusicArtist: types.PlexMusicArtist{
						Name: "Another Artist",
						Albums: []types.PlexMusicAlbum{
							{
								Title: "Another Album",
								Year:  "2021",
							},
							{
								Title: "Another Album 2",
								Year:  "2022",
							},
						},
					},
					MusicSearchResults: []types.MusicSearchResult{
						{
							Name: "Another Artist",
							Albums: []types.MusicSearchAlbumResult{
								{
									Title: "Another Album 3",
									Year:  "2023",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFilteredResults := removeOwnedAlbums(tt.args); !reflect.DeepEqual(gotFilteredResults, tt.wantFilteredResults) {
				t.Errorf("removeOwnedAlbums() = \n%v\n%v", gotFilteredResults, tt.wantFilteredResults)
			}
		})
	}
}
func Test_removeOlderSearchedAlbums(t *testing.T) {
	tests := []struct {
		name                string
		args                []types.SearchResults
		wantFilteredResults []types.SearchResults
	}{
		// Existing test case
		{
			name: "Filter out search results that are older than 5 years",
			args: []types.SearchResults{
				{
					PlexMusicArtist: types.PlexMusicArtist{
						Name: "Test Artist",
						Albums: []types.PlexMusicAlbum{
							{
								Title: "Test Album",
								Year:  "2015",
							},
						},
					},
					MusicSearchResults: []types.MusicSearchResult{
						{
							Name: "Test Artist",
							Albums: []types.MusicSearchAlbumResult{
								{
									Title: "Test Album 2",
									Year:  "2016",
								},
							},
						},
					},
				},
			},
			wantFilteredResults: []types.SearchResults{
				{
					PlexMusicArtist: types.PlexMusicArtist{
						Name: "Test Artist",
						Albums: []types.PlexMusicAlbum{
							{
								Title: "Test Album",
								Year:  "2015",
							},
						},
					},
					MusicSearchResults: []types.MusicSearchResult{
						{
							Name:   "Test Artist",
							Albums: []types.MusicSearchAlbumResult{},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilteredResults := removeOlderSearchedAlbums(tt.args)
			if !reflect.DeepEqual(gotFilteredResults, tt.wantFilteredResults) {
				t.Errorf("removeOlderSearchedAlbums() = %v, want %v", gotFilteredResults, tt.wantFilteredResults)
			}
		})
	}
}
func Test_cleanAlbums(t *testing.T) {
	original := []types.MusicSearchAlbumResult{
		{
			Title: "Album 1",
			Year:  "2022",
		},
		{
			Title: "Album 2",
			Year:  "2021",
		},
		{
			Title: "Album 3",
			Year:  "2020",
		},
	}
	toRemove := []types.MusicSearchAlbumResult{
		{
			Title: "Album 2",
			Year:  "2021",
		},
		{
			Title: "Album 3",
			Year:  "2020",
		},
	}
	want := []types.MusicSearchAlbumResult{
		{
			Title: "Album 1",
			Year:  "2022",
		},
	}

	if got := cleanAlbums(original, toRemove); !reflect.DeepEqual(got, want) {
		t.Errorf("cleanAlbums() got\n%v\nwant\n%v", got, want)
	}
}
