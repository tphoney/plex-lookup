package cinemaparadiso

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/utils"
)

const (
	cinemaparadisoURL = "https://www.cinemaparadiso.co.uk/catalog-w/Search.aspx"
)

func SearchCinemaParadiso(title, year string) (movieSearchResult types.MovieSearchResults, err error) {
	urlEncodedTitle := url.QueryEscape(title)
	rawQuery := []byte(fmt.Sprintf("form-search-field=%s", urlEncodedTitle))
	req, err := http.NewRequestWithContext(context.Background(), "POST", cinemaparadisoURL, bytes.NewBuffer(rawQuery))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Assuming form data

	movieSearchResult.Title = title
	movieSearchResult.Year = year
	movieSearchResult.SearchURL = cinemaparadisoURL + "?form-search-field=" + urlEncodedTitle

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
	// write the raw data to a file
	// os.WriteFile("cinemaparadiso.html", body, 0644)

	moviesFound := findMoviesInResponse(rawData)
	movieSearchResult.SearchResults = moviesFound
	movieSearchResult = utils.MarkBestMatch(&movieSearchResult)

	return movieSearchResult, nil
}

func findMoviesInResponse(response string) (results []types.SearchResult) {
	// look for the movies in the response
	// will be surrounded by <li class="clearfix"> and </li>
	// the url will be in the href attribute of the <a> tag
	// the title will be in the <a> tag

	// Find the start and end index of the movie entry
	for {
		startIndex := strings.Index(response, "<li class=\"clearfix\">")
		// remove everything before the start index
		if startIndex == -1 {
			break
		}
		response = response[startIndex:]
		endIndex := strings.Index(response, "</a></div>")

		// If both start and end index are found
		if endIndex != -1 {
			// Extract the movie entry
			movieEntry := response[0:endIndex]

			// Find the URL of the movie
			urlStartIndex := strings.Index(movieEntry, "href=\"") + len("href=\"")
			urlEndIndex := strings.Index(movieEntry[urlStartIndex:], "\"") + urlStartIndex
			returnURL := movieEntry[urlStartIndex:urlEndIndex]

			// Find the formats of the movies
			formats := extractMovieFormats(movieEntry)

			// Find the title of the movie
			r := regexp.MustCompile(`<a.*?>(.*?)\s*\((.*?)\)</a>`)

			// Find the first match
			match := r.FindStringSubmatch(movieEntry)

			if match != nil {
				// Extract and print title and year
				foundTitle := match[1]
				year := match[2]
				for _, format := range formats {
					results = append(results, types.SearchResult{URL: returnURL, Format: format, Year: year, FoundTitle: foundTitle, UITitle: format})
				}
			}
			// remove the movie entry from the response
			response = response[endIndex:]
		} else {
			break
		}
	}

	return results
}

func extractMovieFormats(movieEntry string) []string {
	ulStartIndex := strings.Index(movieEntry, `<ul class="media-types">`) + len(`<ul class="media-types">`)
	ulEndIndex := strings.Index(movieEntry[ulStartIndex:], "</ul>") + ulStartIndex
	ulChunk := movieEntry[ulStartIndex:ulEndIndex]
	r := regexp.MustCompile(`title="(.*?)"`)

	// Find all matches
	matches := r.FindAllStringSubmatch(ulChunk, -1)

	// Extract and return titles
	var formats []string
	for _, match := range matches {
		formats = append(formats, strings.TrimSpace(match[1]))
	}
	return formats
}
