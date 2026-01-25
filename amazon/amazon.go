package amazon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/conc/iter"
	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/utils"
)

const (
	amazonURL      = "https://www.blu-ray.com/movies/search.php?keyword="
	LanguageGerman = "german"
)

var (
	numberMoviesProcessed atomic.Int32
	numberTVProcessed     atomic.Int32
	// Regex to match date patterns with abbreviated or full month names
	// Note: May appears in both abbreviated and full month lists, but we don't need it twice
	dateRegex = regexp.MustCompile(`(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec|January|February|March|April|June|July|August|September|October|November|December)\s+(\d{1,2}),\s+(\d{4})`)
	//nolint: mnd
	seasonNumberToInt = map[string]int{
		"one":       1,
		"two":       2,
		"three":     3,
		"four":      4,
		"five":      5,
		"six":       6,
		"seven":     7,
		"eight":     8,
		"nine":      9,
		"ten":       10,
		"eleven":    11,
		"twelve":    12,
		"thirteen":  13,
		"fourteen":  14,
		"fifteen":   15,
		"sixteen":   16,
		"seventeen": 17,
		"eighteen":  18,
		"nineteen":  19,
		"twenty":    20,
	}
	//nolint: mnd
	ordinalNumberToSeason = map[string]int{
		"first season":       1,
		"second season":      2,
		"third season":       3,
		"fourth season":      4,
		"fifth season":       5,
		"sixth season":       6,
		"seventh season":     7,
		"eighth season":      8,
		"ninth season":       9,
		"tenth season":       10,
		"eleventh season":    11,
		"twelfth season":     12,
		"thirteenth season":  13,
		"fourteenth season":  14,
		"fifteenth season":   15,
		"sixteenth season":   16,
		"seventeenth season": 17,
		"eighteenth season":  18,
		"nineteenth season":  19,
		"twentieth season":   20,
	}
)

func MoviesInParallel(plexMovies []types.PlexMovie, language, region string) (searchResults []types.MovieSearchResponse) {
	numberMoviesProcessed.Store(0)
	mapper := iter.Mapper[types.PlexMovie, types.MovieSearchResponse]{
		MaxGoroutines: types.ConcurrencyLimit,
	}
	searchResults = mapper.Map(plexMovies, func(m *types.PlexMovie) types.MovieSearchResponse {
		result := searchMovieValue(m, language, region)
		numberMoviesProcessed.Add(1)
		return result
	})
	numberMoviesProcessed.Store(0) // job is done
	fmt.Println("amazon movies found:", len(searchResults))
	return searchResults
}

func TVInParallel(plexTVShows []types.PlexTVShow, language, region string) (searchResults []types.TVSearchResponse) {
	numberTVProcessed.Store(0)
	mapper := iter.Mapper[types.PlexTVShow, types.TVSearchResponse]{
		MaxGoroutines: types.ConcurrencyLimit,
	}
	searchResults = mapper.Map(plexTVShows, func(tv *types.PlexTVShow) types.TVSearchResponse {
		result := searchTVValue(tv, language, region)
		numberTVProcessed.Add(1)
		return result
	})
	numberTVProcessed.Store(0) // job is done
	fmt.Println("amazon TV shows found:", len(searchResults))
	return searchResults
}

func GetMovieJobProgress() int {
	return int(numberMoviesProcessed.Load())
}

func GetTVJobProgress() int {
	return int(numberTVProcessed.Load())
}

// ScrapeTitlesParallel now only handles TV. Use ScrapeMovieTitlesParallel for movies.
func ScrapeTitlesParallel(searchResults []types.TVSearchResponse, region string) (scrapedResults []types.TVSearchResponse) {
	mapper := iter.Mapper[types.TVSearchResponse, types.TVSearchResponse]{
		MaxGoroutines: types.ConcurrencyLimit,
	}
	scrapedResults = mapper.Map(searchResults, func(sr *types.TVSearchResponse) types.TVSearchResponse {
		return scrapeTVTitlesValue(sr, region)
	})
	fmt.Println("amazon TV titles scraped:", len(scrapedResults))
	return scrapedResults
}

