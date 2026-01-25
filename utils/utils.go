package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rainycape/unidecode"
	"github.com/tphoney/plex-lookup/types"
)

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

func MarkBestMatchTV(search *types.SearchResult) types.SearchResult {
	firstEpisodeBoundry := search.FirstEpisodeAired.Year() - 1
	lastEpisodeBoundry := search.LastEpisodeAired.Year() + 1
	fmt.Printf("[DEBUG] MarkBestMatchTV: PlexTitle='%s', FirstAired=%d, LastAired=%d\n", search.Title, search.FirstEpisodeAired.Year(), search.LastEpisodeAired.Year())
	for i := range search.TVSearchResults {
		resultYear := YearToDate(search.TVSearchResults[i].FirstAiredYear)
		fmt.Printf("[DEBUG]   Candidate: FoundTitle='%s', FirstAiredYear='%s', ParsedYear=%d\n", search.TVSearchResults[i].FoundTitle, search.TVSearchResults[i].FirstAiredYear, resultYear.Year())
		// If year is empty, match on title only (skip year check)
		if search.TVSearchResults[i].FirstAiredYear == "" {
			if matchTitleNoYear(search.Title, search.TVSearchResults[i].FoundTitle) {
				fmt.Printf("[DEBUG]     matchTitleNoYear: MATCHED\n")
				search.TVSearchResults[i].BestMatch = true
			} else {
				fmt.Printf("[DEBUG]     matchTitleNoYear: NOT MATCHED\n")
			}
		} else {
			// Original logic for shows with years
			if matchTitle(search.Title, search.TVSearchResults[i].FoundTitle,
				resultYear.Year(), firstEpisodeBoundry, lastEpisodeBoundry) {
				fmt.Printf("[DEBUG]     matchTitle: MATCHED (bounds %d-%d)\n", firstEpisodeBoundry, lastEpisodeBoundry)
				search.TVSearchResults[i].BestMatch = true
			} else {
				fmt.Printf("[DEBUG]     matchTitle: NOT MATCHED (bounds %d-%d)\n", firstEpisodeBoundry, lastEpisodeBoundry)
			}
		}
	}
	return *search
}

func matchTitleNoYear(plexTitle, foundTitle string) bool {
	origPlexTitle := plexTitle
	origFoundTitle := foundTitle
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
	fmt.Printf("[DEBUG]     matchTitleNoYear: '%s' vs '%s' => '%s' vs '%s' => %v\n", origPlexTitle, origFoundTitle, plexTitle, foundTitle, matched)
	return matched
}

func matchTitle(plexTitle, foundTitle string, foundYear, lowerBound, upperBound int) bool {
	origPlexTitle := plexTitle
	origFoundTitle := foundTitle
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
	fmt.Printf("[DEBUG]     matchTitle: '%s' vs '%s' => '%s' vs '%s', year=%d, bounds=%d-%d => %v\n", origPlexTitle, origFoundTitle, plexTitle, foundTitle, foundYear, lowerBound, upperBound, matched)
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
