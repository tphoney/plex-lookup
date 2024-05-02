package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed music.html
	musicPage string

	numberOfArtistsProcessed int  = 0
	artistsJobRunning        bool = false
	totalArtists             int  = 0
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
	lookupType := r.FormValue("lookuptype")
	// plex resolutions

	plexMusic := plex.GetPlexMusicArtists(PlexInformation.IP, PlexInformation.MusicLibraryID, PlexInformation.Token)
	var searchResult types.SearchResults
	artistsJobRunning = true
	numberOfArtistsProcessed = 0
	totalTV = len(plexMusic) - 1

	fmt.Fprintf(w, `<div hx-get="/progressartists" hx-trigger="every 100ms" class="container" id="progress">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)

	if lookupType == "missingalbums" {
		fmt.Println("Processing missing albums")
		fmt.Printf("running %v\n", artistsJobRunning)
	} else {
		fmt.Println("Processing new artists")
		fmt.Printf("running %v\n", artistsJobRunning)
	}
	time.Sleep(1 * time.Second)
	artistsJobRunning = false
	fmt.Println(searchResult)
}

func artistProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if tvJobRunning {
		fmt.Fprintf(w, `<div hx-get="/progressartists" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)
	}
	if totalTV == numberOfArtistsProcessed && totalArtists != 0 {
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
	fmt.Printf("Rendering %d artists\n", len(searchResults))
	return tableRows // Return the generated HTML for table rows
}
