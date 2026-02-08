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
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/conc/iter"
	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/utils"
)

const (
	cinemaparadisoSearchURL = "https://www.cinemaparadiso.co.uk/catalog-w/Search.aspx"
	cinemaparadisoSeriesURL = "https://www.cinemaparadiso.co.uk/ajax/CPMain.wsFilmDescription,CPMain.ashx?_method=ShowSeries&_session=r"
)

var (
	numberMoviesProcessed atomic.Int32
	numberTVProcessed     atomic.Int32
)

// nolint: dupl, nolintlint
func MoviesInParallel(ctx context.Context, progressFunc func(int), plexMovies []types.PlexMovie) (searchResults []types.MovieSearchResponse) {
	numberMoviesProcessed.Store(0)
	mapper := iter.Mapper[types.PlexMovie, types.MovieSearchResponse]{
		MaxGoroutines: types.ConcurrencyLimit,
	}
	searchResults = mapper.Map(plexMovies, func(pm *types.PlexMovie) types.MovieSearchResponse {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return types.MovieSearchResponse{}
		default:
		}
		result := searchCinemaParadisoMovieResponse(pm)
		current := int(numberMoviesProcessed.Add(1))
		if progressFunc != nil {
			progressFunc(current)
		}
		return result
	})
	numberMoviesProcessed.Store(0) // job is done
	return searchResults
}

func ScrapeMoviesParallel(ctx context.Context, searchResults []types.MovieSearchResponse) []types.MovieSearchResponse {
	numberMoviesProcessed.Store(0)
	mapper := iter.Mapper[types.MovieSearchResponse, types.MovieSearchResponse]{
		MaxGoroutines: types.ConcurrencyLimit,
	}
	detailedSearchResults := mapper.Map(searchResults, func(result *types.MovieSearchResponse) types.MovieSearchResponse {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return types.MovieSearchResponse{}
		default:
		}
		res := scrapeMovieTitleResponseValue(result)
		numberMoviesProcessed.Add(1)
		return res
	})
	numberMoviesProcessed.Store(0) // job is done
	return detailedSearchResults
}

// nolint: dupl, nolintlint
func TVInParallel(ctx context.Context, progressFunc func(int), plexTVShows []types.PlexTVShow) (searchResults []types.TVSearchResponse) {
	numberTVProcessed.Store(0)
	mapper := iter.Mapper[types.PlexTVShow, types.TVSearchResponse]{
		MaxGoroutines: types.ConcurrencyLimit,
	}
	searchResults = mapper.Map(plexTVShows, func(tv *types.PlexTVShow) types.TVSearchResponse {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return types.TVSearchResponse{}
		default:
		}
		result := searchTVShowResponseValue(tv)
		current := int(numberTVProcessed.Add(1))
		if progressFunc != nil {
			progressFunc(current)
		}
		return result
	})
	numberTVProcessed.Store(0) // job is done
	return searchResults
}

// searchTVShowResponseValue is a value-returning version for use with iter.Map
func searchTVShowResponseValue(plexTVShow *types.PlexTVShow) types.TVSearchResponse {
	result := types.TVSearchResponse{}
	urlEncodedTitle := url.QueryEscape(plexTVShow.Title)
	result.PlexTVShow = *plexTVShow
	result.SearchURL = cinemaparadisoSearchURL + "?form-search-field=" + urlEncodedTitle
	rawData, err := makeRequest(result.SearchURL, http.MethodGet, "")
	if err != nil {
		fmt.Println("searchTVShow: Error making web request:", err)
		return result
	}

	_, tvFound := findTitlesInResponse(rawData, false)
	result.TVSearchResults = tvFound
	result = utils.MarkBestMatchTVResponse(&result)
	for i := range result.TVSearchResults {
		if result.TVSearchResults[i].BestMatch {
			seasonInfo, _ := findTVSeasonInfo(result.TVSearchResults[i].URL)
			if len(seasonInfo) == 0 {
				seasonInfo = append(seasonInfo, types.TVSeasonResult{Number: 1, Format: "DVD", URL: result.TVSearchResults[i].URL})
				result.TVSearchResults[i].BestMatch = false
			}
			result.TVSearchResults[i].Seasons = seasonInfo
		}
	}
	result.MatchesDVD = 0
	result.MatchesBluray = 0
	result.Matches4k = 0
	for i := range result.TVSearchResults {
		for _, season := range result.TVSearchResults[i].Seasons {
			switch strings.ToLower(season.Format) {
			case "dvd":
				result.MatchesDVD++
			case "blu-ray":
				result.MatchesBluray++
			case "4k":
				result.Matches4k++
			}
		}
	}
	return result
}

func GetMovieJobProgress() int {
	return int(numberMoviesProcessed.Load())
}

func GetTVJobProgress() int {
	return int(numberTVProcessed.Load())
}

func searchCinemaParadisoMovieResponse(plexMovie *types.PlexMovie) types.MovieSearchResponse {
	result := types.MovieSearchResponse{}
	result.PlexMovie = *plexMovie
	urlEncodedTitle := url.QueryEscape(plexMovie.Title)
	result.SearchURL = cinemaparadisoSearchURL + "?form-search-field=" + urlEncodedTitle
	rawData, err := makeRequest(result.SearchURL, http.MethodPost, fmt.Sprintf("form-search-field=%s", urlEncodedTitle))
	if err != nil {
		fmt.Println("searchCinemaParadisoMovie:", err)
		return result
	}

	moviesFound, _ := findTitlesInResponse(rawData, true)
	result.MovieSearchResults = moviesFound
	result = utils.MarkBestMatchMovieResponse(&result)
	return result
}

