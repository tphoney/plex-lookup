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

	if len(searchResult) != 2 {
		t.Errorf("Expected 2 search result, but got %d", len(searchResult))
	}

	if searchResult[0].FoundTitle != "Cats" {
		t.Errorf("Expected title Cats, but got %s", searchResult[0].FoundTitle)
	}
	if searchResult[0].Year != "2019" {
		t.Errorf("Expected year 2019, but got %s", searchResult[0].Year)
	}
	// check formats
	if searchResult[0].Format != types.DiskBluray {
		t.Errorf("Expected format Blu-ray, but got %s", searchResult[0].Format)
	}
}

func TestSearchAmazon(t *testing.T) {
	result, err := SearchAmazon("napoleon dynamite", "2004")
	if err != nil {
		t.Errorf("Error searching Amazon: %s", err)
	}
	fmt.Println(result)
}
