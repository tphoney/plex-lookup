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

func SearchAmazonMoviesInParallel(plexMovies []types.PlexMovie, language string) (searchResults []types.SearchResults) {
	ch := make(chan types.SearchResults, len(plexMovies))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexMovies {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchAmazonMovie(plexMovies[i], language, ch)
		}(i)
	}

	searchResults = make([]types.SearchResults, 0, len(plexMovies))
	for range plexMovies {
		result := <-ch
		searchResults = append(searchResults, result)
		numberMoviesProcessed++
	}
	numberMoviesProcessed = 0 // job is done
	return searchResults
}

func SearchAmazonTVInParallel(plexTVShows []types.PlexTVShow, language string) (searchResults []types.SearchResults) {
	ch := make(chan types.SearchResults, len(plexTVShows))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexTVShows {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchAmazonTV(&plexTVShows[i], language, ch)
		}(i)
	}

	searchResults = make([]types.SearchResults, 0, len(plexTVShows))
	for range plexTVShows {
		result := <-ch
		searchResults = append(searchResults, result)
		numberTVProcessed++
	}
	numberTVProcessed = 0 // job is done
	return searchResults
}

func GetMovieJobProgress() int {
	return numberMoviesProcessed
}

func GetTVJobProgress() int {
	return numberTVProcessed
}

func ScrapeTitlesParallel(searchResults []types.SearchResults) (scrapedResults []types.SearchResults) {
	numberMoviesProcessed = 0

	for i := range searchResults {
		// check if the search result is a movie
		if len(searchResults[i].MovieSearchResults) > 0 {
			ch := make(chan *types.MovieSearchResult, len(searchResults[i].MovieSearchResults))
			semaphore := make(chan struct{}, types.ConcurrencyLimit)
			for j := range searchResults[i].MovieSearchResults {
				go func(j int) {
					semaphore <- struct{}{}
					defer func() { <-semaphore }()
					scrapeTitle(&searchResults[i].MovieSearchResults[j], searchResults[i].PlexMovie.DateAdded, ch)
				}(j)
			}
			movieResults := make([]types.MovieSearchResult, 0, len(searchResults[i].MovieSearchResults))
			for range searchResults[i].MovieSearchResults {
				result := <-ch
				movieResults = append(movieResults, *result)
			}
			fmt.Println("Scraped", len(movieResults), "titles for", searchResults[i].PlexMovie.Title)
			searchResults[i].MovieSearchResults = movieResults
		}
		scrapedResults = append(scrapedResults, searchResults[i])
		numberMoviesProcessed++
	}
	return scrapedResults
}

func scrapeTitle(movie *types.MovieSearchResult, dateAdded time.Time, ch chan<- *types.MovieSearchResult) {
	rawData, err := makeRequest(movie.URL, "")
	if err != nil {
		fmt.Println("scrapeTitle: Error making request:", err)
		ch <- movie
		return
	}
	// Find the release date
	movie.ReleaseDate = time.Time{} // default to zero time
	r := regexp.MustCompile(`<a class="grey noline" alt=".*">(.*?)</a></span>`)
	match := r.FindStringSubmatch(rawData)
	if match != nil {
		stringDate := match[1]
		movie.ReleaseDate, _ = time.Parse("Jan 02, 2006", stringDate)
	}
	if movie.ReleaseDate.After(dateAdded) {
		movie.NewRelease = true
	}
	ch <- movie
}

func searchAmazonMovie(plexMovie types.PlexMovie, language string, movieSearchResult chan<- types.SearchResults) {
	result := types.SearchResults{}
	result.PlexMovie = plexMovie
	result.SearchURL = ""

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

	rawData, err := makeRequest(amazonURL, language)
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

func searchAmazonTV(plexTVShow *types.PlexTVShow, language string, tvSearchResult chan<- types.SearchResults) {
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

	rawData, err := makeRequest(amazonURL, language)
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

func makeRequest(inputURL, language string) (response string, err error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", inputURL, bytes.NewBuffer([]byte{}))

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// this forces results from a specific amazon region
	switch language {
	case LanguageGerman:
		req.Header.Set("Cookie", "country=de;")
	default:
		req.Header.Set("Cookie", "country=uk;")
	}

	if err != nil {
		fmt.Println("Error creating request:", err)
		return response, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return response, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return response, err
	}
	rawResponse := string(body)
	return rawResponse, nil
}