// scrapeMovieTitleResponseValue is a value-returning version for use with iter.Map
func scrapeMovieTitleResponseValue(result *types.MovieSearchResponse) types.MovieSearchResponse {
	// Copy to avoid mutating input
	res := *result
	for i := range res.MovieSearchResults {
		if !res.MovieSearchResults[i].BestMatch {
			continue
		}
		rawData, err := makeRequest(res.MovieSearchResults[i].URL, http.MethodGet, "")
		if err != nil {
			fmt.Println("scrapeMovieTitle:", err)
			return res
		}
		r := regexp.MustCompile(`<section id="format-(.*?)".*?Release Date:<\/dt><dd>(.*?)<\/dd>`)
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
		_, ok := discReleases[res.MovieSearchResults[i].Format]
		if ok {
			res.MovieSearchResults[i].ReleaseDate = discReleases[res.MovieSearchResults[i].Format]
		} else {
			res.MovieSearchResults[i].ReleaseDate = time.Time{}
		}
		if res.MovieSearchResults[i].ReleaseDate.After(res.DateAdded) {
			res.MovieSearchResults[i].NewRelease = true
		}
	}
	return res
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
	// Match list items with data-filmid (case-insensitive attribute name) and capture the id and the inner text
	// Example: <li data-filmid="1832">Series 1<span class="arrow"></span></li>
	r := regexp.MustCompile(`(?i)<li\s+data-filmid="(\d+)">([^<]*)`)
	match := r.FindAllStringSubmatch(response, -1)
	for _, m := range match {
		id := m[1]
		label := strings.TrimSpace(m[2])
		// Only include entries with "Series <number>" (case-insensitive)
		sr := regexp.MustCompile(`(?i)^Series\s+(\d+)$`)
		sMatch := sr.FindStringSubmatch(label)
		if sMatch == nil {
			// skip non-series entries like General info or Specials
			continue
		}
		num, err := strconv.Atoi(sMatch[1])
		if err != nil {
			continue
		}
		tvSeasons = append(tvSeasons, types.TVSeasonResult{Number: num, URL: id})
	}
	scrapedTVSeasonResults := make([]types.TVSeasonResult, 0, len(tvSeasons))
	if len(tvSeasons) > 0 {
		for i := range tvSeasons {
			detailedSeasonResults, err := makeSeasonRequest(&tvSeasons[i])
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

func makeSeasonRequest(tv *types.TVSeasonResult) (result []types.TVSeasonResult, err error) {
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

			// Find the URL of the movie/TV show
			// For TV shows with h4 tags, the URL might be in the h4 link or the image link
			urlStartIndex := strings.Index(movieEntry, "href=\"")
			if urlStartIndex == -1 {
				response = response[endIndex:]
				continue
			}
			urlStartIndex += len("href=\"")
			urlEndIndex := strings.Index(movieEntry[urlStartIndex:], "\"")
			if urlEndIndex == -1 {
				response = response[endIndex:]
				continue
			}
			urlEndIndex += urlStartIndex
			returnURL := movieEntry[urlStartIndex:urlEndIndex]
			// Make sure URL is absolute
			if !strings.HasPrefix(returnURL, "http") {
				returnURL = "https://www.cinemaparadiso.co.uk" + returnURL
			}

			// Find the formats of the movies
			formats := extractDiscFormats(movieEntry)

			// Find the title of the movie/TV show
			// For movies: <a>Title (Year)</a>
			// For TV shows: <a>Title</a> or <a>Title (Year)</a> or <h4><a>Title</a></h4>
			var r *regexp.Regexp
			if movie {
				r = regexp.MustCompile(`<a.*?>(.*?)\s*\((.*?)\)</a>`)
			} else {
				// For TV shows, try both patterns: with year and without year, and handle h4 tags
				// Check for year pattern first (more specific), then h4 with year, then h4 without year
				r = regexp.MustCompile(`<a[^>]*>(.*?)\s*\((.*?)\)</a>|<h4><a[^>]*>(.*?)\s*\((.*?)\)</a></h4>|<h4><a[^>]*>(.*?)</a></h4>`)
			}

			// Find the first match
			match := r.FindStringSubmatch(movieEntry)

			if match != nil {
				// Extract and print title and year
				var foundTitle, year string
				if movie {
					foundTitle = match[1]
					year = match[2]
				} else {
					// For TV shows, check which pattern matched
					if match[1] != "" {
						// Matched <a>Title (Year)</a> pattern
						foundTitle = strings.TrimSpace(match[1])
						year = match[2]
					} else if match[3] != "" {
						// Matched <h4><a>Title (Year)</a></h4> pattern
						foundTitle = strings.TrimSpace(match[3])
						year = match[4]
					} else if match[5] != "" {
						// Matched <h4><a>Title</a></h4> pattern (no year)
						foundTitle = strings.TrimSpace(match[5])
						year = "" // No year available
					} else {
						// No match, skip
						response = response[endIndex:]
						continue
					}
				}
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
		} else {
			// this is to look up individual tv series/seasons
			req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
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
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error response code:", resp.StatusCode, urlEncodedTitle)
	}
	rawData := string(body)

	//nolint
	// bla := strings.Split(urlEncodedTitle, "=")
	// if len(bla) > 1 {
	// 	err = os.WriteFile(fmt.Sprintf("%s.html", bla[1]), body, 0644)
	// 	if err != nil {
	// 		fmt.Println("Error writing file:", bla[1], err)
	// 	}
	// }

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
