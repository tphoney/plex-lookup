package tv

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
	lookup              string
	filters             types.FilteringOptions
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

func (c TVConfig) ProcessHTML(w http.ResponseWriter, r *http.Request) {
	lookup = r.FormValue("lookup")
	// lookup filters
	newFilters := types.FilteringOptions{}
	newFilters.AudioLanguage = r.FormValue("language")
	newFilters.NewerVersion = r.FormValue("newerVersion") == types.StringTrue
	if len(plexTV) == 0 || filters != newFilters {
		plexTV = plex.GetPlexTV(c.Config.PlexIP, c.Config.PlexTVLibraryID, c.Config.PlexToken)
	}
	filters = newFilters
	//nolint: gocritic
	plexTV = plexTV[:10]
	//lint: gocritic

	var searchResult types.SearchResults
	tvJobRunning = true
	numberOfTVProcessed = 0
	totalTV = len(plexTV) - 1

	fmt.Fprintf(w, `<div hx-get="/tvprogress" hx-trigger="every 100ms" class="container" id="progress">
		<progress value="%d" max= "%d"/></div>`, numberOfTVProcessed, totalTV)

	go func() {
		startTime := time.Now()
		if lookup == "cinemaParadiso" {
			tvSearchResults = cinemaparadiso.GetCinemaParadisoTVInParallel(plexTV)
		} else {
			for i := range plexTV {
				fmt.Print(".")
				if filters.AudioLanguage == "german" {
					searchResult, _ = amazon.SearchAmazonTV(&plexTV[i], fmt.Sprintf("&audio=%s", filters.AudioLanguage))
				} else {
					searchResult, _ = amazon.SearchAmazonTV(&plexTV[i], "")
				}
				tvSearchResults = append(tvSearchResults, searchResult)
				numberOfTVProcessed = i
			}
		}
		tvJobRunning = false
		fmt.Printf("\nProcessed %d TV Shows in %v\n", totalTV, time.Since(startTime))
	}()
}

func ProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if lookup == "cinemaParadiso" {
		numberOfTVProcessed = cinemaparadiso.GetTVJobProgress()
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
			tvSearchResults = []types.SearchResults{}
		}
	} else {
		if tvJobRunning {
			fmt.Fprintf(w, `<div hx-get="/tvprogress" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
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
	if filters.NewerVersion {
		searchResults = removeOldDiscReleases(searchResults)
	}
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
					if searchSeasons.Number == plexSeasons.Number && !discBeatsPlexResolution(plexSeasons.LowestResolution, searchSeasons.Format) {
						tvSeasonsToRemove = append(tvSeasonsToRemove, searchSeasons)
					}
				}
			}
			searchResults[i].TVSearchResults[0].Seasons = cleanTVSeasons(searchResults[i].TVSearchResults[0].Seasons, tvSeasonsToRemove)
		}
	}
	return searchResults
}

func removeOldDiscReleases(searchResults []types.SearchResults) []types.SearchResults {
	for i := range searchResults {
		if len(searchResults[i].TVSearchResults) > 0 {
			tvSeasonsToRemove := make([]types.TVSeasonResult, 0)
			// iterate over plex tv season
			for _, plexSeasons := range searchResults[i].PlexTVShow.Seasons {
				// iterate over search results
				for _, searchSeasons := range searchResults[i].TVSearchResults[0].Seasons {
					if searchSeasons.ReleaseDate.Compare(plexSeasons.LastEpisodeAdded) == 1 {
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