// ScrapeMovieTitlesParallel handles scraping movie titles for MovieSearchResponse.
func ScrapeMovieTitlesParallel(searchResults []types.MovieSearchResponse, region string) (scrapedResults []types.MovieSearchResponse) {
	mapper := iter.Mapper[types.MovieSearchResponse, types.MovieSearchResponse]{
		MaxGoroutines: types.ConcurrencyLimit,
	}
	scrapedResults = mapper.Map(searchResults, func(sr *types.MovieSearchResponse) types.MovieSearchResponse {
		return scrapeMovieTitlesValue(sr, region)
	})
	fmt.Println("amazon movies scraped:", len(scrapedResults))
	return scrapedResults
}

// scrapeMovieTitlesValue is a value-returning version for use with iter.Map
// extractReleaseDate extracts and parses a release date from HTML content
func extractReleaseDate(rawData string) (time.Time, error) {
	// Try multiple patterns to find the release date
	// Pattern 1: Look for date in grey noline link (with or without closing span)
	r1 := regexp.MustCompile(`<a class="grey noline"[^>]*>(.*?)</a>`)
	match := r1.FindStringSubmatch(rawData)
	var stringDate string
	if match != nil {
		stringDate = strings.TrimSpace(match[1])
	}

	// Pattern 2: Look for date pattern anywhere (handles cases where HTML structure differs)
	if stringDate == "" {
		// Match abbreviated month: "Feb 03, 2009" or full month: "February 3, 2009"
		match = dateRegex.FindStringSubmatch(rawData)
		if match != nil {
			// Reconstruct date string in standard format
			month := match[1]
			day := match[2]
			year := match[3]
			// Convert full month names to abbreviations
			monthMap := map[string]string{
				"January": "Jan", "February": "Feb", "March": "Mar", "April": "Apr",
				"May": "May", "June": "Jun", "July": "Jul", "August": "Aug",
				"September": "Sep", "October": "Oct", "November": "Nov", "December": "Dec",
			}
			if abbr, ok := monthMap[month]; ok {
				month = abbr
			}
			// Pad day with zero if needed
			if len(day) == 1 {
				day = "0" + day
			}
			stringDate = fmt.Sprintf("%s %s, %s", month, day, year)
		}
	}

	if stringDate == "" {
		return time.Time{}, fmt.Errorf("no date found in HTML")
	}

	// Try parsing with abbreviated month format first
	releaseDate, parseErr := time.Parse("Jan 02, 2006", stringDate)
	if parseErr != nil {
		// Try alternative format for full month names
		releaseDate, parseErr = time.Parse("January 2, 2006", stringDate)
		if parseErr != nil {
			return time.Time{}, fmt.Errorf("error parsing date '%s': %w", stringDate, parseErr)
		}
	}

	return releaseDate, nil
}

//nolint:dupl // Acceptable duplication - type-specific wrapper for movies
func scrapeMovieTitlesValue(searchResult *types.MovieSearchResponse, region string) types.MovieSearchResponse {
	dateAdded := searchResult.DateAdded
	for i := range searchResult.MovieSearchResults {
		// this is to limit the number of requests
		if !searchResult.MovieSearchResults[i].BestMatch {
			continue
		}
		rawData, err := makeRequest(searchResult.MovieSearchResults[i].URL, region)
		if err != nil {
			fmt.Println("scrapeTitle: Error making request:", err)
			return *searchResult
		}

		// Extract and parse the release date
		releaseDate, err := extractReleaseDate(rawData)
		if err != nil {
			fmt.Printf("scrapeMovieTitles: %v\n", err)
			searchResult.MovieSearchResults[i].ReleaseDate = time.Time{} // default to zero time
		} else {
			searchResult.MovieSearchResults[i].ReleaseDate = releaseDate
		}

		if searchResult.MovieSearchResults[i].ReleaseDate.After(dateAdded) {
			searchResult.MovieSearchResults[i].NewRelease = true
		}
	}
	return *searchResult
}

