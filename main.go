package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
)

type Movie struct {
	Title string
	Year  string
}

func main() {
	// get environment variables
	ipAddress := os.Getenv("PLEX_IP")
	libraryId := os.Getenv("PLEX_LIBRARY_ID")
	plexToken := os.Getenv("PLEX_TOKEN")
	var allMovies []Movie
	allMovies = append(allMovies, getMovies(ipAddress, libraryId, "sd", plexToken)...)
	allMovies = append(allMovies, getMovies(ipAddress, libraryId, "480", plexToken)...)
	allMovies = append(allMovies, getMovies(ipAddress, libraryId, "576", plexToken)...)
	allMovies = append(allMovies, getMovies(ipAddress, libraryId, "720", plexToken)...)

	fmt.Printf("There are a total of %d movies in the library.", len(allMovies))

	// lets search movies in cinemaparadiso
	for _, movie := range allMovies {
		hit, url, formats := searchCinemaParadiso(movie.Title, movie.Year)
		// if hit, and contains any format that isnt dvd, print the movie
		if hit && !slices.Contains(formats, "DVD") {
			fmt.Printf("\n%s %v: %s\n", movie.Title, formats, url)
		}
	}
}

func searchCinemaParadiso(title, year string) (hit bool, returnUrl string, formats []string) {
	cinemaparadisoURL := "https://www.cinemaparadiso.co.uk/catalog-w/Search.aspx"
	urlEncodedTitle := url.QueryEscape(title)
	rawQuery := []byte(fmt.Sprintf("form-search-field=%s", urlEncodedTitle))
	req, err := http.NewRequest("POST", cinemaparadisoURL, bytes.NewBuffer(rawQuery))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Assuming form data

	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	rawData := string(body)
	hit, returnUrl, formats = findMoviesInResponse(title, year, rawData)
	return hit, returnUrl, formats
}

type searchResult struct {
	hit     bool
	url     string
	formats []string
}

func findMoviesInResponse(title, year, response string) (hit bool, url string, formats []string) {
	// look for the movies in the response
	// will be surrounded by <li class="clearfix"> and </li>
	// the url will be in the href attribute of the <a> tag
	// the title will be in the <a> tag

	// Find the start and end index of the movie entry
	//for {
	startIndex := strings.Index(response, "<li class=\"clearfix\">")
	// remove everything before the start index
	response = response[startIndex:]
	endIndex := strings.Index(response, "</li>")

	// If both start and end index are found
	if startIndex != -1 && endIndex != -1 {
		// Extract the movie entry
		movieEntry := response[0:endIndex]

		// Find the URL of the movie
		urlStartIndex := strings.Index(movieEntry, "href=\"") + len("href=\"")
		urlEndIndex := strings.Index(movieEntry[urlStartIndex:], "\"") + urlStartIndex
		url = movieEntry[urlStartIndex:urlEndIndex]

		// Find the title of the movie
		// titleStartIndex := strings.Index(movieEntry, ">") + 1
		// titleEndIndex := strings.Index(movieEntry[titleStartIndex:], "<") + titleStartIndex
		// foundTitle := movieEntry[titleStartIndex:titleEndIndex]
		// fmt.Println("Found title:", foundTitle)
		// Set hit to true
		hit = true

		// Find the formats of the movie
		// get the UL line
		ulStartIndex := strings.Index(movieEntry, `<ul class="media-types">`) + len(`<ul class="media-types">`)
		ulChunk := movieEntry[ulStartIndex:]
		r := regexp.MustCompile(`title="(.*?)"`)

		// Find all matches
		matches := r.FindAllStringSubmatch(ulChunk, -1)

		// Extract and print titles
		for _, match := range matches {
			formats = append(formats, match[1])
		}
	}
	// else {
	// 	// If no more movies are found, break the loop
	// 	break
	// }

	return hit, url, formats
}
