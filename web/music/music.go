package music

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tphoney/plex-lookup/musicbrainz"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/spotify"
	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/utils"
)

var (
	//go:embed music.html
	musicPage string

	numberOfArtistsProcessed int  = 0
	artistsJobRunning        bool = false
	totalArtists             int  = 0

	plexMusic              []types.PlexMusicArtist
	artistsSearchResults   []types.SearchResults
	similarArtistsResults  map[string]types.MusicSimilarArtistResult
	lookup                 string
	lookupType             string
	spotifyThreads         int = 2
	albumReleaseYearCutoff int = 5
)

type MusicConfig struct {
	Config *types.Configuration
}

func MusicHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("music").Parse(musicPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render music page", http.StatusInternalServerError)
		return
	}
}

// nolint: lll, nolintlint
func (c MusicConfig) ProcessHTML(w http.ResponseWriter, r *http.Request) {
	lookup = r.FormValue("lookup")
	if lookup == "musicbrainz" {
		if c.Config.MusicBrainzURL == "" {
			fmt.Fprintf(w, `<div class="container"><b>MusicBrainz URL is not set</b>. Please set in <a href="/settings">settings.</a></div>`)
			return
		}
	}
	if lookup == "spotify" {
		if c.Config.SpotifyClientID == "" || c.Config.SpotifyClientSecret == "" {
			fmt.Fprintf(w, `<div class="container"><b>Spotify Client ID or Secret is not set</b>. Please set in <a href="/settings">settings.</a></div>`)
			return
		}
	}
	lookupType = r.FormValue("lookuptype")
	// only get the artists from plex once
	if len(plexMusic) == 0 {
		plexMusic = plex.GetPlexMusicArtists(c.Config.PlexIP, c.Config.PlexMusicLibraryID, c.Config.PlexToken)
	}
	//nolint: gocritic
	// plexMusic = plexMusic[:30]
	//lint: gocritic
	var searchResult types.SearchResults
	artistsJobRunning = true
	numberOfArtistsProcessed = 0
	totalArtists = len(plexMusic) - 1

	fmt.Fprintf(w, `<div hx-get="/musicprogress" hx-trigger="every 100ms" class="container" id="progress">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)

	if lookupType == "missingalbums" {
		switch lookup {
		case "musicbrainz":
			// limit the number of artists to 50 for nonlocal musicbrainz instances
			if strings.Contains(c.Config.MusicBrainzURL, "musicbrainz.org") {
				plexMusic = plexMusic[:50]
				totalArtists = len(plexMusic) - 1
			}
			go func() {
				startTime := time.Now()
				for i := range plexMusic {
					fmt.Print(".")
					searchResult, _ = musicbrainz.SearchMusicBrainzArtist(&plexMusic[i], c.Config.MusicBrainzURL)
					artistsSearchResults = append(artistsSearchResults, searchResult)
					numberOfArtistsProcessed = i
				}
				totalArtists = numberOfArtistsProcessed
				artistsJobRunning = false
				fmt.Printf("Processed %d artists in %v\n", numberOfArtistsProcessed, time.Since(startTime))
			}()
		default:
			// search spotify
			go func() {
				getSpotifyArtistsInParallel(c.Config.SpotifyClientID, c.Config.SpotifyClientSecret)
				numberOfArtistsProcessed = 0
				getSpotifyAlbumsInParallel(c.Config.SpotifyClientID, c.Config.SpotifyClientSecret)
				totalArtists = numberOfArtistsProcessed
				artistsJobRunning = false
			}()
		}
	} else {
		switch lookup {
		case "spotify":
			go func() {
				getSpotifyArtistsInParallel(c.Config.SpotifyClientID, c.Config.SpotifyClientSecret)
				numberOfArtistsProcessed = 0
				getSpotifySimilarArtistsInParallel(c.Config.SpotifyClientID, c.Config.SpotifyClientSecret)
				totalArtists = numberOfArtistsProcessed
				artistsJobRunning = false
			}()
			fmt.Println("Searching Spotify for similar artists")
		default:
			fmt.Fprintf(w, `<div class="alert alert-danger" role="alert">Similar Artist search is not available for this lookup provider</div>`)
		}
	}
}

func ProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if artistsJobRunning {
		fmt.Fprintf(w, `<div hx-get="/musicprogress" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)
	} else if totalArtists == numberOfArtistsProcessed && totalArtists != 0 {
		tableContents := ""
		if lookupType == "missingalbums" {
			tableContents = renderArtistAlbumsTable()
		} else {
			tableContents = renderSimilarArtistsTable()
		}
		fmt.Fprintf(w,
			`<table class="table-sortable">%s</tbody></table>
		</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
			tableContents)
		// reset variables
		numberOfArtistsProcessed = 0
		totalArtists = 0
	}
}

func getSpotifyArtistsInParallel(id, token string) {
	startTime := time.Now()
	ch := make(chan *types.SearchResults, len(plexMusic))
	semaphore := make(chan struct{}, spotifyThreads)
	for i := range len(plexMusic) {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			spotify.SearchSpotifyArtist(&plexMusic[i], id, token, ch)
		}(i)
	}
	// gather results
	artistsSearchResults = make([]types.SearchResults, 0)
	for range len(plexMusic) {
		result := <-ch
		artistsSearchResults = append(artistsSearchResults, *result)
		fmt.Print(".")
		numberOfArtistsProcessed++
	}
	fmt.Printf("Processed %d artists in %v\n", len(plexMusic), time.Since(startTime))
}

func getSpotifyAlbumsInParallel(id, token string) {
	fmt.Println("Searching Spotify for artist albums")
	startTime := time.Now()
	ch := make(chan *types.SearchResults, len(artistsSearchResults))
	semaphore := make(chan struct{}, spotifyThreads)
	for i := range artistsSearchResults {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			spotify.SearchSpotifyAlbum(&artistsSearchResults[i], id, token, ch)
		}(i)
	}
	// gather results
	bla := make([]types.SearchResults, 0)
	for range artistsSearchResults {
		result := <-ch
		bla = append(bla, *result)
		fmt.Print(".")
		numberOfArtistsProcessed++
	}
	artistsSearchResults = bla
	fmt.Printf("Processed %d artists in %v", len(bla), time.Since(startTime))
}

func getSpotifySimilarArtistsInParallel(id, token string) {
	fmt.Println("Searching Spotify for similar artists")
	startTime := time.Now()
	ch := make(chan spotify.SimilarArtistsResponse, len(artistsSearchResults))
	semaphore := make(chan struct{}, spotifyThreads)
	for i := range artistsSearchResults {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			spotify.SearchSpotifySimilarArtist(&artistsSearchResults[i], id, token, ch)
		}(i)
	}
	// gather results
	rawSimilarArtists := make([]spotify.SimilarArtistsResponse, 0)
	for range artistsSearchResults {
		result := <-ch
		rawSimilarArtists = append(rawSimilarArtists, result)
		fmt.Print(".")
		numberOfArtistsProcessed++
	}
	fmt.Printf("Retrieved %d similar artists in %v\n", len(rawSimilarArtists), time.Since(startTime))
	// seed the similar artists map with our owned artists
	similarArtistsResults = make(map[string]types.MusicSimilarArtistResult)
	for i := range artistsSearchResults {
		// skip artists with no search results
		if len(artistsSearchResults[i].MusicSearchResults) == 0 {
			continue
		}
		similarArtistsResults[artistsSearchResults[i].MusicSearchResults[0].ID] = types.MusicSimilarArtistResult{
			Name:            artistsSearchResults[i].MusicSearchResults[0].Name,
			URL:             artistsSearchResults[i].MusicSearchResults[0].URL,
			Owned:           true,
			SimilarityCount: 0,
		}
	}
	// iterate over searches
	for i := range rawSimilarArtists {
		// iterate over artists in each search
		for j := range rawSimilarArtists[i].Artists {
			artist, ok := similarArtistsResults[rawSimilarArtists[i].Artists[j].ID]
			if !ok {
				similarArtistsResults[rawSimilarArtists[i].Artists[j].ID] = types.MusicSimilarArtistResult{
					Name:            rawSimilarArtists[i].Artists[j].Name,
					URL:             fmt.Sprintf("https://open.spotify.com/artist/%s", rawSimilarArtists[i].Artists[j].ID),
					Owned:           false,
					SimilarityCount: 1,
				}
			} else {
				// increment the similarity count
				artist.SimilarityCount++
				similarArtistsResults[rawSimilarArtists[i].Artists[j].ID] = artist
			}
		}
	}
	fmt.Printf("Processed %d similar artists in %v\n", len(rawSimilarArtists), time.Since(startTime))
}

func renderArtistAlbumsTable() (tableRows string) {
	searchResults := filterMusicSearchResults(artistsSearchResults)
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Artist</strong></th><th data-sort="int"><strong>Owned Albums</strong></th><th data-sort="int"><strong>Wanted Albums</strong></th><th><strong>Album</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			tableRows += fmt.Sprintf(`<tr><td><a href=%q target="_blank">%s</a></td><td>%d</td><td>%d</td><td><ul>`,
				searchResults[i].MusicSearchResults[0].URL,
				searchResults[i].PlexMusicArtist.Name,
				searchResults[i].MusicSearchResults[0].OwnedAlbums,
				len(searchResults[i].MusicSearchResults[0].Albums))
			for j := range searchResults[i].MusicSearchResults[0].Albums {
				tableRows += fmt.Sprintf(`<li><a href=%q target="_blank">%s</a> (%s)</li>`,
					searchResults[i].MusicSearchResults[0].Albums[j].URL,
					searchResults[i].MusicSearchResults[0].Albums[j].Title,
					searchResults[i].MusicSearchResults[0].Albums[j].Year)
			}
			tableRows += "</ul></td></tr>"
		}
	}
	return tableRows // Return the generated HTML for table rows
}

func renderSimilarArtistsTable() (tableRows string) {
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Artist</strong></th><th data-sort="string"><strong>Owned</strong></th><th data-sort="int"><strong>Similarity Count</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range similarArtistsResults {
		ownedString := "No"
		if similarArtistsResults[i].Owned {
			ownedString = "Yes"
		}
		tableRows += fmt.Sprintf(`<tr><td><a href=%q target="_blank">%s</a></td><td>%s</td><td>%d</td></tr>`,
			similarArtistsResults[i].URL,
			similarArtistsResults[i].Name,
			ownedString,
			similarArtistsResults[i].SimilarityCount)
	}
	return tableRows // Return the generated HTML for table rows
}

func filterMusicSearchResults(searchResults []types.SearchResults) []types.SearchResults {
	searchResults = removeOwnedAlbums(searchResults)
	searchResults = removeOlderSearchedAlbums(searchResults)
	return searchResults
}

func removeOlderSearchedAlbums(searchResults []types.SearchResults) []types.SearchResults {
	cutoffYear := time.Now().Year() - albumReleaseYearCutoff
	filteredResults := make([]types.SearchResults, 0)
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			filteredAlbums := make([]types.MusicAlbumSearchResult, 0)
			for _, album := range searchResults[i].MusicSearchResults[0].Albums {
				albumYear, _ := strconv.Atoi(album.Year)
				if albumYear >= cutoffYear {
					filteredAlbums = append(filteredAlbums, album)
				}
			}
			searchResults[i].MusicSearchResults[0].Albums = filteredAlbums
			filteredResults = append(filteredResults, searchResults[i])
		}
	}
	return filteredResults
}

func removeOwnedAlbums(searchResults []types.SearchResults) []types.SearchResults {
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			albumsToRemove := make([]types.MusicAlbumSearchResult, 0)
			// set the number of owned albums
			searchResults[i].MusicSearchResults[0].OwnedAlbums = len(searchResults[i].PlexMusicArtist.Albums)
			// iterate over plex albums
			for _, plexAlbum := range searchResults[i].PlexMusicArtist.Albums {
				// iterate over search results
				for _, album := range searchResults[i].MusicSearchResults[0].Albums {
					if utils.CompareTitles(plexAlbum.Title, album.Title) {
						albumsToRemove = append(albumsToRemove, album)
					}
				}
			}
			searchResults[i].MusicSearchResults[0].Albums = cleanAlbums(searchResults[i].MusicSearchResults[0].Albums, albumsToRemove)
		}
	}
	return searchResults
}

func cleanAlbums(original, toRemove []types.MusicAlbumSearchResult) []types.MusicAlbumSearchResult {
	cleaned := make([]types.MusicAlbumSearchResult, 0)
	for _, album := range original {
		found := false
		for _, remove := range toRemove {
			if album.Title == remove.Title && album.Year == remove.Year {
				found = true
				break
			}
		}
		if !found {
			cleaned = append(cleaned, album)
		}
	}
	return cleaned
}
