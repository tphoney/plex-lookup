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
	lookup := r.FormValue("lookup")
	// lookup filters
	german := r.FormValue("german")
	newerVersion := r.FormValue("newerVersion")
	// fetch from plex
	plexMovies = fetchPlexMovies(c.Config.PlexIP, c.Config.PlexMovieLibraryID, c.Config.PlexToken, german)
	//nolint: gocritic
	// plexMovies = plexMovies[:10]
	//nolint: gocritic
	var searchResult types.SearchResults
	jobRunning = true
	numberOfMoviesProcessed = 0
	totalMovies = len(plexMovies) - 1

	// write progress bar
	fmt.Fprintf(w, `<div hx-get="/moviesprogress" hx-trigger="every 100ms" class="container" id="progress">
	<progress value="%d" max= "%d"/></div>`, numberOfMoviesProcessed, totalMovies)

	go func() {
		startTime := time.Now()
		for i, movie := range plexMovies {
			fmt.Print(".")
			if lookup == "cinemaParadiso" {
				searchResult, _ = cinemaparadiso.SearchCinemaParadisoMovie(movie)
			} else {
				if german == types.StringTrue {
					searchResult, _ = amazon.SearchAmazonMovie(movie, "&audio=german")
				} else {
					searchResult, _ = amazon.SearchAmazonMovie(movie, "")
				}
				// if we are filtering by newer version, we need to search again
				if newerVersion == types.StringTrue {
					scrapedResults := amazon.ScrapeTitles(&searchResult)
					searchResult.MovieSearchResults = scrapedResults
				}
			}
			searchResults = append(searchResults, searchResult)
			numberOfMoviesProcessed = i
		}
		jobRunning = false
		fmt.Printf("\nProcessed %d movies in %v\n", totalMovies, time.Since(startTime))
	}()
}

func ProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if jobRunning {
		fmt.Fprintf(w, `<div hx-get="/moviesprogress" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
		<progress value="%d" max= "%d"/></div>`, numberOfMoviesProcessed, totalMovies)
	}
	if totalMovies == numberOfMoviesProcessed && totalMovies != 0 {
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

func renderTable(movieCollection []types.SearchResults) (tableRows string) {
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="string"><strong>Plex Resolution</strong></th><th data-sort="int"><strong>Blu-ray</strong></th><th data-sort="int"><strong>4K-ray</strong></th><th><strong>Disc</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range movieCollection {
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]</a></td><td>%s</td><td>%d</td><td>%d</td>`,
			movieCollection[i].SearchURL, movieCollection[i].PlexMovie.Title, movieCollection[i].PlexMovie.Year,
			movieCollection[i].PlexMovie.Resolution, movieCollection[i].MatchesBluray, movieCollection[i].Matches4k)
		if movieCollection[i].MatchesBluray+movieCollection[i].Matches4k > 0 {
			tableRows += "<td>"
			for _, result := range movieCollection[i].MovieSearchResults {
				if result.BestMatch && (result.Format == types.DiskBluray || result.Format == types.Disk4K) {
					tableRows += fmt.Sprintf(
						`<a href=%q target="_blank">%v`,
						result.URL, result.UITitle)
					if result.NewRelease {
						tableRows += "(new)"
					}
					tableRows += " </a>"
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

func fetchPlexMovies(plexIP, plexMovieLibraryID, plexToken, german string) (allMovies []types.PlexMovie) {
	filter := []plex.Filter{}
	if german == types.StringTrue {
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
