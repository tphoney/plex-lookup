package tv

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/tphoney/plex-lookup/amazon"
	"github.com/tphoney/plex-lookup/cinemaparadiso"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed tv.html
	tvPage string

	numberOfTVProcessed int  = 0
	tvJobRunning        bool = false
	totalTV             int  = 0
	plexTV              []types.PlexTVShow
	tvSearchResults     []types.SearchResult
	lookup              string
	filters             types.MovieLookupFilters
)

type TVConfig struct {
	Config *types.Configuration
}

func TVHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("tv").Parse(tvPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render tv page", http.StatusInternalServerError)
		return
	}
}

func (c TVConfig) PlaylistHTML(w http.ResponseWriter, _ *http.Request) {
	playlistHTML := `<fieldset id="playlist">
	 <label for="All">
		 <input type="radio" id="playlist" name="playlist" value="all" checked />
		 All: dont use a playlist. (SLOW, only use for small libraries)
	 </label>`
	playlists, _ := plex.GetPlaylists(c.Config.PlexIP, c.Config.PlexToken, c.Config.PlexTVLibraryID)
	fmt.Println("Playlists:", len(playlists))
	for i := range playlists {
		playlistHTML += fmt.Sprintf(
			`<label for=%q>
				<input type="radio" id=%q name="playlist" value=%q/>%s
			</label>`,
			playlists[i].RatingKey, playlists[i].RatingKey, playlists[i].RatingKey, playlists[i].Title)
	}
	playlistHTML += `</fieldset>`
	fmt.Fprint(w, playlistHTML)
}

func (c TVConfig) ProcessHTML(w http.ResponseWriter, r *http.Request) {
	playlist := r.FormValue("playlist")
	lookup = r.FormValue("lookup")
	// lookup filters
	filters.AudioLanguage = r.FormValue("language")
	filters.NewerVersion = r.FormValue("newerVersion") == types.StringTrue
	// get TV shows from plex
	if playlist == "all" {
		plexTV = plex.AllTV(c.Config.PlexIP, c.Config.PlexToken, c.Config.PlexTVLibraryID)
	} else {
		plexTV = plex.GetTVFromPlaylist(c.Config.PlexIP, c.Config.PlexToken, playlist)
	}
	//nolint: gocritic
	// plexTV = plexTV[:20]
	//lint: gocritic

	tvJobRunning = true
	numberOfTVProcessed = 0
	totalTV = len(plexTV) - 1

	fmt.Fprintf(w, `<div hx-get="/tvprogress" hx-trigger="every 100ms" class="container" id="progress">
		<progress value="%d" max= "%d"/></div>`, numberOfTVProcessed, totalTV)

	go func() {
		startTime := time.Now()
		if lookup == "cinemaParadiso" {
			tvSearchResults = cinemaparadiso.TVInParallel(plexTV)
		} else {
			tvSearchResults = amazon.TVInParallel(plexTV, filters.AudioLanguage, c.Config.AmazonRegion)
			tvSearchResults = amazon.ScrapeTitlesParallel(tvSearchResults, c.Config.AmazonRegion, true)
		}
		tvJobRunning = false
		fmt.Printf("\nProcessed %d TV Shows in %v\n", totalTV, time.Since(startTime))
	}()
}

func ProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if lookup == "cinemaParadiso" {
		numberOfTVProcessed = cinemaparadiso.GetTVJobProgress()
	} else {
		numberOfTVProcessed = amazon.GetTVJobProgress()
	}
	if tvJobRunning {
		fmt.Fprintf(w, `<div hx-get="/tvprogress" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
		<progress value="%d" max= "%d"/></div>`, numberOfTVProcessed, totalTV)
	} else {
		fmt.Fprintf(w,
			`<table class="table-sortable">%s</tbody></table>
		</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
			renderTVTable(tvSearchResults))
		// reset variables
		numberOfTVProcessed = 0
		totalTV = 0
		tvSearchResults = []types.SearchResult{}
	}
}

func renderTVTable(searchResults []types.SearchResult) (tableRows string) {
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="int"><strong>DVD</strong></th><th data-sort="int"><strong>Blu-ray</strong></th><th data-sort="int"><strong>4K-ray</strong></th><th><strong>Disc</strong></th></tr></thead><tbody>`
	for i := range searchResults {
		// build up plex season / resolution row
		plexSeasonsString := ""
		for j := range searchResults[i].Seasons {
			plexSeasonsString += fmt.Sprintf("Season %d %s<br>",
				searchResults[i].PlexTVShow.Seasons[j].Number, searchResults[i].PlexTVShow.Seasons[j].LowestResolution)
		}
		plexSeasonsString = plexSeasonsString[:len(plexSeasonsString)-1] // remove trailing comma
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]:<br>%s</a></td><td>%d</td><td>%d</td><td>%d</td>`,
			searchResults[i].SearchURL, searchResults[i].PlexTVShow.Title, searchResults[i].PlexTVShow.Year, plexSeasonsString,
			searchResults[i].MatchesDVD, searchResults[i].MatchesBluray, searchResults[i].Matches4k)
		if (searchResults[i].MatchesDVD + searchResults[i].MatchesBluray + searchResults[i].Matches4k) > 0 {
			tableRows += "<td>"
			for j := range searchResults[i].TVSearchResults {
				if searchResults[i].TVSearchResults[j].BestMatch {
					for _, season := range searchResults[i].TVSearchResults[j].Seasons {
						if season.BoxSet {
							tableRows += fmt.Sprintf(
								`<a href=%q target="_blank">%s %s`,
								searchResults[i].TVSearchResults[j].URL, season.BoxSetName, season.Format)
						} else if season.Number == 999 { //nolint:mnd
							tableRows += fmt.Sprintf(
								`<a href=%q target="_blank">Final Season`,
								searchResults[i].TVSearchResults[j].URL)
						} else {
							tableRows += fmt.Sprintf(
								`<a href=%q target="_blank">Season %d %s`,
								searchResults[i].TVSearchResults[j].URL, season.Number, season.Format)
						}
						tableRows += "</a><br>"
					}
				}
			}
			tableRows += "</td>"
		} else {
			tableRows += `<td>No results found</td>`
		}
		tableRows += "</tr>"
	}
	return tableRows // Return the generated HTML for table rows
}