// scrapeTVTitlesValue is a value-returning version for use with iter.Map
//
//nolint:dupl // Acceptable duplication - type-specific wrapper for TV shows
func scrapeTVTitlesValue(searchResult *types.TVSearchResponse, region string) types.TVSearchResponse {
	dateAdded := searchResult.DateAdded
	for i := range searchResult.TVSearchResults {
		// this is to limit the number of requests
		if !searchResult.TVSearchResults[i].BestMatch {
			continue
		}
		rawData, err := makeRequest(searchResult.TVSearchResults[i].URL, region)
		if err != nil {
			fmt.Println("scrapeTitle: Error making request:", err)
			return *searchResult
		}

		// Extract and parse the release date
		releaseDate, err := extractReleaseDate(rawData)
		if err != nil {
			fmt.Printf("scrapeTVTitles: %v\n", err)
			searchResult.TVSearchResults[i].ReleaseDate = time.Time{} // default to zero time
		} else {
			searchResult.TVSearchResults[i].ReleaseDate = releaseDate
		}

		if searchResult.TVSearchResults[i].ReleaseDate.After(dateAdded) {
			searchResult.TVSearchResults[i].NewRelease = true
		}
	}
	return *searchResult
}

// searchMovieValue is a value-returning version for use with iter.Map
func searchMovieValue(plexMovie *types.PlexMovie, language, region string) types.MovieSearchResponse {
	result := types.MovieSearchResponse{}
	result.PlexMovie = *plexMovie

	urlEncodedTitle := url.QueryEscape(plexMovie.Title)
	searchURL := amazonURL + urlEncodedTitle
	// this searches for the movie in a language
	switch language {
	case LanguageGerman:
		searchURL += "&audio=" + language
	default:
		// do nothing
	}
	searchURL += "&submit=Search&action=search"

	result.SearchURL = searchURL
	rawData, err := makeRequest(searchURL, region)
	if err != nil {
		fmt.Println("searchMovie: Error making request:", err)
		return result
	}

	moviesFound, _ := findTitlesInResponse(rawData, true)
	result.MovieSearchResults = moviesFound
	result = utils.MarkBestMatchMovieResponse(&result)
	return result
}

// searchTVValue is a value-returning version for use with iter.Map
func searchTVValue(plexTVShow *types.PlexTVShow, language, region string) types.TVSearchResponse {
	result := types.TVSearchResponse{}
	result.PlexTVShow = *plexTVShow

	urlEncodedTitle := url.QueryEscape(plexTVShow.Title)
	searchURL := amazonURL + urlEncodedTitle
	// this searches for the TV show in a language
	switch language {
	case LanguageGerman:
		searchURL += "&audio=" + language
	default:
		// do nothing
	}
	searchURL += "&submit=Search&action=search"
	result.SearchURL = searchURL
	rawData, err := makeRequest(searchURL, region)
	if err != nil {
		fmt.Println("searchTV: Error making request:", err)
		return result
	}

	_, titlesFound := findTitlesInResponse(rawData, false)
	// sort the seasons
	sort.Slice(titlesFound, func(i, j int) bool {
		if len(titlesFound[i].Seasons) == 0 || len(titlesFound[j].Seasons) == 0 {
			return false
		}
		return titlesFound[i].Seasons[0].Number < titlesFound[j].Seasons[0].Number
	})
	result.TVSearchResults = titlesFound
	result = utils.MarkBestMatchTVResponse(&result)
	// Count disc formats for UI rendering (MatchesDVD, MatchesBluray, Matches4k)
	var matchesDVD, matchesBluray, matches4k int
	for i := range result.TVSearchResults {
		tvResult := &result.TVSearchResults[i]
		for _, season := range tvResult.Seasons {
			switch season.Format {
			case types.DiskDVD:
				matchesDVD++
			case types.DiskBluray:
				matchesBluray++
			case types.Disk4K:
				matches4k++
			}
		}
	}
	result.MatchesDVD = matchesDVD
	result.MatchesBluray = matchesBluray
	result.Matches4k = matches4k
	return result
}

