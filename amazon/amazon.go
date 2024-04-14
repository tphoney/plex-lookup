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

func ScrapeMovies(movieSearchResult *types.MovieSearchResults) (scrapedResults []types.SearchResult) {
	var results []types.SearchResult
	for _, searchResult := range movieSearchResult.SearchResults {
		if !searchResult.BestMatch {
			results = append(results, searchResult)
			continue
		}
		scrapedDate, err := scrapeMovie(searchResult.URL)
		if err != nil {
			fmt.Println("Error scraping movie:", err)
		}
		// compare dates
		searchResult.ReleaseDate = scrapedDate
		if scrapedDate.After(movieSearchResult.DateAdded) {
			searchResult.NewRelease = true
		}
		results = append(results, searchResult)
	}
	return results
}

func scrapeMovie(movieURL string) (date time.Time, err error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", movieURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return time.Time{}, err
	}

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return time.Time{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return time.Time{}, err
	}
	rawData := string(body)

	date = findMovieDetails(rawData)
	return date, nil
}

func findMovieDetails(response string) (releaseDate time.Time) {
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

func SearchAmazon(plexMovie types.PlexMovie, filter string) (movieSearchResult types.MovieSearchResults, err error) {
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

	moviesFound := findMoviesInResponse(rawData)
	movieSearchResult.SearchResults = moviesFound
	movieSearchResult = utils.MarkBestMatch(&movieSearchResult)
	return movieSearchResult, nil
}

func findMoviesInResponse(response string) (results []types.SearchResult) {
	// Find the start and end index of the movie entry
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
			// Extract the movie entry
			movieEntry := response[0:endIndex]

			// fmt.Println(movieEntry)
			// Find the URL of the movie
			urlStartIndex := strings.Index(movieEntry, "href=\"") + len("href=\"")
			urlEndIndex := strings.Index(movieEntry[urlStartIndex:], "\"") + urlStartIndex
			returnURL := movieEntry[urlStartIndex:urlEndIndex]
			// Find the title of the movie
			r := regexp.MustCompile(`title="(.*?)\s*\((.*?)\)"`)
			// Find the first match
			match := r.FindStringSubmatch(movieEntry)

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

				results = append(results, types.SearchResult{URL: returnURL, Format: format, Year: year, FoundTitle: foundTitle, UITitle: format})
			}
			// remove the movie entry from the response
			response = response[endIndex:]
		} else {
			break
		}
	}

	return results
}
