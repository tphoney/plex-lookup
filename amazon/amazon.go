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
	"time"

	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/utils"
)

const (
	amazonURL      = "https://www.blu-ray.com/movies/search.php?keyword="
	LanguageGerman = "german"
)

var (
	numberMoviesProcessed int = 0
	numberTVProcessed     int = 0
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

// nolint: dupl, nolintlint
func MoviesInParallel(plexMovies []types.PlexMovie, language, region string) (searchResults []types.SearchResult) {
	numberMoviesProcessed = 0
	ch := make(chan types.SearchResult, len(plexMovies))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexMovies {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchMovie(&plexMovies[i], language, region, ch)
		}(i)
	}

	searchResults = make([]types.SearchResult, 0, len(plexMovies))
	for range plexMovies {
		result := <-ch
		searchResults = append(searchResults, result)
		numberMoviesProcessed++
	}
	numberMoviesProcessed = 0 // job is done
	fmt.Println("amazon movies found:", len(searchResults))
	return searchResults
}

// nolint: dupl, nolintlint
func TVInParallel(plexTVShows []types.PlexTVShow, language, region string) (searchResults []types.SearchResult) {
	numberMoviesProcessed = 0
	ch := make(chan types.SearchResult, len(plexTVShows))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexTVShows {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchTV(&plexTVShows[i], language, region, ch)
		}(i)
	}

	searchResults = make([]types.SearchResult, 0, len(plexTVShows))
	for range plexTVShows {
		result := <-ch
		searchResults = append(searchResults, result)
		numberTVProcessed++
	}
	numberTVProcessed = 0 // job is done
	fmt.Println("amazon TV shows found:", len(searchResults))
	return searchResults
}

func GetMovieJobProgress() int {
	return numberMoviesProcessed
}

func GetTVJobProgress() int {
	return numberTVProcessed
}

func ScrapeTitlesParallel(searchResults []types.SearchResult, region string, isTV bool) (scrapedResults []types.SearchResult) {
	// are we tv or movie
	if isTV {
		numberTVProcessed = 0
	} else {
		numberMoviesProcessed = 0
	}
	ch := make(chan types.SearchResult, len(searchResults))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)
	for i := range searchResults {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			if isTV {
				scrapeTVTitles(&searchResults[i], region, ch)
			} else {
				scrapeMovieTitles(&searchResults[i], region, ch)
			}
		}(i)
	}

	scrapedResults = make([]types.SearchResult, 0, len(searchResults))
	for range searchResults {
		result := <-ch
		scrapedResults = append(scrapedResults, result)
		if isTV {
			numberTVProcessed++
		} else {
			numberMoviesProcessed++
		}
	}
	if isTV {
		numberTVProcessed = 0
	} else {
		numberMoviesProcessed = 0
	}
	fmt.Println("amazon titles scraped:", len(scrapedResults))
	return scrapedResults
}

// nolint: dupl, nolintlint
func scrapeMovieTitles(searchResult *types.SearchResult, region string, ch chan<- types.SearchResult) {
	dateAdded := searchResult.PlexMovie.DateAdded
	for i := range searchResult.MovieSearchResults {
		// this is to limit the number of requests
		if !searchResult.MovieSearchResults[i].BestMatch {
			continue
		}
		rawData, err := makeRequest(searchResult.MovieSearchResults[i].URL, region)
		if err != nil {
			fmt.Println("scrapeTitle: Error making request:", err)
			ch <- *searchResult
			return
		}
		// Find the release date
		searchResult.MovieSearchResults[i].ReleaseDate = time.Time{} // default to zero time
		r := regexp.MustCompile(`<a class="grey noline" alt=".*">(.*?)</a></span>`)
		match := r.FindStringSubmatch(rawData)
		if match != nil {
			stringDate := match[1]
			searchResult.MovieSearchResults[i].ReleaseDate, _ = time.Parse("Jan 02, 2006", stringDate)
		}
		if searchResult.MovieSearchResults[i].ReleaseDate.After(dateAdded) {
			searchResult.MovieSearchResults[i].NewRelease = true
		}
	}
	ch <- *searchResult
}

// nolint: dupl, nolintlint
func scrapeTVTitles(searchResult *types.SearchResult, region string, ch chan<- types.SearchResult) {
	dateAdded := searchResult.PlexTVShow.DateAdded
	for i := range searchResult.TVSearchResults {
		// this is to limit the number of requests
		if !searchResult.TVSearchResults[i].BestMatch {
			continue
		}
		rawData, err := makeRequest(searchResult.TVSearchResults[i].URL, region)
		if err != nil {
			fmt.Println("scrapeTitle: Error making request:", err)
			ch <- *searchResult
			return
		}
		// Find the release date
		searchResult.TVSearchResults[i].ReleaseDate = time.Time{} // default to zero time
		r := regexp.MustCompile(`<a class="grey noline" alt=".*">(.*?)</a></span>`)
		match := r.FindStringSubmatch(rawData)
		if match != nil {
			stringDate := match[1]
			searchResult.TVSearchResults[i].ReleaseDate, _ = time.Parse("Jan 02, 2006", stringDate)
		}
		if searchResult.TVSearchResults[i].ReleaseDate.After(dateAdded) {
			searchResult.TVSearchResults[i].NewRelease = true
		}
	}
	ch <- *searchResult
}

// nolint: dupl, nolintlint
func searchMovie(plexMovie *types.PlexMovie, language, region string, movieSearchResult chan<- types.SearchResult) {
	result := types.SearchResult{}
	result.PlexMovie = *plexMovie

	urlEncodedTitle := url.QueryEscape(plexMovie.Title)
	amazonURL := amazonURL + urlEncodedTitle
	// this searches for the movie in a language
	switch language {
	case LanguageGerman:
		amazonURL += "&audio=" + language
	default:
		// do nothing
	}
	amazonURL += "&submit=Search&action=search"

	result.SearchURL = amazonURL
	rawData, err := makeRequest(amazonURL, region)
	if err != nil {
		fmt.Println("searchMovie: Error making request:", err)
		movieSearchResult <- result
		return
	}

	moviesFound, _ := findTitlesInResponse(rawData, true)
	result.MovieSearchResults = moviesFound
	result = utils.MarkBestMatchMovie(&result)
	movieSearchResult <- result
}

// nolint: dupl, nolintlint
func searchTV(plexTVShow *types.PlexTVShow, language, region string, tvSearchResult chan<- types.SearchResult) {
	result := types.SearchResult{}
	result.PlexTVShow = *plexTVShow

	urlEncodedTitle := url.QueryEscape(plexTVShow.Title)
	amazonURL := amazonURL + urlEncodedTitle
	// this searches for the movie in a language
	switch language {
	case LanguageGerman:
		amazonURL += "&audio=" + language
	default:
		// do nothing
	}
	amazonURL += "&submit=Search&action=search"
	result.SearchURL = amazonURL
	rawData, err := makeRequest(amazonURL, region)
	if err != nil {
		fmt.Println("searchTV: Error making request:", err)
		tvSearchResult <- result
		return
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
	result = utils.MarkBestMatchTV(&result)
	tvSearchResult <- result
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
					year = strings.Trim(splitYear[0], " ")
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
