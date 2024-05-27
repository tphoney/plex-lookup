package utils

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tphoney/plex-lookup/types"
)

func MarkBestMatch(search *types.SearchResults) types.SearchResults {
	expectedYear := YearToDate(search.PlexMovie.Year)
	for i := range search.MovieSearchResults {
		// normally a match if the year is within 1 year of each other
		resultYear := YearToDate(search.MovieSearchResults[i].Year)
		if search.MovieSearchResults[i].FoundTitle == search.PlexMovie.Title && WitinOneYear(resultYear.Year(), expectedYear.Year()) {
			search.MovieSearchResults[i].BestMatch = true
			if search.MovieSearchResults[i].Format == types.DiskBluray {
				search.MatchesBluray++
			}
			if search.MovieSearchResults[i].Format == types.Disk4K {
				search.Matches4k++
			}
		}
	}
	expectedYear = YearToDate(search.PlexTVShow.Year)
	for i := range search.TVSearchResults {
		resultYear := YearToDate(search.TVSearchResults[i].Year)
		if search.TVSearchResults[i].FoundTitle == search.PlexTVShow.Title && WitinOneYear(resultYear.Year(), expectedYear.Year()) {
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

func CompareTitles(title1, title2 string) bool {
	// remove anything between ()
	r := regexp.MustCompile(`\((.*?)\)`)
	title1 = r.ReplaceAllString(title1, "")
	title2 = r.ReplaceAllString(title2, "")
	// remove anything between []
	r = regexp.MustCompile(`\[(.*?)\]`)
	title1 = r.ReplaceAllString(title1, "")
	title2 = r.ReplaceAllString(title2, "")
	// remove anything between {}
	r = regexp.MustCompile(`\{(.*?)\}`)
	title1 = r.ReplaceAllString(title1, "")
	title2 = r.ReplaceAllString(title2, "")
	// strip whitespace
	title1 = strings.TrimSpace(title1)
	title2 = strings.TrimSpace(title2)
	// lowercase
	title1 = strings.ToLower(title1)
	title2 = strings.ToLower(title2)
	return title1 == title2
}

func WitinOneYear(year1, year2 int) bool {
	return year1 == year2 || year1 == year2-1 || year1 == year2+1
}
