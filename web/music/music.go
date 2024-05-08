package music

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
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
	plexMusic                []types.PlexMusicArtist
	artistsSearchResults     []types.SearchResults
	albumReleaseYearCutoff   int = 5
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
	lookup := r.FormValue("lookup")
	lookupType := r.FormValue("lookuptype")
	// only get the artists from plex once
	if len(plexMusic) == 0 {
		plexMusic = plex.GetPlexMusicArtists(c.Config.PlexIP, c.Config.PlexMusicLibraryID, c.Config.PlexToken)
	}
	var searchResult types.SearchResults
	artistsJobRunning = true
	numberOfArtistsProcessed = 0
	totalArtists = len(plexMusic) - 1

	fmt.Fprintf(w, `<div hx-get="/musicprogress" hx-trigger="every 100ms" class="container" id="progress">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)

	if lookupType == "missingalbums" {
		switch lookup {
		case "musicbrainz":
			go func() {
				startTime := time.Now()
				for i := 0; i < 50; i++ {
					fmt.Print(".")
					searchResult, _ = musicbrainz.SearchMusicBrainzArtist(&plexMusic[i])
					artistsSearchResults = append(artistsSearchResults, searchResult)
					numberOfArtistsProcessed = i
				}
				artistsJobRunning = false
				fmt.Printf("Processed %d artists in %v\n", numberOfArtistsProcessed, time.Since(startTime))
			}()
		default:
			// search spotify
			go func() {
				startTime := time.Now()
				for i := range plexMusic {
					fmt.Print(".")
					searchResult, _ = spotify.SearchSpotifyArtist(&plexMusic[i], c.Config.SpotifyClientID, c.Config.SpotifyClientSecret)
					artistsSearchResults = append(artistsSearchResults, searchResult)
					numberOfArtistsProcessed = i
				}
				artistsJobRunning = false
				fmt.Printf("Processed %d artists in %v\n", numberOfArtistsProcessed, time.Since(startTime))
			}()
		}
	} else {
		fmt.Println("Processing new artists")
		fmt.Printf("running %v\n", artistsJobRunning)
		time.Sleep(1 * time.Second)
		artistsJobRunning = false
		fmt.Println(searchResult)
	}
}

func ProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if artistsJobRunning {
		fmt.Fprintf(w, `<div hx-get="/musicprogress" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)
	}
	if totalArtists == numberOfArtistsProcessed && totalArtists != 0 {
		fmt.Fprintf(w,
			`<table class="table-sortable">%s</tbody></table>
		</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
			renderArtistsTable(artistsSearchResults))
		// reset variables
		numberOfArtistsProcessed = 0
		totalArtists = 0
		artistsSearchResults = []types.SearchResults{}
	}
}

func renderArtistsTable(searchResults []types.SearchResults) (tableRows string) {
	searchResults = filterMusicSearchResults(searchResults)
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Artist</strong></th><th data-sort="int"><strong>Albums</strong></th><th><strong>Album</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			tableRows += fmt.Sprintf(`<tr><td><a href=%q target="_blank">%s</a></td><td>%d</td><td><ul>`,
				searchResults[i].MusicSearchResults[0].URL,
				searchResults[i].PlexMusicArtist.Name,
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
			filteredAlbums := make([]types.MusicSearchAlbumResult, 0)
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
			albumsToRemove := make([]types.MusicSearchAlbumResult, 0)
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

func cleanAlbums(original, toRemove []types.MusicSearchAlbumResult) []types.MusicSearchAlbumResult {
	cleaned := make([]types.MusicSearchAlbumResult, 0)
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
