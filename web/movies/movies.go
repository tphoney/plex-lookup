package movies

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
	//go:embed movies.html
	moviesPage string

	numberOfMoviesProcessed int  = 0
	jobRunning              bool = false
	totalMovies             int  = 0
	searchResults           []types.SearchResults
	plexMovies              []types.PlexMovie
	lookup                  string
	filters                 types.FilteringOptions
)

type MoviesConfig struct {
	Config *types.Configuration
}

func MoviesHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("movies").Parse(moviesPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render movies page", http.StatusInternalServerError)
		return
	}
}

func (c MoviesConfig) ProcessHTML(w http.ResponseWriter, r *http.Request) {
	lookup = r.FormValue("lookup")
	// lookup filters
	newfilters := types.FilteringOptions{}
	newfilters.AudioLanguage = r.FormValue("language")
	newfilters.NewerVersion = r.FormValue("newerVersion") == types.StringTrue
	// fetch from plex
	if len(plexMovies) == 0 || filters != newfilters {
		plexMovies = fetchPlexMovies(c.Config.PlexIP, c.Config.PlexMovieLibraryID, c.Config.PlexToken, filters.AudioLanguage)
	}
	filters = newfilters
	//nolint: gocritic
	// plexMovies = plexMovies[:100]
	//lint: gocritic
	jobRunning = true
	numberOfMoviesProcessed = 0
	totalMovies = len(plexMovies) - 1

	// write progress bar
	fmt.Fprintf(w, `<div hx-get="/moviesprogress" hx-trigger="every 100ms" class="container" id="progress">
	<progress value="%d" max= "%d"/></div>`, numberOfMoviesProcessed, totalMovies)

	go func() {
		startTime := time.Now()
		if lookup == "cinemaParadiso" {
			searchResults = cinemaparadiso.GetCinemaParadisoMoviesInParallel(plexMovies)
		} else {
			searchResults = amazon.SearchAmazonMoviesInParallel(plexMovies, filters.AudioLanguage)
			// if we are filtering by newer version, we need to search again
			if filters.NewerVersion {
				searchResults = amazon.ScrapeTitlesParallel(searchResults)
			}
		}

		jobRunning = false
		fmt.Printf("\nProcessed %d movies in %v\n", totalMovies, time.Since(startTime))
	}()
}

func ProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if lookup == "cinemaParadiso" {
		// check job status
		numberOfMoviesProcessed = cinemaparadiso.GetMovieJobProgress()
	} else {
		// check job status
		numberOfMoviesProcessed = amazon.GetMovieJobProgress()
	}
	if jobRunning {
		fmt.Fprintf(w, `<div hx-get="/moviesprogress" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
			<progress value="%d" max= "%d"/></div>`, numberOfMoviesProcessed, totalMovies)
	} else {
		// display a table
		fmt.Fprintf(w,
			`<table class="table-sortable">%s</tbody></table>
				 <script>document.querySelector('.table-sortable').tsortable()</script>`,
			renderTable(searchResults))
		// reset variables
		numberOfMoviesProcessed = 0
		totalMovies = 0
		searchResults = []types.SearchResults{}
	}
}

func renderTable(searchResults []types.SearchResults) (tableRows string) {
	searchResults = filterMovieSearchResults(searchResults)
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="string"><strong>Plex Audio</strong></th><th data-sort="string"><strong>Plex Resolution</strong></th><th data-sort="int"><strong>Blu-ray</strong></th><th data-sort="int"><strong>4K-ray</strong></th><th data-sort="string"><strong>New release</strong></th><th><strong>Available Discs</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		newRelease := "no"
		if len(searchResults[i].MovieSearchResults) > 0 && searchResults[i].MovieSearchResults[0].NewRelease {
			newRelease = "yes"
		}
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]</a></td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td>`,
			searchResults[i].SearchURL, searchResults[i].PlexMovie.Title, searchResults[i].PlexMovie.Year, searchResults[i].PlexMovie.AudioLanguages,
			searchResults[i].PlexMovie.Resolution, searchResults[i].MatchesBluray, searchResults[i].Matches4k, newRelease)
		if searchResults[i].MatchesBluray+searchResults[i].Matches4k > 0 {
			tableRows += "<td>"
			for _, result := range searchResults[i].MovieSearchResults {
				if result.BestMatch && (result.Format == types.DiskBluray || result.Format == types.Disk4K) {
					tableRows += fmt.Sprintf(`<a href=%q target="_blank">%v </a>`, result.URL, result.UITitle)
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

func fetchPlexMovies(plexIP, plexMovieLibraryID, plexToken, language string) (allMovies []types.PlexMovie) {
	filter := []plex.Filter{}
	if language == "german" {
		filter = []plex.Filter{
			{
				Name:     "audioLanguage",
				Value:    "de",
				Modifier: "\u0021=",
			},
			// {
			// 	Name:     "audioLanguage",
			// 	Value:    "de",
			// 	Modifier: "=",
			// },
		}
	}
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, filter)...)
	return allMovies
}

func filterMovieSearchResults(searchResults []types.SearchResults) []types.SearchResults {
	return searchResults
}
