package music

import (
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
