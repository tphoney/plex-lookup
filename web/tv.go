package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strings"
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
	tvSearchResults     []types.SearchResults
)

func tvHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("tv").Parse(tvPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render tv page", http.StatusInternalServerError)
		return
	}
}

// nolint: lll, nolintlint
func processTVHTML(w http.ResponseWriter, r *http.Request) {
	lookup := r.FormValue("lookup")
	// plex resolutions
	sd := r.FormValue("sd")
	r240 := r.FormValue("240p")
	r480 := r.FormValue("480p")
	r576 := r.FormValue("576p")
	r720 := r.FormValue("720p")
	r1080 := r.FormValue("1080p")
	r4k := r.FormValue("4k")
	plexResolutions := []string{sd, r240, r480, r576, r720, r1080, r4k}
	// remove empty resolutions
	var filteredResolutions []string
	for _, resolution := range plexResolutions {
		if resolution != "" {
			filteredResolutions = append(filteredResolutions, resolution)
		}
	}
	// lookup filters
	german := r.FormValue("german")
	// newerVersion := r.FormValue("newerVersion")
	// Prepare table plexTV
	plexTV := plex.GetPlexTV(PlexInformation.IP, PlexInformation.TVLibraryID, PlexInformation.Token, filteredResolutions)
	var searchResult types.SearchResults
	tvJobRunning = true
	numberOfTVProcessed = 0
	totalTV = len(plexTV) - 1

	fmt.Fprintf(w, `<div hx-get="/progresstv" hx-trigger="every 100ms" class="container" id="progress">
		<progress value="%d" max= "%d"/></div>`, numberOfTVProcessed, totalTV)

	go func() {
		startTime := time.Now()
		for i := range plexTV {
			fmt.Print(".")
			if lookup == "cinemaParadiso" {
				searchResult, _ = cinemaparadiso.SearchCinemaParadisoTV(&plexTV[i])
			} else {
				if german == stringTrue {
					searchResult, _ = amazon.SearchAmazonTV(&plexTV[i], "&audio=german")
				} else {
					searchResult, _ = amazon.SearchAmazonTV(&plexTV[i], "")
				}
				// if we are filtering by newer version, we need to search again
				// if newerVersion == stringTrue {
				// 	scrapedResults := amazon.ScrapeMovies(&searchResult)
				// 	searchResult.MovieSearchResults = scrapedResults
				// }
			}
			tvSearchResults = append(tvSearchResults, searchResult)
			numberOfTVProcessed = i
		}
		tvJobRunning = false
		fmt.Printf("\nProcessed %d TV Shows in %v\n", totalTV, time.Since(startTime))
	}()
}

func tvProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if tvJobRunning {
		fmt.Fprintf(w, `<div hx-get="/progresstv" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
		<progress value="%d" max= "%d"/></div>`, numberOfTVProcessed, totalTV)
	}
	if totalTV == numberOfTVProcessed && totalTV != 0 {
		fmt.Fprintf(w,
			`<table class="table-sortable">%s</tbody></table>
		</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
			renderTVTable(tvSearchResults))
		// reset variables
		numberOfTVProcessed = 0
		totalTV = 0
		tvSearchResults = []types.SearchResults{}
	}
}

func renderTVTable(searchResults []types.SearchResults) (tableRows string) {
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="int"><strong>Blu-ray Seasons</strong></th><th data-sort="int"><strong>4K-ray Seasons</strong></th><th><strong>Disc</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]</a></td><td>%d</td><td>%d</td>`,
			searchResults[i].SearchURL, searchResults[i].PlexTVShow.Title, searchResults[i].PlexTVShow.Year,
			searchResults[i].MatchesBluray, searchResults[i].Matches4k)
		if (searchResults[i].MatchesBluray + searchResults[i].Matches4k) > 0 {
			tableRows += "<td>"
			for j := range searchResults[i].TVSearchResults {
				if searchResults[i].TVSearchResults[j].BestMatch {
					if searchResults[i].TVSearchResults[j].BoxSet {
						tableRows += fmt.Sprintf(`<a href=%q target="_blank">%s Box Set</a></br>`,
							searchResults[i].TVSearchResults[j].URL, searchResults[i].TVSearchResults[j].Format[0])
					} else {
						for _, series := range searchResults[i].TVSearchResults[j].Series {
							if slices.Contains(series.Format, types.DiskBluray) || slices.Contains(series.Format, types.Disk4K) {
								// remove the dvd format
								disks := fmt.Sprintf("%v", series.Format)
								disks = strings.ReplaceAll(disks, "DVD ", "")
								tableRows += fmt.Sprintf(
									`<a href=%q target="_blank">Season %d: %v`,
									searchResults[i].TVSearchResults[j].URL, series.Number, disks)
								tableRows += "</a><br>"
							}
						}
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
