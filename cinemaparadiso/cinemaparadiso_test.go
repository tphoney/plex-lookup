package cinemaparadiso

import (
	"os"
	"testing"
	"time"

	"github.com/tphoney/plex-lookup/types"
)

var (
	plexIP    = os.Getenv("PLEX_IP")
	plexToken = os.Getenv("PLEX_TOKEN")
)

func TestExtractDiscFormats(t *testing.T) {
	t.Parallel()
	movieEntry := `<ul class="media-types"><li><span class="cpi-dvd cp-tab" title="DVD" data-json={"action":"media-format","filmId":0,"mediaTypeId":1}></span></li><li><span class="cpi-blu-ray cp-tab" title=" Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":3}></span></li><li><span class="cpi-4-k cp-tab" title=" 4K Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":14}></span></li></ul>`

	expectedFormats := []string{types.DiskDVD, types.DiskBluray, types.Disk4K}
	formats := extractDiscFormats(movieEntry)

	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, but got %d", len(expectedFormats), len(formats))
	}

	for i, format := range formats {
		if format != expectedFormats[i] {
			t.Errorf("Expected format %s, but got %s", expectedFormats[i], format)
		}
	}
}

func TestFindTitlesInResponse(t *testing.T) {
	t.Parallel()
	// read response from testdata/cats.html
	rawdata, err := os.ReadFile("testdata/cats.html")
	if err != nil {
		t.Errorf("Error reading testdata/cats.html: %s", err)
	}

	searchResult, _ := findTitlesInResponse(string(rawdata), true)

	if len(searchResult) != 21 {
		t.Fatalf("Expected 21 search result, but got %d", len(searchResult))
	}

	if searchResult[0].FoundTitle != "Cats" {
		t.Errorf("Expected title Cats, but got %s", searchResult[0].FoundTitle)
	}
	if searchResult[0].Year != "1998" {
		t.Errorf("Expected year 1998, but got %s", searchResult[0].Year)
	}
	// check formats
	if searchResult[0].Format != "DVD" {
		t.Errorf("Expected format DVD, but got %s", searchResult[0].Format)
	}
	if searchResult[0].URL != "https://www.cinemaparadiso.co.uk/rentals/cats-3449.html" {
		t.Errorf("Expected url https://www.cinemaparadiso.co.uk/rentals/cats-3449.html, but got %s", searchResult[0].URL)
	}
}

func TestFindTVSeriesInResponse(t *testing.T) {
	t.Parallel()
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	rawdata, err := os.ReadFile("testdata/friends.html")
	if err != nil {
		t.Errorf("Error reading testdata/friends.html: %s", err)
	}

	tvSeries := findTVSeasonsInResponse(string(rawdata))

	if len(tvSeries) != 30 {
		t.Fatalf("Expected 30 tv series, but got %d", len(tvSeries))
	}
	// check the first tv series
	if tvSeries[1].Number != 1 {
		t.Errorf("Expected number 1, but got %d", tvSeries[0].Number)
	}
	expected := time.Date(2004, time.October, 25, 0, 0, 0, 0, time.UTC)
	if tvSeries[0].ReleaseDate.Compare(expected) != 0 {
		t.Errorf("Expected date %s, but got %s", expected, tvSeries[0].ReleaseDate)
	}
	if tvSeries[0].Number != 1 {
		t.Errorf("Expected number 1, but got %d", tvSeries[0].Number)
	}
	if tvSeries[0].Format != types.DiskDVD {
		t.Errorf("Expected dvd, but got %s", tvSeries[0].Format)
	}
	if len(tvSeries) < 2 {
		t.Fatalf("Expected at least 2 tv series, but got %d", len(tvSeries))
	}
	if tvSeries[1].Number != 1 {
		t.Errorf("Expected number 1, but got %d", tvSeries[0].Number)
	}
	expected = time.Date(2012, time.November, 12, 0, 0, 0, 0, time.UTC)
	if tvSeries[1].ReleaseDate.Compare(expected) != 0 {
		t.Errorf("Expected date %s, but got %s", expected, tvSeries[0].ReleaseDate)
	}
	if tvSeries[1].Number != 1 {
		t.Errorf("Expected number 1, but got %d", tvSeries[0].Number)
	}
	if tvSeries[1].Format != types.DiskBluray {
		t.Errorf("Expected Blu-ray, but got %s", tvSeries[0].Format)
	}
}

