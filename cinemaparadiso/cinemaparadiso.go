package cinemaparadiso

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	cinemaparadisoURL = "https://www.cinemaparadiso.co.uk/catalog-w/Search.aspx"
)

type searchResult struct {
	title   string
	url     string
	formats []string
	year    string
}

func SearchCinemaParadiso(title, year string) (hit bool, returnURL string, formats []string) {
	urlEncodedTitle := url.QueryEscape(title)
	rawQuery := []byte(fmt.Sprintf("form-search-field=%s", urlEncodedTitle))
	req, err := http.NewRequestWithContext(context.Background(), "POST", cinemaparadisoURL, bytes.NewBuffer(rawQuery))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Assuming form data

	if err != nil {
		fmt.Println("Error creating request:", err)
		return false, "", nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false, "", nil
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false, "", nil
	}
	rawData := string(body)
	moviesFound := findMoviesInResponse(rawData)
	if len(moviesFound) > 0 {
		return matchTitle(title, year, moviesFound)
	}
	return false, "", nil
}

func findMoviesInResponse(response string) (results []searchResult) {
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

				results = append(results, searchResult{title: foundTitle, year: year, url: returnURL, formats: formats})
			}
			// remove the movie entry from the response
			response = response[endIndex:]
		} else {
			break
		}
	}

	return results
}

func matchTitle(title, year string, results []searchResult) (hit bool, returnURL string, formats []string) {
	expectedYear := yearToDate(year)
	for _, result := range results {
		// normally a match if the year is within 1 year of each other
		resultYear := yearToDate(result.year)
		if result.title == title && (resultYear.Year() == expectedYear.Year() ||
			resultYear.Year() == expectedYear.Year()-1 || resultYear.Year() == expectedYear.Year()+1) {
			return true, result.url, result.formats
		}
	}
	return false, "", nil
}

func yearToDate(yearString string) time.Time {
	year, err := strconv.Atoi(yearString)
	if err != nil {
		return time.Time{}
	}
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
}

func extractMovieFormats(movieEntry string) []string {
	ulStartIndex := strings.Index(movieEntry, `<ul class="media-types">`) + len(`<ul class="media-types">`)
	ulChunk := movieEntry[ulStartIndex:]
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
