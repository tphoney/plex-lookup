package amazon

import (
	"fmt"
	"os"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

const (
	amazonRegion = "uk"
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
		t.Fatalf("Expected 2 search result, but got %d", len(searchResult))
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
	result := MoviesInParallel([]types.PlexMovie{{Title: "napoleon dynamite", Year: "2004"}}, "", amazonRegion)
	if len(result) == 0 {
		t.Errorf("Expected search results, but got none")
	}
	if len(result) == 0 {
		t.Errorf("Expected search results, but got none")
	}
}

func TestSearchAmazonTV(t *testing.T) {
	if plexIP == "" || plexToken == "" {
		t.Skip("ACCEPTANCE TEST: PLEX environment variables not set")
	}
	show := types.PlexTVShow{
		// Title: "Friends",
		// Year:  "1994",
		// Title: "Charmed",
		// Year:  "1998",
		// Title: "Adventure Time",
		// Year:  "2010",
		Title: "Star Trek: Enterprise",
		Year:  "2001",
	}
	result := TVInParallel([]types.PlexTVShow{show}, "", amazonRegion)

	if len(result) == 0 {
		t.Errorf("Expected search results, but got none")
	}
}

func TestScrapeTitlesParallel(t *testing.T) {
	result := ScrapeTitlesParallel([]types.TVSearchResponse{
		{
			PlexTVShow: types.PlexTVShow{
				Title: "Some TV Show",
				Year:  "2020",
			},
			TVSearchResults: []types.TVSearchResult{
				{
					FoundTitle: "Some TV Show",
					URL:        "https://www.example.com/tv/some-tv-show",
					BestMatch:  true,
				},
			},
		},
	}, amazonRegion)

	if len(result) == 0 {
		t.Fatalf("Expected search results, but got none")
	}
	if len(result[0].TVSearchResults) == 0 {
		t.Fatalf("Expected TV search results, but got none")
	}
	fmt.Println(result)
}

func Test_decipherTVName(t *testing.T) {
	tests := []struct {
		testName        string
		wantTitle       string
		wantNumber      int
		wantBoxSet      bool
		wantBoxSetTitle string
	}{
		{
			testName:        "The Big Bang Theory: The Complete Series",
			wantTitle:       "The Big Bang Theory",
			wantNumber:      0,
			wantBoxSet:      true,
			wantBoxSetTitle: "The Complete Series",
		},
		{
			testName:        "The Big Bang Theory: The Complete Fourth Season Blu-ray",
			wantTitle:       "The Big Bang Theory",
			wantNumber:      4,
			wantBoxSet:      false,
			wantBoxSetTitle: "",
		},
		{
			testName:        "MACGYVER: THE COMPLETE COLLECTION",
			wantTitle:       "MACGYVER",
			wantNumber:      0,
			wantBoxSet:      true,
			wantBoxSetTitle: "THE COMPLETE COLLECTION",
		},
		{
			testName:        "The Big Bang Theory: Seasons 1-6",
			wantTitle:       "The Big Bang Theory",
			wantNumber:      0,
			wantBoxSet:      true,
			wantBoxSetTitle: "Seasons 1-6",
		},
		{
			testName:        "Star Wars: The Clone Wars: The Complete Season One",
			wantTitle:       "Star Wars: The Clone Wars",
			wantNumber:      1,
			wantBoxSet:      false,
			wantBoxSetTitle: "The Complete Season One",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gotTitle, gotNumber, gotBoxSet, gotBoxSetTitle := decipherTVName(tt.testName)
			if gotTitle != tt.wantTitle {
				t.Errorf("decipherTVName() gotTitle = %v, want %v", gotTitle, tt.wantTitle)
			}
			if gotNumber != tt.wantNumber {
				t.Errorf("decipherTVName() gotNumber = %v, want %v", gotNumber, tt.wantNumber)
			}
			if gotBoxSet != tt.wantBoxSet {
				t.Errorf("decipherTVName() gotBoxSet = %v, want %v", gotBoxSet, tt.wantBoxSet)
			}
			if gotBoxSetTitle != tt.wantBoxSetTitle {
				t.Errorf("decipherTVName() gotBoxSetTitle = %v, want %v", gotBoxSetTitle, tt.wantBoxSetTitle)
			}
		})
	}
}
