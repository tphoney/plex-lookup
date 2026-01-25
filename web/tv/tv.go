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
)

type TVConfig struct {
	Config     *types.Configuration
	JobTracker types.JobTracker
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
	tracker := c.JobTracker
	if tracker == nil {
		http.Error(w, "Job tracker not available", http.StatusInternalServerError)
		return
	}

	playlist := r.FormValue("playlist")
	lookup := r.FormValue("lookup")
	// lookup filters
	filters := types.MovieLookupFilters{
		AudioLanguage: r.FormValue("language"),
		NewerVersion:  r.FormValue("newerVersion") == types.StringTrue,
	}

	// get TV shows from plex
	var plexTV []types.PlexTVShow
	if playlist == "all" {
		plexTV = plex.AllTV(c.Config.PlexIP, c.Config.PlexToken, c.Config.PlexTVLibraryID)
	} else {
		plexTV = plex.GetTVFromPlaylist(c.Config.PlexIP, c.Config.PlexToken, playlist)
	}

	totalTV := len(plexTV)
	jobID, ctx := tracker.CreateJob("tv", totalTV)

	fmt.Fprintf(w, `<div hx-get="/progress/%s" hx-trigger="every 250ms" class="container" id="progress">
		<progress value="0" max="%d"></progress></div>`, jobID, totalTV)

	go func() {
		startTime := time.Now()
		var tvSearchResults []types.TVSearchResponse

		progressFunc := func(current int) {
			tracker.UpdateProgress(jobID, current, "Processing TV shows")
		}

		if lookup == "cinemaParadiso" {
			tvSearchResults = cinemaparadiso.TVInParallel(ctx, progressFunc, plexTV)
		} else {
			tvSearchResults = amazon.TVInParallel(ctx, progressFunc, plexTV, filters.AudioLanguage, c.Config.AmazonRegion)
			tvSearchResults = amazon.ScrapeTitlesParallel(ctx, tvSearchResults, c.Config.AmazonRegion)
		}

		resultsHTML := fmt.Sprintf(`<table class="table-sortable">%s</tbody></table>
		<script>document.querySelector('.table-sortable').tsortable()</script>`,
			renderTVTable(tvSearchResults))

		tracker.MarkComplete(jobID, resultsHTML)
		fmt.Printf("\nProcessed %d TV Shows in %v\n", totalTV, time.Since(startTime))
	}()
}

func renderTVTable(searchResults []types.TVSearchResponse) (tableRows string) {
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
			searchResults[i].SearchURL, searchResults[i].Title, searchResults[i].Year, plexSeasonsString,
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
