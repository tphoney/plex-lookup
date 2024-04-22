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
	search := types.SearchResults{
		PlexMovie: types.PlexMovie{
			Title: "Movie Title",
			Year:  "2022",
		},
		MovieSearchResults: []types.MovieSearchResult{
			{
				FoundTitle: "Movie Title",
				Year:       "2022",
			},
		},
	}
	expectedResults := []types.MovieSearchResult{
		{
			FoundTitle: "Movie Title",
			Year:       "2022",
			BestMatch:  true,
		},
	}
	result := MarkBestMatch(&search)
	if len(result.MovieSearchResults) != len(expectedResults) {
		t.Errorf("Expected %d search results, but got %d", len(expectedResults), len(result.MovieSearchResults))
	} else {
		for i := range result.MovieSearchResults {
			if result.MovieSearchResults[i] != expectedResults[i] {
				t.Errorf("Expected search result %v, but got %v", expectedResults[i], result.MovieSearchResults[i])
			}
		}
	}

	// Test case 2: Non-matching title
	search = types.SearchResults{
		PlexMovie: types.PlexMovie{
			Title: "Movie Title",
			Year:  "2022",
		},
		MovieSearchResults: []types.MovieSearchResult{
			{
				FoundTitle: "Other Movie",
				Year:       "2022",
			},
		},
	}
	expectedResults = []types.MovieSearchResult{
		{
			FoundTitle: "Other Movie",
			Year:       "2022",
		},
	}
	result = MarkBestMatch(&search)
	if len(result.MovieSearchResults) != len(expectedResults) {
		t.Errorf("Expected %d search results, but got %d", len(expectedResults), len(result.MovieSearchResults))
	} else {
		for i := range result.MovieSearchResults {
			if result.MovieSearchResults[i] != expectedResults[i] {
				t.Errorf("Expected search result %v, but got %v", expectedResults[i], result.MovieSearchResults[i])
			}
		}
	}

	// Test case 3: Non-matching year
	search = types.SearchResults{
		PlexMovie: types.PlexMovie{
			Title: "Movie Title",
			Year:  "2022",
		},
		MovieSearchResults: []types.MovieSearchResult{
			{
				FoundTitle: "Movie Title",
				Year:       "2024",
			},
		},
	}
	expectedResults = []types.MovieSearchResult{
		{
			FoundTitle: "Movie Title",
			Year:       "2024",
		},
	}
	result = MarkBestMatch(&search)
	if len(result.MovieSearchResults) != len(expectedResults) {
		t.Errorf("Expected %d search results, but got %d", len(expectedResults), len(result.MovieSearchResults))
	} else {
		for i := range result.MovieSearchResults {
			if result.MovieSearchResults[i] != expectedResults[i] {
				t.Errorf("Expected search result %v, but got %v", expectedResults[i], result.MovieSearchResults[i])
			}
		}
	}
}
