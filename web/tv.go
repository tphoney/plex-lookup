package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"slices"
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
	// sd := r.FormValue(types.PlexResolutionSD)
	// r240 := r.FormValue(types.PlexResolution240)
	// r480 := r.FormValue(types.PlexResolution480)
	// r576 := r.FormValue(types.PlexResolution576)
	// r720 := r.FormValue(types.PlexResolution720)
	// r1080 := r.FormValue(types.PlexResolution1080)
	// r4k := r.FormValue(types.PlexResolution4K)
	// plexResolutions := []string{sd, r240, r480, r576, r720, r1080, r4k}
	// // remove empty resolutions
	// var filteredResolutions []string
	// for _, resolution := range plexResolutions {
	// 	if resolution != "" {
	// 		filteredResolutions = append(filteredResolutions, resolution)
	// 	}
	// }
	// lookup filters
	german := r.FormValue("german")
	// newerVersion := r.FormValue("newerVersion")

	if len(plexTV) == 0 {
		plexTV = plex.GetPlexTV(config.PlexIP, config.PlexTVLibraryID, config.PlexToken)
	}

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
	searchResults = filterTVSearchResults(searchResults)
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="int"><strong>Blu-ray Seasons</strong></th><th data-sort="int"><strong>4K-ray Seasons</strong></th><th><strong>Disc</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		// build up plex season / resolution row
		seasony := "Season:"
		for _, season := range searchResults[i].PlexTVShow.Seasons {
			seasony += fmt.Sprintf(" %d@%s,", season.Number, season.LowestResolution)
		}
		seasony = seasony[:len(seasony)-1] // remove trailing comma
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]<br>%s</a></td><td>%d</td><td>%d</td>`,
			searchResults[i].SearchURL, searchResults[i].PlexTVShow.Title, searchResults[i].PlexTVShow.Year, seasony,
			searchResults[i].MatchesBluray, searchResults[i].Matches4k)
		if (searchResults[i].MatchesBluray + searchResults[i].Matches4k) > 0 {
			tableRows += "<td>"
			for j := range searchResults[i].TVSearchResults {
				if searchResults[i].TVSearchResults[j].BestMatch {
					if searchResults[i].TVSearchResults[j].BoxSet {
						tableRows += fmt.Sprintf(`<a href=%q target="_blank">%s Box Set</a></br>`,
							searchResults[i].TVSearchResults[j].URL, searchResults[i].TVSearchResults[j].Format[0])
					} else {
						for _, season := range searchResults[i].TVSearchResults[j].Seasons {
							disks := fmt.Sprintf("%v", season.Format)
							tableRows += fmt.Sprintf(
								`<a href=%q target="_blank">Season %d: %v`,
								searchResults[i].TVSearchResults[j].URL, season.Number, disks)
							tableRows += "</a><br>"
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

func filterTVSearchResults(searchResults []types.SearchResults) []types.SearchResults {
	searchResults = removeOwnedTVSeasons(searchResults)
	return searchResults
}

func removeOwnedTVSeasons(searchResults []types.SearchResults) []types.SearchResults {
	for i := range searchResults {
		if len(searchResults[i].TVSearchResults) > 0 {
			tvSeasonsToRemove := make([]types.TVSeasonResult, 0)
			// iterate over plex tv season
			for _, plexSeasons := range searchResults[i].PlexTVShow.Seasons {
				// iterate over search results
				for _, searchSeasons := range searchResults[i].TVSearchResults[0].Seasons {
					if searchSeasons.Number == plexSeasons.Number && discBeatsPlexResolution(plexSeasons.LowestResolution, searchSeasons.Format) {
						tvSeasonsToRemove = append(tvSeasonsToRemove, searchSeasons)
					}
				}
			}
			searchResults[i].TVSearchResults[0].Seasons = cleanTVSeasons(searchResults[i].TVSearchResults[0].Seasons, tvSeasonsToRemove)
		}
	}
	return searchResults
}

func cleanTVSeasons(original, toRemove []types.TVSeasonResult) []types.TVSeasonResult {
	cleaned := make([]types.TVSeasonResult, 0)
	for _, season := range original {
		found := false
		for _, remove := range toRemove {
			if season.Number == remove.Number {
				found = true
				break
			}
		}
		if !found {
			cleaned = append(cleaned, season)
		}
	}
	return cleaned
}

func discBeatsPlexResolution(lowestPlexResolution string, format []string) bool {
	for i := range format {
		switch format[i] {
		case types.Disk4K:
			return true // 4K beats everything
		case types.DiskBluray:
			if slices.Contains([]string{types.PlexResolution1080, types.PlexResolution720, // HD
				types.PlexResolution576, types.PlexResolution480, types.PlexResolution240, types.PlexResolutionSD}, // SD
				lowestPlexResolution) {
				return true
			}
		} // DVD is not considered
	}
	return false
}
