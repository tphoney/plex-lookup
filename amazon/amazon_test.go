package amazon

import (
	"fmt"
	"os"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

func TestFindMoviesInResponse(t *testing.T) {
	// read response from testdata/cats.html
	rawdata, err := os.ReadFile("testdata/cats_search.html")
	if err != nil {
		t.Errorf("Error reading testdata/cats_search.html: %s", err)
	}

	searchResult := findMoviesInResponse(string(rawdata))

	if len(searchResult) != 19 {
		t.Errorf("Expected 2 search result, but got %d", len(searchResult))
	}

	if searchResult[0].FoundTitle != "Cats" {
		t.Errorf("Expected title Cats, but got %s", searchResult[0].FoundTitle)
	}
	if searchResult[0].Year != "1998" {
		t.Errorf("Expected year 1998, but got %s", searchResult[0].Year)
	}
	// check formats
	if searchResult[0].Format != types.DiskBluray {
		t.Errorf("Expected format Blu-ray, but got %s", searchResult[0].Format)
	}
}

func TestSearchAmazon(t *testing.T) {
	result, err := SearchAmazon(types.PlexMovie{Title: "napoleon dynamite", Year: "2004"}, "")
	if err != nil {
		t.Errorf("Error searching Amazon: %s", err)
	}
	fmt.Println(result)
}

func TestFindMovieDetails(t *testing.T) {
	rawdata, err := os.ReadFile("testdata/anchorman.html")
	if err != nil {
		t.Errorf("Error reading testdata/anchorman.html: %s", err)
	}

	processed := findMovieDetails(string(rawdata))

	if processed != "Oct 04, 2010" {
		t.Errorf("Expected Oct 04, 2010, but got %s", processed)
	}
}
