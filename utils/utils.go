package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rainycape/unidecode"

	"github.com/tphoney/plex-lookup/types"
)

// MarkBestMatchTVResponse marks best matches for TVSearchResponse (new type)
func MarkBestMatchTVResponse(search *types.TVSearchResponse) types.TVSearchResponse {
	firstEpisodeBoundry := search.FirstEpisodeAired.Year() - 1
	lastEpisodeBoundry := search.LastEpisodeAired.Year() + 1
	for i := range search.TVSearchResults {
		resultYear := YearToDate(search.TVSearchResults[i].FirstAiredYear)
		// If year is empty, match on title only (skip year check)
		if search.TVSearchResults[i].FirstAiredYear == "" {
			if matchTitleNoYear(search.Title, search.TVSearchResults[i].FoundTitle) {
				search.TVSearchResults[i].BestMatch = true
			}
		} else {
			// Original logic for shows with years
			if matchTitle(search.Title, search.TVSearchResults[i].FoundTitle,
				resultYear.Year(), firstEpisodeBoundry, lastEpisodeBoundry) {
				search.TVSearchResults[i].BestMatch = true
			}
		}
	}
	return *search
}

// MarkBestMatchMovieResponse marks best matches for MovieSearchResponse (new type)
func MarkBestMatchMovieResponse(search *types.MovieSearchResponse) types.MovieSearchResponse {
	lowerBound := YearToDate(search.PlexMovie.Year).Year() - 1
	upperBound := YearToDate(search.PlexMovie.Year).Year() + 1
	for i := range search.MovieSearchResults {
		// normally a match if the year is within 1 year of each other
		resultYear := YearToDate(search.MovieSearchResults[i].Year)
		if matchTitle(search.Title, search.MovieSearchResults[i].FoundTitle, resultYear.Year(), lowerBound, upperBound) {
			search.MovieSearchResults[i].BestMatch = true
			if search.MovieSearchResults[i].Format == types.DiskBluray {
				search.MatchesBluray++
			}
			if search.MovieSearchResults[i].Format == types.Disk4K {
				search.Matches4k++
			}
		}
	}
	return *search
}

func matchTitleNoYear(plexTitle, foundTitle string) bool {
	// ...existing code...
	plexTitle = strings.ToLower(plexTitle)
	foundTitle = strings.ToLower(foundTitle)
	remove := []string{"the"}
	for _, word := range remove {
		plexTitle = strings.ReplaceAll(plexTitle, word, "")
		foundTitle = strings.ReplaceAll(foundTitle, word, "")
	}
	// remove colons
	plexTitle = strings.ReplaceAll(plexTitle, ":", "")
	foundTitle = strings.ReplaceAll(foundTitle, ":", "")
	// trim whitespace
	plexTitle = strings.TrimSpace(plexTitle)
	foundTitle = strings.TrimSpace(foundTitle)
	matched := strings.EqualFold(plexTitle, foundTitle)
	return matched
}

func matchTitle(plexTitle, foundTitle string, foundYear, lowerBound, upperBound int) bool {
	// ...existing code...
	plexTitle = strings.ToLower(plexTitle)
	foundTitle = strings.ToLower(foundTitle)
	remove := []string{"the"}
	for _, word := range remove {
		plexTitle = strings.ReplaceAll(plexTitle, word, "")
		foundTitle = strings.ReplaceAll(foundTitle, word, "")
	}
	// remove colons
	plexTitle = strings.ReplaceAll(plexTitle, ":", "")
	foundTitle = strings.ReplaceAll(foundTitle, ":", "")
	// trim whitespace
	plexTitle = strings.TrimSpace(plexTitle)
	foundTitle = strings.TrimSpace(foundTitle)
	matched := strings.EqualFold(plexTitle, foundTitle) &&
		foundYear >= lowerBound && foundYear <= upperBound
	return matched
}

func YearToDate(yearString string) time.Time {
	year, err := strconv.Atoi(yearString)
	if err != nil {
		return time.Time{}
	}
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
}

func SanitizedAlbumTitle(title string) string {
	// convert all unicode to ascii
	title = unidecode.Unidecode(title)
	// remove anything between ()
	r := regexp.MustCompile(`\((.*?)\)`)
	title = r.ReplaceAllString(title, "")
	// remove anything between []
	r = regexp.MustCompile(`\[(.*?)\]`)
	title = r.ReplaceAllString(title, "")
	// remove anything between {}
	r = regexp.MustCompile(`\{(.*?)\}`)
	title = r.ReplaceAllString(title, "")
	// remove plurals ' and ’
	r = regexp.MustCompile(`[',’]`)
	title = r.ReplaceAllString(title, "")
	// strip whitespace
	title = strings.TrimSpace(title)
	// lowercase
	title = strings.ToLower(title)

	return title
}
