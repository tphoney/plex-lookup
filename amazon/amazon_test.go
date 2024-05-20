package amazon

import (
	"fmt"
	"os"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

var (
	plexIP    = os.Getenv("PLEX_IP")
	plexToken = os.Getenv("PLEX_TOKEN")
)

func TestFindMoviesInResponse(t *testing.T) {
	// read response from testdata/cats.html
	rawdata, err := os.ReadFile("testdata/cats_search.html")
	if err != nil {
		t.Errorf("Error reading testdata/cats_search.html: %s", err)
	}

	searchResult, _ := findTitlesInResponse(string(rawdata), true)

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
	result := SearchAmazonMoviesInParallel([]types.PlexMovie{{Title: "napoleon dynamite", Year: "2004"}}, "")
	if len(result) == 0 {
		t.Errorf("Expected search results, but got none")
	}
	fmt.Println(result)
}

func TestSearchAmazonTV(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	show := types.PlexTVShow{
		Title: "Friends",
		Year:  "1994",
		// Title: "Charmed",
		// Year:  "1998",
		// Title: "Adventure Time",
		// Year:  "2010",
	}
	result := SearchAmazonTVInParallel([]types.PlexTVShow{show}, "")

	if len(result) == 0 {
		t.Errorf("Expected search results, but got none")
	}
	fmt.Println(result)
}
