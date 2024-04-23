package utils

import (
	"slices"
	"strconv"
	"time"

	"github.com/tphoney/plex-lookup/types"
)

// nolint: dupl, nolintlint
func MarkBestMatch(search *types.SearchResults) types.SearchResults {
	expectedYear := YearToDate(search.PlexMovie.Year)
	for i := range search.MovieSearchResults {
		// normally a match if the year is within 1 year of each other
		resultYear := YearToDate(search.MovieSearchResults[i].Year)
		if search.MovieSearchResults[i].FoundTitle == search.PlexMovie.Title && (resultYear.Year() == expectedYear.Year() ||
			resultYear.Year() == expectedYear.Year()-1 || resultYear.Year() == expectedYear.Year()+1) {
			search.MovieSearchResults[i].BestMatch = true
			if search.MovieSearchResults[i].Format == types.DiskBluray {
				search.MatchesBluray++
			}
			if search.MovieSearchResults[i].Format == types.Disk4K {
				search.Matches4k++
			}
		}
	}
	for i := range search.TVSearchResults {
		resultYear := YearToDate(search.TVSearchResults[i].Year)
		if search.TVSearchResults[i].FoundTitle == search.PlexTVShow.Title && (resultYear.Year() == expectedYear.Year() ||
			resultYear.Year() == expectedYear.Year()-1 || resultYear.Year() == expectedYear.Year()+1) {
			search.TVSearchResults[i].BestMatch = true
			if slices.Contains(search.TVSearchResults[i].Format, types.DiskBluray) {
				search.MatchesBluray++
			}
			if slices.Contains(search.TVSearchResults[i].Format, types.Disk4K) {
				search.Matches4k++
			}
		}
	}
	return *search
}

func YearToDate(yearString string) time.Time {
	year, err := strconv.Atoi(yearString)
	if err != nil {
		return time.Time{}
	}
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
}
