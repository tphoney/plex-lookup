package utils

import (
	"strconv"
	"time"

	"github.com/tphoney/plex-lookup/types"
)

func MarkBestMatch(search types.MovieSearchResults) []types.SearchResult {
	expectedYear := YearToDate(search.Movie.Year)
	for i := range search.SearchResults {
		// normally a match if the year is within 1 year of each other
		resultYear := YearToDate(search.SearchResults[i].Year)
		if search.SearchResults[i].FoundTitle == search.Movie.Title && (resultYear.Year() == expectedYear.Year() ||
			resultYear.Year() == expectedYear.Year()-1 || resultYear.Year() == expectedYear.Year()+1) {
			search.SearchResults[i].BestMatch = true
		}
	}
	return search.SearchResults
}

func YearToDate(yearString string) time.Time {
	year, err := strconv.Atoi(yearString)
	if err != nil {
		return time.Time{}
	}
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
}
