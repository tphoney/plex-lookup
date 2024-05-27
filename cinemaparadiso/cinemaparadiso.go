package cinemaparadiso

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/utils"
)

const (
	cinemaparadisoSearchURL = "https://www.cinemaparadiso.co.uk/catalog-w/Search.aspx"
	cinemaparadisoSeriesURL = "https://www.cinemaparadiso.co.uk/ajax/CPMain.wsFilmDescription,CPMain.ashx?_method=ShowSeries&_session=r"
)

var (
	numberMoviesProcessed int = 0
	numberTVProcessed     int = 0
)

// nolint: dupl, nolintlint
func MoviesInParallel(plexMovies []types.PlexMovie) (searchResults []types.SearchResults) {
	numberMoviesProcessed = 0
	ch := make(chan types.SearchResults, len(plexMovies))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexMovies {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchCinemaParadisoMovie(&plexMovies[i], ch)
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

func ScrapeMoviesParallel(searchResults []types.SearchResults) []types.SearchResults {
	numberMoviesProcessed = 0
	ch := make(chan types.SearchResults, len(searchResults))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range searchResults {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			scrapeMovieTitle(&searchResults[i], ch)
		}(i)
	}
	detailedSearchResults := make([]types.SearchResults, 0, len(searchResults))
	for range searchResults {
		result := <-ch
		detailedSearchResults = append(detailedSearchResults, result)
		numberMoviesProcessed++
	}
	numberMoviesProcessed = 0 // job is done
	return detailedSearchResults
}

// nolint: dupl, nolintlint
func TVInParallel(plexTVShows []types.PlexTVShow) (searchResults []types.SearchResults) {
	ch := make(chan types.SearchResults, len(plexTVShows))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexTVShows {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchTVShow(&plexTVShows[i], ch)
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

func searchCinemaParadisoMovie(plexMovie *types.PlexMovie, movieSearchResult chan<- types.SearchResults) {
	result := types.SearchResults{}
	result.PlexMovie = *plexMovie
	urlEncodedTitle := url.QueryEscape(plexMovie.Title)
	result.SearchURL = cinemaparadisoSearchURL + "?form-search-field=" + urlEncodedTitle
	rawData, err := makeRequest(result.SearchURL, http.MethodPost, fmt.Sprintf("form-search-field=%s", urlEncodedTitle))
	if err != nil {
		fmt.Println("searchCinemaParadisoMovie:", err)
		movieSearchResult <- result
		return
	}

	moviesFound, _ := findTitlesInResponse(rawData, true)
	result.MovieSearchResults = moviesFound
	result = utils.MarkBestMatchMovie(&result)
	movieSearchResult <- result
}

func scrapeMovieTitle(result *types.SearchResults, movieSearchResult chan<- types.SearchResults) {
	// now we can get the season information for each best match
	for i := range result.MovieSearchResults {
		if !result.MovieSearchResults[i].BestMatch {
			continue
		}
		rawData, err := makeRequest(result.MovieSearchResults[i].URL, http.MethodGet, "")
		if err != nil {
			fmt.Println("scrapeMovieTitle:", err)
			movieSearchResult <- *result
			return
		}
		// search for the release date <dt>Release Date:</dt><dd>29/07/2013</dd>
		r := regexp.MustCompile(`<section id="format-(.*?)".*?Release Date:<\/dt><dd>(.*?)<\/dd>`)
		// this will match multiple times for different formats eg DVD, Blu-ray, 4K
		match := r.FindAllStringSubmatch(rawData, -1)
		discReleases := make(map[string]time.Time)
		for i := range match {
			switch match[i][1] {
			case "1":
				discReleases[types.DiskDVD], _ = time.Parse("02/01/2006", match[i][2])
			case "3":
				discReleases[types.DiskBluray], _ = time.Parse("02/01/2006", match[i][2])
			case "14":
				discReleases[types.Disk4K], _ = time.Parse("02/01/2006", match[i][2])
			}
		}
		_, ok := discReleases[result.MovieSearchResults[i].Format]
		if ok {
			result.MovieSearchResults[i].ReleaseDate = discReleases[result.MovieSearchResults[i].Format]
		} else {
			result.MovieSearchResults[i].ReleaseDate = time.Time{}
		}
		// check if the release date is after the date the movie was added to plexs
		if result.MovieSearchResults[i].ReleaseDate.After(result.PlexMovie.DateAdded) {
			result.MovieSearchResults[i].NewRelease = true
		}
	}
	movieSearchResult <- *result
}

func searchTVShow(plexTVShow *types.PlexTVShow, tvSearchResult chan<- types.SearchResults) {
	result := types.SearchResults{}
	urlEncodedTitle := url.QueryEscape(plexTVShow.Title)
	result.PlexTVShow = *plexTVShow
	result.SearchURL = cinemaparadisoSearchURL + "?form-search-field=" + urlEncodedTitle
	rawData, err := makeRequest(result.SearchURL, http.MethodPost, fmt.Sprintf("form-search-field=%s", urlEncodedTitle))
	if err != nil {
		fmt.Println("searchTVShow: Error making web request:", err)
		tvSearchResult <- result
		return
	}

	_, tvFound := findTitlesInResponse(rawData, false)
	result.TVSearchResults = tvFound
	result = utils.MarkBestMatchTV(&result)
	// now we can get the season information for each best match
	for i := range result.TVSearchResults {
		if result.TVSearchResults[i].BestMatch {
			result.TVSearchResults[i].Seasons, _ = findTVSeasonInfo(result.TVSearchResults[i].URL)
		}
	}
	tvSearchResult <- result
}

func findTVSeasonInfo(seriesURL string) (tvSeasons []types.TVSeasonResult, err error) {
	// make a request to the url
	rawData, err := makeRequest(seriesURL, http.MethodGet, "")
	if err != nil {
		fmt.Println("findTVSeasonInfo: Error making web request:", err)
		return tvSeasons, err
	}
	tvSeasons = findTVSeasonsInResponse(rawData)
	return tvSeasons, nil
}

func findTVSeasonsInResponse(response string) (tvSeasons []types.TVSeasonResult) {
	// look for the series in the response
	r := regexp.MustCompile(`<li data-filmId="(\d*)">`)
	match := r.FindAllStringSubmatch(response, -1)
	for i, m := range match {
		tvSeasons = append(tvSeasons, types.TVSeasonResult{Number: i, URL: m[1]})
	}
	// remove the first entry as it is general information
	scrapedTVSeasonResults := make([]types.TVSeasonResult, 0, len(tvSeasons))
	if len(tvSeasons) > 0 {
		tvSeasons = tvSeasons[1:]

		for i := range tvSeasons {
			detailedSeasonResults, err := makeSeasonRequest(tvSeasons[i])
			if err != nil {
				fmt.Println("findTVSeasonsInResponse: Error making season request:", err)
				continue
			}
			scrapedTVSeasonResults = append(scrapedTVSeasonResults, detailedSeasonResults...)
		}
	}
	// sort the results by number
	sort.Slice(scrapedTVSeasonResults, func(i, j int) bool {
		return scrapedTVSeasonResults[i].Number < scrapedTVSeasonResults[j].Number
	})
	return scrapedTVSeasonResults
}

func makeSeasonRequest(tv types.TVSeasonResult) (result []types.TVSeasonResult, err error) {
	rawData, err := makeRequest(cinemaparadisoSeriesURL, http.MethodPost, fmt.Sprintf("FilmID=%s", tv.URL))
	if err != nil {
		return result, fmt.Errorf("makeSeasonRequest: error making request: %w", err)
	}
	// os.WriteFile("series.html", []byte(rawData), 0644)
	// write the raw data to a file
	r := regexp.MustCompile(`{.."Media..":.."(.*?)",.."ReleaseDate..":.."(.*?)"}`)
	// Find all matches
	matches := r.FindAllStringSubmatch(rawData, -1)
	// there will be multiple formats for each season eg https://www.cinemaparadiso.co.uk/rentals/airwolf-171955.html#dvd
	for _, match := range matches {
		newSeason := types.TVSeasonResult{}
		newSeason.Number = tv.Number
		newSeason.Format = strings.ReplaceAll(match[1], "\\", "")
		newSeason.URL = fmt.Sprintf("https://www.cinemaparadiso.co.uk/rentals/%s.html#%s", tv.URL, newSeason.Format)
		// strip slashes from the date
		date := strings.ReplaceAll(match[2], "\\", "")
		var releaseDate time.Time
		releaseDate, err = time.Parse("02/01/2006", date)
		if err != nil {
			releaseDate = time.Time{}
		}
		newSeason.ReleaseDate = releaseDate
		result = append(result, newSeason)
	}
	return result, nil
}

func findTitlesInResponse(response string, movie bool) (movieResults []types.MovieSearchResult, tvResults []types.TVSearchResult) {
	// look for the movies in the response
	// will be surrounded by <li class="clearfix"> and </li>
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
			formats := extractDiscFormats(movieEntry)

			// Find the title of the movie
			r := regexp.MustCompile(`<a.*?>(.*?)\s*\((.*?)\)</a>`)

			// Find the first match
			match := r.FindStringSubmatch(movieEntry)

			if match != nil {
				// Extract and print title and year
				foundTitle := match[1]
				year := match[2]
				splitYear := strings.Split(year, "-")
				year = strings.Trim(splitYear[0], " ")
				if movie {
					for _, format := range formats {
						movieResults = append(movieResults, types.MovieSearchResult{
							URL: returnURL, Format: format, Year: year, FoundTitle: foundTitle, UITitle: format})
					}
				} else {
					tvResults = append(tvResults, types.TVSearchResult{
						URL: returnURL, Format: formats, FirstAiredYear: year, FoundTitle: foundTitle, UITitle: foundTitle})
				}
			}
			// remove the entry from the response
			response = response[endIndex:]
		} else {
			break
		}
	}

	return movieResults, tvResults
}

func makeRequest(urlEncodedTitle, method, content string) (rawResponse string, err error) {
	var req *http.Request
	switch method {
	case http.MethodPost:
		req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, urlEncodedTitle, bytes.NewBuffer([]byte(content)))
		if strings.Contains(content, "form-search-field") {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Assuming form data
		}
	case http.MethodGet:
		req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, urlEncodedTitle, http.NoBody)
	}

	if err != nil {
		fmt.Println("Error creating request:", err)
		return rawResponse, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return rawResponse, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return rawResponse, err
	}
	rawData := string(body)
	// write the raw data to a file
	// os.WriteFile("search.html", body, 0644)
	return rawData, nil
}

func extractDiscFormats(movieEntry string) []string {
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
