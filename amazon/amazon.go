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
	movie.ReleaseDate = time.Time{}
	rawData, err := makeRequest(movie.URL, "")
	if err != nil {
		fmt.Println("scrapeTitle: Error making request:", err)
		ch <- movie
		return
	}
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
	movieSearchResult.PlexMovie = plexMovie
	movieSearchResult.SearchURL = amazonURL

	urlEncodedTitle := url.QueryEscape(plexMovie.Title)
	amazonURL := amazonURL + urlEncodedTitle
	if filter != "" {
		amazonURL += filter
	}
	amazonURL += "&submit=Search&action=search"

	rawData, err := makeRequest(amazonURL, "") // fix the german filter here
	if err != nil {
		return movieSearchResult, err
	}

	moviesFound, _ := findTitlesInResponse(rawData, true)
	movieSearchResult.MovieSearchResults = moviesFound
	movieSearchResult = utils.MarkBestMatch(&movieSearchResult)
	return movieSearchResult, nil
}

func makeRequest(inputURL, country string) (response string, err error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", inputURL, bytes.NewBuffer([]byte{}))

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	switch country {
	case "german":
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

func SearchAmazonTV(plexTVShow *types.PlexTVShow, filter string) (tvSearchResult types.SearchResults, err error) {
	tvSearchResult.PlexTVShow = *plexTVShow
	tvSearchResult.SearchURL = amazonURL

	urlEncodedTitle := url.QueryEscape(fmt.Sprintf("%s complete series", plexTVShow.Title)) // complete series
	amazonURL := amazonURL + urlEncodedTitle
	if filter != "" {
		amazonURL += filter
	}
	amazonURL += "&submit=Search&action=search"
	//
	//fix the filter for german here
	//
	//
	rawData, err := makeRequest(amazonURL, "")
	if err != nil {
		return tvSearchResult, err
	}

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
