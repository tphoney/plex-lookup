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
	amazonURL = "https://www.blu-ray.com/movies/search.php?keyword="
)

func ScrapeTitles(searchResults *types.SearchResults) (scrapedResults []types.MovieSearchResult) {
	var results, lookups []types.MovieSearchResult
	for _, searchResult := range searchResults.MovieSearchResults {
		if !searchResult.BestMatch {
			results = append(results, searchResult)
		} else {
			lookups = append(lookups, searchResult)
		}
	}

	if len(lookups) > 0 {
		ch := make(chan *types.MovieSearchResult, len(lookups))
		// Limit number of concurrent requests
		semaphore := make(chan struct{}, types.ConcurrencyLimit)
		for i := range lookups {
			go func() {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				scrapeTitle(&lookups[i], searchResults.PlexMovie.DateAdded, ch)
			}()
		}

		for i := 0; i < len(lookups); i++ {
			lookup := <-ch
			results = append(results, *lookup)
		}
	}
	return results
}

func scrapeTitle(movie *types.MovieSearchResult, dateAdded time.Time, ch chan<- *types.MovieSearchResult) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", movie.URL, bytes.NewBuffer([]byte{}))
	movie.ReleaseDate = time.Time{}
	if err != nil {
		fmt.Println("Error creating request:", err)
		ch <- movie
		return
	}

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		ch <- movie
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		ch <- movie
		return
	}
	rawData := string(body)
	movie.ReleaseDate = findTitleDetails(rawData)
	if movie.ReleaseDate.After(dateAdded) {
		movie.NewRelease = true
	}
	ch <- movie
}

func findTitleDetails(response string) (releaseDate time.Time) {
	r := regexp.MustCompile(`<a class="grey noline" alt=".*">(.*?)</a></span>`)

	match := r.FindStringSubmatch(response)
	if match != nil {
		stringDate := match[1]
		var err error
		releaseDate, err = time.Parse("Jan 02, 2006", stringDate)
		if err != nil {
			releaseDate = time.Time{}
		}
	} else {
		releaseDate = time.Time{}
	}

	return releaseDate
}

func SearchAmazonMovie(plexMovie types.PlexMovie, filter string) (movieSearchResult types.SearchResults, err error) {
	urlEncodedTitle := url.QueryEscape(plexMovie.Title)
	amazonURL := amazonURL + urlEncodedTitle
	if filter != "" {
		amazonURL += filter
	}
	amazonURL += "&submit=Search&action=search"
	req, err := http.NewRequestWithContext(context.Background(), "GET", amazonURL, bytes.NewBuffer([]byte{}))

	movieSearchResult.PlexMovie = plexMovie
	movieSearchResult.SearchURL = amazonURL

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	country := "uk"
	if strings.Contains(filter, "german") {
		country = "de"
	}
	req.Header.Set("Cookie", fmt.Sprintf("country=%s;", country))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return movieSearchResult, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return movieSearchResult, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return movieSearchResult, err
	}
	rawData := string(body)

	moviesFound, _ := findTitlesInResponse(rawData, true)
	movieSearchResult.MovieSearchResults = moviesFound
	movieSearchResult = utils.MarkBestMatch(&movieSearchResult)
	return movieSearchResult, nil
}

func SearchAmazonTV(plexTVShow *types.PlexTVShow, filter string) (tvSearchResult types.SearchResults, err error) {
	urlEncodedTitle := url.QueryEscape(fmt.Sprintf("%s complete series", plexTVShow.Title)) // complete series
	amazonURL := amazonURL + urlEncodedTitle
	if filter != "" {
		amazonURL += filter
	}
	amazonURL += "&submit=Search&action=search"
	req, err := http.NewRequestWithContext(context.Background(), "GET", amazonURL, bytes.NewBuffer([]byte{}))

	tvSearchResult.PlexTVShow = *plexTVShow
	tvSearchResult.SearchURL = amazonURL

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	country := "uk"
	if strings.Contains(filter, "german") {
		country = "de"
	}
	req.Header.Set("Cookie", fmt.Sprintf("country=%s;", country))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return tvSearchResult, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return tvSearchResult, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return tvSearchResult, err
	}
	rawData := string(body)

	_, titlesFound := findTitlesInResponse(rawData, false)
	tvSearchResult.TVSearchResults = titlesFound
	tvSearchResult = utils.MarkBestMatch(&tvSearchResult)
	return tvSearchResult, nil
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
