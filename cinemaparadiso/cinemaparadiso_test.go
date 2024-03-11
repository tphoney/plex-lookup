package cinemaparadiso

import (
	"os"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

func TestExtractMovieFormats(t *testing.T) {
	movieEntry := `<ul class="media-types"><li><span class="cpi-dvd cp-tab" title="DVD" data-json={"action":"media-format","filmId":0,"mediaTypeId":1}></span></li><li><span class="cpi-blu-ray cp-tab" title=" Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":3}></span></li><li><span class="cpi-4-k cp-tab" title=" 4K Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":14}></span></li></ul>` //nolint: lll

	expectedFormats := []string{types.DiskDVD, types.DiskBluray, types.Disk4K}
	formats := extractMovieFormats(movieEntry)

	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, but got %d", len(expectedFormats), len(formats))
	}

	for i, format := range formats {
		if format != expectedFormats[i] {
			t.Errorf("Expected format %s, but got %s", expectedFormats[i], format)
		}
	}
}

func TestFindMoviesInResponse(t *testing.T) {
	// read response from testdata/cats.html
	rawdata, err := os.ReadFile("testdata/cats.html")
	if err != nil {
		t.Errorf("Error reading testdata/cats.html: %s", err)
	}

	searchResult := findMoviesInResponse(string(rawdata))

	if len(searchResult) != 21 {
		t.Errorf("Expected 21 search result, but got %d", len(searchResult))
	}

	if searchResult[0].FormattedTitle != "Cats" {
		t.Errorf("Expected title Cats, but got %s", searchResult[0].FormattedTitle)
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
