package amazon

import (
	"fmt"
	"os"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

func TestFindMoviesInResponse(t *testing.T) {
	// read response from testdata/cats.html
	rawdata, err := os.ReadFile("testdata/cats.html")
	if err != nil {
		t.Errorf("Error reading testdata/cats.html: %s", err)
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

// func TestFindSingleMovieInResponse(t *testing.T) {
// 	rawdata, err := os.ReadFile("testdata/erdman.html")
// 	if err != nil {
// 		t.Errorf("Error reading testdata/erdman.html: %s", err)
// 	}

// 	searchResult := findMoviesInResponse(string(rawdata))
// 	if len(searchResult) != 1 {
// 		t.Errorf("Expected 1 search result, but got %d", len(searchResult))
// 	}

// 	if searchResult[0].FoundTitle != "Toni Erdmann" {
// 		t.Errorf("Expected title Erdmen, but got %s", searchResult[0].FoundTitle)
// 	}

// 	if searchResult[0].Year != "2016" {
// 		t.Errorf("Expected year 2016, but got %s", searchResult[0].Year)
// 	}

// }

func TestSearchAmazon(t *testing.T) {
	result, err := SearchAmazon("napoleon dynamite", "2004", "")
	if err != nil {
		t.Errorf("Error searching Amazon: %s", err)
	}
	fmt.Println(result)
}