func TestSearchCinemaParadisoTV(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	t.Parallel()

	tests := []struct {
		name                    string
		show                    types.PlexTVShow
		numberOfSeasonsExpected int
		expectFound             bool
	}{
		{
			name: "Friends",
			show: types.PlexTVShow{
				Title:             "Friends",
				FirstEpisodeAired: time.Date(1994, time.September, 22, 0, 0, 0, 0, time.UTC),
				LastEpisodeAired:  time.Date(2004, time.May, 6, 0, 0, 0, 0, time.UTC),
			},
			numberOfSeasonsExpected: 30,
			expectFound:             true,
		},
		{
			name: "Wellington Paranormal",
			show: types.PlexTVShow{
				Title:             "Wellington Paranormal",
				FirstEpisodeAired: time.Date(2018, time.July, 11, 0, 0, 0, 0, time.UTC),
				LastEpisodeAired:  time.Date(2022, time.June, 28, 0, 0, 0, 0, time.UTC),
			},
			numberOfSeasonsExpected: 8, // 4 series × 2 formats (DVD, Blu-ray) = 8 season results
			expectFound:             true,
		},
		{
			name: "Girls",
			show: types.PlexTVShow{
				Title:             "Girls",
				FirstEpisodeAired: time.Date(2012, time.April, 15, 0, 0, 0, 0, time.UTC),
				LastEpisodeAired:  time.Date(2017, time.April, 16, 0, 0, 0, 0, time.UTC),
			},
			numberOfSeasonsExpected: 12, // 6 series × 2 formats (DVD, Blu-ray) = 12 season results
			expectFound:             true,
		},
		// {
		// 	name: "OnceUponATimeInWonderland",
		// 	show: types.PlexTVShow{
		// 		Title:             "Once Upon a Time in Wonderland",
		// 		FirstEpisodeAired: time.Date(2013, time.October, 10, 0, 0, 0, 0, time.UTC),
		// 		LastEpisodeAired:  time.Date(2014, time.April, 3, 0, 0, 0, 0, time.UTC),
		// 	},
		// 	numberOfSeasonsExpected: 0, // Cinema Paradiso does not have this series
		// },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := searchTVShowResponseValue(&tc.show)

			if got.SearchURL == "" {
				t.Errorf("%s: expected searchurl, but got none", tc.name)
			}

			if tc.expectFound {
				if len(got.TVSearchResults) == 0 {
					t.Errorf("%s: expected search results, but got none", tc.name)
					return
				}
				// Find the best match or use the first result
				var bestMatch *types.TVSearchResult
				for i := range got.TVSearchResults {
					if got.TVSearchResults[i].BestMatch {
						bestMatch = &got.TVSearchResults[i]
						break
					}
				}
				if bestMatch == nil {
					// No best match found, use first result
					bestMatch = &got.TVSearchResults[0]
					t.Errorf("%s: expected best match, but none found. FoundTitle: %s, FirstAiredYear: %s", tc.name, bestMatch.FoundTitle, bestMatch.FirstAiredYear)
					return
				}
				if len(bestMatch.Seasons) != tc.numberOfSeasonsExpected {
					t.Errorf("%s: expected %d seasons, but got %d", tc.name, tc.numberOfSeasonsExpected, len(bestMatch.Seasons))
				}
			} else if len(got.TVSearchResults) != 0 {
				t.Errorf("%s: expected no search results, but got %d", tc.name, len(got.TVSearchResults))
			}
		})
	}
}

func TestSearchCinemaParadisoMovies(t *testing.T) {
	t.Parallel()
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	movie := types.PlexMovie{
		Title: "Cats",
		Year:  "1998",
	}
	result := searchCinemaParadisoMovieResponse(&movie)

	if len(result.MovieSearchResults) == 0 {
		t.Errorf("Expected search results, but got none")
	}

	if result.SearchURL == "" {
		t.Errorf("Expected searchurl, but got none")
	}
}
func TestScrapeMovieTitlesParallel(t *testing.T) {
	t.Parallel()
	searchResults := []types.MovieSearchResponse{
		{
			PlexMovie: types.PlexMovie{
				Title: "Elf",
				Year:  "2021",
			},
			MovieSearchResults: []types.MovieSearchResult{
				{
					URL:       "https://www.cinemaparadiso.co.uk/rentals/elf-10167.html",
					Format:    "Blu-ray",
					Year:      "2003",
					BestMatch: true,
				},
			},
		},
	}

	detailedSearchResults := ScrapeMoviesParallel(t.Context(), searchResults)

	if len(detailedSearchResults) != len(searchResults) {
		t.Errorf("Expected %d detailed search results, but got %d", len(searchResults), len(detailedSearchResults))
	}
	// we should have a release date
	if detailedSearchResults[0].MovieSearchResults[0].ReleaseDate.IsZero() {
		t.Errorf("Expected release date, but got none")
	}
}
