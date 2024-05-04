package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/tphoney/plex-lookup/musicbrainz"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/spotify"
	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed music.html
	musicPage string

	numberOfArtistsProcessed int  = 0
	artistsJobRunning        bool = false
	totalArtists             int  = 0
	plexMusic                []types.PlexMusicArtist
	artistsSearchResults     []types.SearchResults
)

func musicHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("music").Parse(musicPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render music page", http.StatusInternalServerError)
		return
	}
}

// nolint: lll, nolintlint
func processArtistHTML(w http.ResponseWriter, r *http.Request) {
	lookup := r.FormValue("lookup")
	lookupType := r.FormValue("lookuptype")
	// only get the artists from plex once
	if len(plexMusic) == 0 {
		plexMusic = plex.GetPlexMusicArtists(config.PlexIP, config.PlexMusicLibraryID, config.PlexToken)
	}
	var searchResult types.SearchResults
	artistsJobRunning = true
	numberOfArtistsProcessed = 0
	totalArtists = 49 // len(plexMusic) - 1

	fmt.Fprintf(w, `<div hx-get="/progressartists" hx-trigger="every 100ms" class="container" id="progress">
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
				for i := 0; i < 50; i++ {
					fmt.Print(".")
					searchResult, _ = spotify.SearchSpotifyArtist(&plexMusic[i], config.SpotifyClientID, config.SpotifyClientSecret)
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

func artistProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if artistsJobRunning {
		fmt.Fprintf(w, `<div hx-get="/progressartists" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
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
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Artist</strong></th><th data-sort="int"><strong>Albums</strong></th><th><strong>Album</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			tableRows += fmt.Sprintf("<tr><td><a href='%s'>%s</a></td><td>%d</td><td><ul>",
				searchResults[i].MusicSearchResults[0].URL,
				searchResults[i].PlexMusicArtist.Name,
				len(searchResults[i].MusicSearchResults[0].Albums))
			for j := range searchResults[i].MusicSearchResults[0].Albums {
				tableRows += fmt.Sprintf("<li><a href='%s'>%s</a> (%s)</li>",
					searchResults[i].MusicSearchResults[0].Albums[j].URL,
					searchResults[i].MusicSearchResults[0].Albums[j].Title,
					searchResults[i].MusicSearchResults[0].Albums[j].Year)
			}
			tableRows += "</ul></td></tr>"
		}
	}
	return tableRows // Return the generated HTML for table rows
}
