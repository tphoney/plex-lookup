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
)

func GetCinemaParadisoMoviesInParallel(plexMovies []types.PlexMovie) (searchResults []types.SearchResults) {
	ch := make(chan types.SearchResults, len(plexMovies))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range plexMovies {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			SearchCinemaParadisoMovie(plexMovies[i], ch)
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

func GetJobProgress() int {
	return numberMoviesProcessed
}

func SearchCinemaParadisoMovie(plexMovie types.PlexMovie, movieSearchResult chan<- types.SearchResults) {
	result := types.SearchResults{}
	urlEncodedTitle := url.QueryEscape(plexMovie.Title)
	result.PlexMovie = plexMovie
	result.SearchURL = cinemaparadisoSearchURL + "?form-search-field=" + urlEncodedTitle
	rawData, err := makeSearchRequest(urlEncodedTitle)
	if err != nil {
		fmt.Println("Error making web request:", err)
		movieSearchResult <- result
		return
	}

	moviesFound, _ := findTitlesInResponse(rawData, true)
	result.MovieSearchResults = moviesFound
	result = utils.MarkBestMatch(&result)
	movieSearchResult <- result
}

func SearchCinemaParadisoTV(plexTVShow *types.PlexTVShow) (tvSearchResult types.SearchResults, err error) {
	urlEncodedTitle := url.QueryEscape(plexTVShow.Title)
	tvSearchResult.PlexTVShow = *plexTVShow
	tvSearchResult.SearchURL = cinemaparadisoSearchURL + "?form-search-field=" + urlEncodedTitle
	rawData, err := makeSearchRequest(urlEncodedTitle)
	if err != nil {
		fmt.Println("Error making web request:", err)
		return tvSearchResult, err
	}

	_, tvFound := findTitlesInResponse(rawData, false)
	tvSearchResult.TVSearchResults = tvFound
	tvSearchResult = utils.MarkBestMatch(&tvSearchResult)
	// now we can get the series information for each best match
	for i := range tvSearchResult.TVSearchResults {
		if tvSearchResult.TVSearchResults[i].BestMatch {
			tvSearchResult.TVSearchResults[i].Seasons, _ = findTVSeriesInfo(tvSearchResult.TVSearchResults[i].URL)
		}
	}
	return tvSearchResult, nil
}

func findTVSeriesInfo(seriesURL string) (tvSeries []types.TVSeasonResult, err error) {
	// make a request to the url
	req, err := http.NewRequestWithContext(context.Background(), "GET", seriesURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return tvSeries, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return tvSeries, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return tvSeries, err
	}
	rawData := string(body)
	// write the raw data to a file
	// os.WriteFile("series.html", body, 0644)
	tvSeries = findTVSeriesInResponse(rawData)
	return tvSeries, nil
}

func makeSearchRequest(urlEncodedTitle string) (rawResponse string, err error) {
	rawQuery := []byte(fmt.Sprintf("form-search-field=%s", urlEncodedTitle))
	req, err := http.NewRequestWithContext(context.Background(), "POST", cinemaparadisoSearchURL, bytes.NewBuffer(rawQuery))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Assuming form data

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

func findTVSeriesInResponse(response string) (tvSeries []types.TVSeasonResult) {
	// look for the series in the response
	r := regexp.MustCompile(`<li data-filmId="(\d*)">`)
	match := r.FindAllStringSubmatch(response, -1)
	for i, m := range match {
		tvSeries = append(tvSeries, types.TVSeasonResult{Number: i, URL: m[1]})
	}
	// remove the first entry as it is general information
	results := make([]types.TVSeasonResult, 0, len(tvSeries))
	if len(tvSeries) > 0 {
		tvSeries = tvSeries[1:]
		ch := make(chan *types.TVSeasonResult, len(tvSeries))

		semaphore := make(chan struct{}, types.ConcurrencyLimit)
		for i := range tvSeries {
			go func() {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				makeSeriesRequest(tvSeries[i], ch)
			}()
		}

		for i := 0; i < len(tvSeries); i++ {
			result := <-ch
			results = append(results, *result)
		}
	}
	// sort the results by number
	sort.Slice(results, func(i, j int) bool {
		return results[i].Number < results[j].Number
	})
	return results
}

func makeSeriesRequest(tv types.TVSeasonResult, ch chan<- *types.TVSeasonResult) {
	content := []byte(fmt.Sprintf("FilmID=%s", tv.URL))
	req, err := http.NewRequestWithContext(context.Background(), "POST", cinemaparadisoSeriesURL, bytes.NewBuffer(content))
	if err != nil {
		fmt.Println("Error creating request:", err)
		ch <- &tv
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		ch <- &tv
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		ch <- &tv
		return
	}
	rawData := string(body)
	// write the raw data to a file
	r := regexp.MustCompile(`{.."Media..":.."(.*?)",.."ReleaseDate..":.."(.*?)"}`)

	// Find all matches
	matches := r.FindAllStringSubmatch(rawData, -1)
	for _, match := range matches {
		tv.Format = append(tv.Format, strings.ReplaceAll(match[1], "\\", ""))
		// strip slashes from the date
		date := strings.ReplaceAll(match[2], "\\", "")
		var releaseDate time.Time
		releaseDate, err = time.Parse("02/01/2006", date)
		if err != nil {
			releaseDate = time.Time{}
		}
		tv.ReleaseDate = releaseDate
	}
	ch <- &tv
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
				if movie {
					for _, format := range formats {
						movieResults = append(movieResults, types.MovieSearchResult{
							URL: returnURL, Format: format, Year: year, FoundTitle: foundTitle, UITitle: format})
					}
				} else {
					tvResults = append(tvResults, types.TVSearchResult{
						URL: returnURL, Format: formats, Year: year, FoundTitle: foundTitle, UITitle: foundTitle})
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
