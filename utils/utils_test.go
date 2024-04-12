package utils

import (
	"testing"
	"time"

	"github.com/tphoney/plex-lookup/types"
)

func TestYearToDate(t *testing.T) {
	// Test case 1: Valid year string
	yearString := "2022"
	expectedDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	result := YearToDate(yearString)
	if result != expectedDate {
		t.Errorf("Expected date %v, but got %v", expectedDate, result)
	}

	// Test case 2: Invalid year string
	yearString = "abcd"
	expectedDate = time.Time{}
	result = YearToDate(yearString)
	if result != expectedDate {
		t.Errorf("Expected date %v, but got %v", expectedDate, result)
	}
}
func TestMarkBestMatch(t *testing.T) {
	// Test case 1: Matching title and year within 1 year
	search := types.MovieSearchResults{
		PlexMovie: types.PlexMovie{
			Title: "Movie Title",
			Year:  "2022",
		},
		SearchResults: []types.SearchResult{
			{
				FoundTitle: "Movie Title",
				Year:       "2022",
			},
		},
	}
	expectedResults := []types.SearchResult{
		{
			FoundTitle: "Movie Title",
			Year:       "2022",
			BestMatch:  true,
		},
	}
	result := MarkBestMatch(&search)
	if len(result.SearchResults) != len(expectedResults) {
		t.Errorf("Expected %d search results, but got %d", len(expectedResults), len(result.SearchResults))
	} else {
		for i := range result.SearchResults {
			if result.SearchResults[i] != expectedResults[i] {
				t.Errorf("Expected search result %v, but got %v", expectedResults[i], result.SearchResults[i])
			}
		}
	}

	// Test case 2: Non-matching title
	search = types.MovieSearchResults{
		PlexMovie: types.PlexMovie{
			Title: "Movie Title",
			Year:  "2022",
		},
		SearchResults: []types.SearchResult{
			{
				FoundTitle: "Other Movie",
				Year:       "2022",
			},
		},
	}
	expectedResults = []types.SearchResult{
		{
			FoundTitle: "Other Movie",
			Year:       "2022",
		},
	}
	result = MarkBestMatch(&search)
	if len(result.SearchResults) != len(expectedResults) {
		t.Errorf("Expected %d search results, but got %d", len(expectedResults), len(result.SearchResults))
	} else {
		for i := range result.SearchResults {
			if result.SearchResults[i] != expectedResults[i] {
				t.Errorf("Expected search result %v, but got %v", expectedResults[i], result.SearchResults[i])
			}
		}
	}

	// Test case 3: Non-matching year
	search = types.MovieSearchResults{
		PlexMovie: types.PlexMovie{
			Title: "Movie Title",
			Year:  "2022",
		},
		SearchResults: []types.SearchResult{
			{
				FoundTitle: "Movie Title",
				Year:       "2024",
			},
		},
	}
	expectedResults = []types.SearchResult{
		{
			FoundTitle: "Movie Title",
			Year:       "2024",
		},
	}
	result = MarkBestMatch(&search)
	if len(result.SearchResults) != len(expectedResults) {
		t.Errorf("Expected %d search results, but got %d", len(expectedResults), len(result.SearchResults))
	} else {
		for i := range result.SearchResults {
			if result.SearchResults[i] != expectedResults[i] {
				t.Errorf("Expected search result %v, but got %v", expectedResults[i], result.SearchResults[i])
			}
		}
	}
}