func findTitlesInResponse(response string, movie bool) (movieResults []types.MovieSearchResult, tvResults []types.TVSearchResult) {
	// Find the start and end index of the entry
	for {
		startIndex := strings.Index(response, `<a class="hoverlink" data-globalproductid=`)
		// remove everything before the start index
		if startIndex == -1 {
			break
		}
		response = response[startIndex:]
		endIndex := strings.Index(response, `</div></div>`)
		// If both start and end index are found
		if endIndex != -1 {
			// Extract the entry
			entry := response[0:endIndex]
			// Find the URL
			urlStartIndex := strings.Index(entry, "href=\"") + len("href=\"")
			urlEndIndex := strings.Index(entry[urlStartIndex:], "\"") + urlStartIndex
			returnURL := entry[urlStartIndex:urlEndIndex]
			// Find the title
			r := regexp.MustCompile(`title="(.*?)\s*\((.*?)\)"`)
			// Find the first match
			match := r.FindStringSubmatch(entry)

			if match != nil {
				// Extract and print title and year
				foundTitle := match[1]
				year := match[2]
				// Find the formats of the movie
				// if the title ends with 4k, then it is 4k
				var format string
				if strings.HasSuffix(foundTitle, " 4K") {
					format = types.Disk4K
					foundTitle = strings.TrimSuffix(foundTitle, " 4K")
				} else {
					foundTitle = strings.TrimSuffix(foundTitle, " Blu-ray")
					format = types.DiskBluray
				}

				if movie {
					movieResults = append(movieResults, types.MovieSearchResult{
						URL: returnURL, Format: format, Year: year, FoundTitle: foundTitle, UITitle: format})
				} else {
					decipheredTitle, number, boxSet, boxSetTitle := decipherTVName(foundTitle)
					// split year
					splitYear := strings.Split(year, "-")
					if len(splitYear) > 0 {
						year = strings.Trim(splitYear[0], " ")
					}
					tvResult := types.TVSearchResult{
						URL: returnURL, Format: []string{format}, FirstAiredYear: year, FoundTitle: decipheredTitle, UITitle: decipheredTitle}
					tvResult.Seasons = append(tvResult.Seasons, types.TVSeasonResult{
						URL: returnURL, Format: format, Number: number, BoxSet: boxSet, BoxSetName: boxSetTitle})
					tvResults = append(tvResults, tvResult)
				}
			}
			// remove the movie entry from the response
			response = response[endIndex:]
		} else {
			break
		}
	}
	return movieResults, tvResults
}

func makeRequest(inputURL, region string) (response string, err error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", inputURL, bytes.NewBuffer([]byte{}))

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	req.Header.Set("Cookie", fmt.Sprintf("country=%s;", region))

	if err != nil {
		fmt.Println("makeRequest: error creating request:", err)
		return response, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("makeRequest: error sending request:", err)
		return response, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("makeRequest: error reading response body:", err)
		return response, err
	}

	// check for a 200 status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("amazon: status code not OK, probably rate limited:", resp.StatusCode)
		return response, fmt.Errorf("amazon: status code not OK: %d", resp.StatusCode)
	}

	rawResponse := string(body)
	return rawResponse, nil
}

func decipherTVName(name string) (title string, number int, boxSet bool, boxSetTitle string) {
	parts := strings.Split(name, ":")
	if len(parts) == 1 {
		title = parts[0]
		return title, -1, false, ""
	}
	// everything after the first colon
	discTitle := strings.Join(parts[len(parts)-1:], "")
	title = strings.TrimSuffix(name, (":" + discTitle))
	boxSetTitle = strings.Trim(discTitle, " ")
	seasonBlock := strings.ToLower(discTitle)
	// complete boxsets
	if strings.Contains(seasonBlock, "complete series") || strings.Contains(seasonBlock, "complete seasons") ||
		strings.Contains(seasonBlock, "complete collection") {
		return title, number, true, boxSetTitle
	}
	// final season
	if strings.Contains(seasonBlock, "final season") {
		return title, 999, false, "" //nolint:mnd
	}
	r := regexp.MustCompile(`seasons?\ (\d+).(\d+)`)
	match := r.FindStringSubmatch(seasonBlock)
	if len(match) > 1 {
		return title, number, true, boxSetTitle
	}
	// does the second part have a number as an integer or as a word.
	r = regexp.MustCompile(`seasons?\ (\d+)`)
	match = r.FindStringSubmatch(seasonBlock)
	if len(match) > 1 {
		number, _ = strconv.Atoi(match[1])
		return title, number, false, boxSetTitle
	}

	for k, v := range seasonNumberToInt {
		if strings.Contains(seasonBlock, ("season "+k)) || strings.Contains(seasonBlock, ("seasons "+k)) {
			return title, v, false, boxSetTitle
		}
	}

	for k, v := range ordinalNumberToSeason {
		if strings.Contains(seasonBlock, k) {
			return title, v, false, ""
		}
	}
	//nolint: gocritic
	// fmt.Printf("warn: decipherTVName, got to the end%q\n", name)
	return title, -1, false, ""
}
