package utils

import (
	"fmt"
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
func TestMarkBestMatchMovie(t *testing.T) {
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
	result := MarkBestMatchMovie(&search)
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
	result = MarkBestMatchMovie(&search)
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
	result = MarkBestMatchMovie(&search)
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

func Test_albumTitlesMatch(t *testing.T) {
	tests := []struct {
		title1 string
		title2 string
		want   bool
	}{
		{
			title1: "Test Album",
			title2: "Test Album",
			want:   true,
		},
		{
			title1: "Test Album (Deluxe Edition)",
			title2: "Test Album",
			want:   true,
		},
		{
			title1: "Test Album [Remastered]",
			title2: "Test Album",
			want:   true,
		},
		{
			title1: "Test Album {Special Edition}",
			title2: "Test Album",
			want:   true,
		},
		{
			title1: "Test Album (Deluxe Edition) [Remastered] {Special Edition}",
			title2: "Test Album",
			want:   true,
		},
		{
			title1: "Test Album (Live)",
			title2: "Test Album (Studio)",
			want:   true,
		},
		{
			title1: "Test Album [Remastered]",
			title2: "Test Album2 [Deluxe Edition]",
			want:   false,
		},
		// test for case insensitivity
		{
			title1: "Test Album",
			title2: "test album",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("title1=%s, title2=%s", tt.title1, tt.title2), func(t *testing.T) {
			if got := CompareAlbumTitles(tt.title1, tt.title2); got != tt.want {
				t.Errorf("albumTitlesMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestWitinOneYear(t *testing.T) {
	// Test case 1: Same year
	year1 := 2022
	year2 := 2022
	expectedResult := true
	result := WitinOneYear(year1, year2)
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// Test case 2: Year difference of 1
	year1 = 2022
	year2 = 2021
	expectedResult = true
	result = WitinOneYear(year1, year2)
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// Test case 3: Year difference of -1
	year1 = 2022
	year2 = 2023
	expectedResult = true
	result = WitinOneYear(year1, year2)
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// Test case 4: Year difference greater than 1
	year1 = 2022
	year2 = 2020
	expectedResult = false
	result = WitinOneYear(year1, year2)
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}
}

func Test_matchTVShow(t *testing.T) {
	type args struct {
		plexTitle  string
		foundTitle string
		foundYear  int
		lowerBound int
		upperBound int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "colons in title",
			args: args{
				plexTitle:  "Stargate origins",
				foundTitle: "Stargate: Origins",
				foundYear:  2018,
				lowerBound: 2018,
				upperBound: 2018,
			},
			want: true,
		},
		{
			name: "Peter Serafinowicz Show",
			args: args{
				plexTitle:  "The Peter Serafinowicz Show",
				foundTitle: "Peter Serafinowicz Show",
				foundYear:  2008,
				lowerBound: 2007,
				upperBound: 2009,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchTitle(tt.args.plexTitle, tt.args.foundTitle, tt.args.foundYear, tt.args.lowerBound, tt.args.upperBound); got != tt.want {
				t.Errorf("matchTVShow() = %v, want %v", got, tt.want)
			}
		})
	}
}
