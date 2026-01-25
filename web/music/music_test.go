package music

import (
	"reflect"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

func TestRemoveOwnedFromSearchResults(t *testing.T) {
	tests := []struct {
		name     string
		original []types.MusicAlbumSearchResult
		toRemove []string
		wantIDs  []string
	}{
		{
			name: "NoToRemove",
			original: []types.MusicAlbumSearchResult{
				{ID: "1", Title: "Album1"},
				{ID: "2", Title: "Album2"},
			},
			toRemove: []string{},
			wantIDs:  []string{"1", "2"},
		},
		{
			name: "RemoveOne",
			original: []types.MusicAlbumSearchResult{
				{ID: "1", Title: "Album1"},
				{ID: "2", Title: "Album2"},
				{ID: "3", Title: "Album3"},
			},
			toRemove: []string{"2"},
			wantIDs:  []string{"1", "3"},
		},
		{
			name: "RemoveAll",
			original: []types.MusicAlbumSearchResult{
				{ID: "1", Title: "Album1"},
				{ID: "2", Title: "Album2"},
			},
			toRemove: []string{"1", "2"},
			wantIDs:  []string{},
		},
		{
			name: "RemoveNoneMatch",
			original: []types.MusicAlbumSearchResult{
				{ID: "1", Title: "Album1"},
				{ID: "2", Title: "Album2"},
			},
			toRemove: []string{"3", "4"},
			wantIDs:  []string{"1", "2"},
		},
		{
			name:     "EmptyOriginal",
			original: []types.MusicAlbumSearchResult{},
			toRemove: []string{"1"},
			wantIDs:  []string{},
		},
	}

	for _, tt := range tests {
		result := removeOwnedFromSearchResults(tt.original, tt.toRemove)
		if len(result) != len(tt.wantIDs) {
			t.Errorf("Test %q: Expected %d albums, got %d", tt.name, len(tt.wantIDs), len(result))
		}
		gotIDs := make(map[string]bool)
		for _, album := range result {
			gotIDs[album.ID] = true
		}
		for _, wantID := range tt.wantIDs {
			if !gotIDs[wantID] {
				t.Errorf("Test %q: Expected album ID %q in result, but it was missing", tt.name, wantID)
			}
		}
		if len(result) != len(tt.wantIDs) {
			continue
		}
		// Optionally check order if needed
		for i, wantID := range tt.wantIDs {
			if result[i].ID != wantID {
				t.Errorf("Test %q: Expected album ID at index %d to be %q, got %q", tt.name, i, wantID, result[i].ID)
			}
		}
	}
}

func TestFindMatchingAlbumFromSearch(t *testing.T) {
	plexAlbum := types.PlexMusicAlbum{
		Title:     "So‐Called Chaos",
		RatingKey: "94388",
		Year:      "2004",
	}

	original := []types.MusicAlbumSearchResult{
		{SanitizedTitle: "the storm before the calm", ID: "id0"},
		{SanitizedTitle: "such pretty forks in the mix", ID: "id1"},
		{SanitizedTitle: "such pretty forks in the road", ID: "id2"},
		{SanitizedTitle: "jagged little pill", ID: "id3"},
		{SanitizedTitle: "jagged little pill", ID: "id4"},
		{SanitizedTitle: "jagged little pill", ID: "id5"},
		{SanitizedTitle: "jagged little pill", ID: "id6"},
		{SanitizedTitle: "live at montreux 2012", ID: "id7"},
		{SanitizedTitle: "havoc and bright lights", ID: "id8"},
		{SanitizedTitle: "flavours of entanglement", ID: "id9"},
		{SanitizedTitle: "flavours of entanglement", ID: "id10"},
		{SanitizedTitle: "jagged little pill", ID: "id11"},
		{SanitizedTitle: "so-called chaos", ID: "id12"},
		{SanitizedTitle: "feast on scraps", ID: "id13"},
		{SanitizedTitle: "under rug swept", ID: "id14"},
		{SanitizedTitle: "live / unplugged", ID: "id15"},
		{SanitizedTitle: "supposed former infatuation junkie", ID: "id16"},
		{SanitizedTitle: "supposed former infatuation junkie", ID: "id17"},
		{SanitizedTitle: "jagged little pill", ID: "id18"},
	}

	foundIDs := findMatchingAlbumFromSearch(plexAlbum, original)

	expected := []string{"id12"}
	if !reflect.DeepEqual(foundIDs, expected) {
		t.Errorf("Expected %v, got %v", expected, foundIDs)
	}
}
