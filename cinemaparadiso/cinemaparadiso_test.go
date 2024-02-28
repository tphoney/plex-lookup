package cinemaparadiso

import (
	"os"
	"testing"
	"time"
)

func TestExtractMovieFormats(t *testing.T) {
	movieEntry := `<ul class="media-types"><li><span class="cpi-dvd cp-tab" title="DVD" data-json={"action":"media-format","filmId":0,"mediaTypeId":1}></span></li><li><span class="cpi-blu-ray cp-tab" title=" Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":3}></span></li><li><span class="cpi-4-k cp-tab" title=" 4K Blu-ray" data-json={"action":"media-format","filmId":0,"mediaTypeId":14}></span></li></ul>` //nolint: lll

	expectedFormats := []string{"DVD", "Blu-ray", "4K Blu-ray"}
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

	if len(searchResult) != 15 {
		t.Errorf("Expected 15 search result, but got %d", len(searchResult))
	}

	if searchResult[0].title != "Cats" {
		t.Errorf("Expected title Cats, but got %s", searchResult[0].title)
	}
	if searchResult[0].year != "1998" {
		t.Errorf("Expected year 1998, but got %s", searchResult[0].year)
	}
	// check formats
	if searchResult[0].formats[0] != "DVD" {
		t.Errorf("Expected format DVD, but got %s", searchResult[0].formats[0])
	}
	if searchResult[0].formats[1] != "Blu-ray" {
		t.Errorf("Expected format Blu-ray, but got %s", searchResult[0].formats[0])
	}
}

func TestYearToDate(t *testing.T) {
	// Test case 1: Valid year string
	yearString := "2022"
	expectedDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	result := yearToDate(yearString)
	if result != expectedDate {
		t.Errorf("Expected date %v, but got %v", expectedDate, result)
	}

	// Test case 2: Invalid year string
	yearString = "abcd"
	expectedDate = time.Time{}
	result = yearToDate(yearString)
	if result != expectedDate {
		t.Errorf("Expected date %v, but got %v", expectedDate, result)
	}
}
func TestMatchTitle(t *testing.T) { //nolint: gocyclo
	results := []searchResult{
		{title: "Cats", year: "1998", url: "https://example.com/cats", formats: []string{"DVD", "Blu-ray"}},
		{title: "Dogs", year: "2000", url: "https://example.com/dogs", formats: []string{"DVD"}},
		{title: "Birds", year: "2002", url: "https://example.com/birds", formats: []string{"Blu-ray"}},
	}

	// Test case 1: Matching title and year
	hit, returnURL, formats := markBestMatch("Cats", "1998", results)
	if !hit {
		t.Errorf("Expected hit to be true, but got false")
	}
	if returnURL != "https://example.com/cats" {
		t.Errorf("Expected returnURL to be 'https://example.com/cats', but got '%s'", returnURL)
	}
	expectedFormats := []string{"DVD", "Blu-ray"}
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, but got %d", len(expectedFormats), len(formats))
	}
	for i, format := range formats {
		if format != expectedFormats[i] {
			t.Errorf("Expected format %s, but got %s", expectedFormats[i], format)
		}
	}

	// Test case 2: Non-matching title
	hit, returnURL, formats = markBestMatch("Dogs", "1998", results)
	if hit {
		t.Errorf("Expected hit to be false, but got true")
	}
	if returnURL != "" {
		t.Errorf("Expected returnURL to be empty, but got '%s'", returnURL)
	}
	if formats != nil {
		t.Errorf("Expected formats to be nil, but got %v", formats)
	}

	// Test case 3: Non-matching year
	hit, returnURL, formats = markBestMatch("Cats", "2000", results)
	if hit {
		t.Errorf("Expected hit to be false, but got true")
	}
	if returnURL != "" {
		t.Errorf("Expected returnURL to be empty, but got '%s'", returnURL)
	}
	if formats != nil {
		t.Errorf("Expected formats to be nil, but got %v", formats)
	}

	// Test case 4: Matching title and year within 1 year difference
	hit, returnURL, formats = markBestMatch("Cats", "1999", results)
	if !hit {
		t.Errorf("Expected hit to be true, but got false")
	}
	if returnURL != "https://example.com/cats" {
		t.Errorf("Expected returnURL to be 'https://example.com/cats', but got '%s'", returnURL)
	}
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, but got %d", len(expectedFormats), len(formats))
	}
	for i, format := range formats {
		if format != expectedFormats[i] {
			t.Errorf("Expected format %s, but got %s", expectedFormats[i], format)
		}
	}
}
