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
	movieEntry := `<ul class="media-types"><li><span class="cpi-dvd cp-tab" title="DVD" data-json={"action":"media-format","filmId":0,"mediaTypeId":1}></span></li><li><span class="cpi-blu-ray cp-tab" title=" Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":3}></span></li><li><span class="cpi-4-k cp-tab" title=" 4K Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":14}></span></li></ul>` //nolint: lll

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
	// read response from testdata/cats.html
	rawdata, err := os.ReadFile("testdata/cats.html")
	if err != nil {
		t.Errorf("Error reading testdata/cats.html: %s", err)
	}

	searchResult, _ := findTitlesInResponse(string(rawdata), true)

	if len(searchResult) != 21 {
		t.Errorf("Expected 21 search result, but got %d", len(searchResult))
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
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	rawdata, err := os.ReadFile("testdata/friends.html")
	if err != nil {
		t.Errorf("Error reading testdata/friends.html: %s", err)
	}

	tvSeries := findTVSeriesInResponse(string(rawdata))

	if len(tvSeries) != 10 {
		t.Errorf("Expected 10 tv series, but got %d", len(tvSeries))
	}
	// check the first tv series
	if tvSeries[0].Number != 1 {
		t.Errorf("Expected number 1, but got %d", tvSeries[0].Number)
	}
	expected := time.Date(2012, time.November, 12, 0, 0, 0, 0, time.UTC)
	if tvSeries[0].ReleaseDate.Compare(expected) != 0 {
		t.Errorf("Expected date %s, but got %s", expected, tvSeries[0].ReleaseDate)
	}
	if tvSeries[0].Number != 1 {
		t.Errorf("Expected number 1, but got %d", tvSeries[0].Number)
	}
	if tvSeries[0].Format[0] != types.DiskDVD {
		t.Errorf("Expected format %s, but got %s", types.DiskDVD, tvSeries[0].Format[0])
	}
	if tvSeries[0].Format[1] != types.DiskBluray {
		t.Errorf("Expected format %s, but got %s", types.DiskBluray, tvSeries[0].Format[1])
	}
}

func TestSearchCinemaParadisoTV(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	show := types.PlexTVShow{
		// Title: "Friends",
		// Year:  "1994",
		Title: "Charmed",
		Year:  "1998",
	}
	result, err := SearchCinemaParadisoTV(&show)
	if err != nil {
		t.Errorf("Error searching for TV show: %s", err)
	}
	if result.SearchURL == "" {
		t.Errorf("Expected searchurl, but got none")
	}
}

func TestSearchCinemaParadisoMovies(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	movie := types.PlexMovie{
		Title: "Cats",
		Year:  "1998",
	}
	result, err := SearchCinemaParadisoMovie(movie)
	if err != nil {
		t.Errorf("Error searching for Movie show: %s", err)
	}
	if result.SearchURL == "" {
		t.Errorf("Expected searchurl, but got none")
	}
}
