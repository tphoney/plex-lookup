package amazon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
)

// nolint: dupl, nolintlint
func SearchAmazonMoviesInParallel(plexMovies []types.PlexMovie, language, region string) (searchResults []types.SearchResults) {
	numberMoviesProcessed = 0
	ch := make(chan types.SearchResults, len(plexMovies))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexMovies {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchAmazonMovie(&plexMovies[i], language, region, ch)
		}(i)
	}

	searchResults = make([]types.SearchResults, 0, len(plexMovies))
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
func SearchAmazonTVInParallel(plexTVShows []types.PlexTVShow, language, region string) (searchResults []types.SearchResults) {
	numberMoviesProcessed = 0
	ch := make(chan types.SearchResults, len(plexTVShows))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexTVShows {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchAmazonTV(&plexTVShows[i], language, region, ch)
		}(i)
	}

	searchResults = make([]types.SearchResults, 0, len(plexTVShows))
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

func ScrapeTitlesParallel(searchResults []types.SearchResults, region string) (scrapedResults []types.SearchResults) {
	numberMoviesProcessed = 0
	ch := make(chan types.SearchResults, len(searchResults))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)
	for i := range searchResults {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			scrapeTitles(&searchResults[i], region, ch)
		}(i)
	}

	scrapedResults = make([]types.SearchResults, 0, len(searchResults))
	for range searchResults {
		result := <-ch
		scrapedResults = append(scrapedResults, result)
		numberMoviesProcessed++
	}
	numberMoviesProcessed = 0
	fmt.Println("amazon Movie titles scraped:", len(scrapedResults))
	return scrapedResults
}

func scrapeTitles(searchResult *types.SearchResults, region string, ch chan<- types.SearchResults) {
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

func searchAmazonMovie(plexMovie *types.PlexMovie, language, region string, movieSearchResult chan<- types.SearchResults) {
	result := types.SearchResults{}
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
		fmt.Println("searchAmazonMovie: Error making request:", err)
		movieSearchResult <- result
		return
	}

	moviesFound, _ := findTitlesInResponse(rawData, true)
	result.MovieSearchResults = moviesFound
	result = utils.MarkBestMatch(&result)
	movieSearchResult <- result
}

func searchAmazonTV(plexTVShow *types.PlexTVShow, language, region string, tvSearchResult chan<- types.SearchResults) {
	result := types.SearchResults{}
	result.PlexTVShow = *plexTVShow
	result.SearchURL = amazonURL

	urlEncodedTitle := url.QueryEscape(fmt.Sprintf("%s complete series", plexTVShow.Title)) // complete series
	amazonURL := amazonURL + urlEncodedTitle
	// this searches for the movie in a language
	switch language {
	case LanguageGerman:
		amazonURL += "&audio=" + language
	default:
		// do nothing
	}
	amazonURL += "&submit=Search&action=search"

	rawData, err := makeRequest(amazonURL, region)
	if err != nil {
		fmt.Println("searchAmazonTV: Error making request:", err)
		tvSearchResult <- result
		return
	}

	_, titlesFound := findTitlesInResponse(rawData, false)
	result.TVSearchResults = titlesFound
	result = utils.MarkBestMatch(&result)
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
					boxSet := false
					if strings.Contains(foundTitle, ": The Complete Series") {
						foundTitle = strings.TrimSuffix(foundTitle, ": The Complete Series")
						boxSet = true
					}
					tvResults = append(tvResults, types.TVSearchResult{
						URL: returnURL, Format: []string{format}, Year: year, FoundTitle: foundTitle, UITitle: foundTitle, BoxSet: boxSet})
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
