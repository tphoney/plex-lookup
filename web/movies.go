package web

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

const (
	stringTrue = "true"
)

var (
	//go:embed movies.html
	moviesPage string

	numberOfMoviesProcessed int  = 0
	jobRunning              bool = false
	totalMovies             int  = 0
	searchResults           []types.SearchResults
)

func moviesHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("movies").Parse(moviesPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render movies page", http.StatusInternalServerError)
		return
	}
}

func processMoviesHTML(w http.ResponseWriter, r *http.Request) {
	lookup := r.FormValue("lookup")
	// plex resolutions
	sd := r.FormValue("SD")
	r240 := r.FormValue("240p")
	r480 := r.FormValue("480p")
	r576 := r.FormValue("576p")
	r720 := r.FormValue("720p")
	r1080 := r.FormValue("1080p")
	r4k := r.FormValue("4K")
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
	newerVersion := r.FormValue("newerVersion")
	// Prepare table plexMovies
	plexMovies := fetchPlexMovies(config.PlexIP, config.PlexMovieLibraryID, config.PlexToken, filteredResolutions, german)
	var searchResult types.SearchResults
	jobRunning = true
	numberOfMoviesProcessed = 0
	totalMovies = len(plexMovies) - 1

	// write progress bar
	fmt.Fprintf(w, `<div hx-get="/progress" hx-trigger="every 100ms" class="container" id="progress">
	<progress value="%d" max= "%d"/></div>`, numberOfMoviesProcessed, totalMovies)

	go func() {
		startTime := time.Now()
		for i, movie := range plexMovies {
			fmt.Print(".")
			if lookup == "cinemaParadiso" {
				searchResult, _ = cinemaparadiso.SearchCinemaParadisoMovie(movie)
			} else {
				if german == stringTrue {
					searchResult, _ = amazon.SearchAmazonMovie(movie, "&audio=german")
				} else {
					searchResult, _ = amazon.SearchAmazonMovie(movie, "")
				}
				// if we are filtering by newer version, we need to search again
				if newerVersion == stringTrue {
					scrapedResults := amazon.ScrapeMovies(&searchResult)
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

func progressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if jobRunning {
		fmt.Fprintf(w, `<div hx-get="/progress" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
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
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="int"><strong>Blu-ray</strong></th><th data-sort="int"><strong>4K-ray</strong></th><th><strong>Disc</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range movieCollection {
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]</a></td><td>%d</td><td>%d</td>`,
			movieCollection[i].SearchURL, movieCollection[i].PlexMovie.Title, movieCollection[i].PlexMovie.Year,
			movieCollection[i].MatchesBluray, movieCollection[i].Matches4k)
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

func fetchPlexMovies(plexIP, plexMovieLibraryID, plexToken string, plexResolutions []string, german string) (allMovies []types.PlexMovie) {
	filter := []plex.Filter{}
	if german == stringTrue {
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
	if len(plexResolutions) == 0 {
		allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, "", filter)...)
	} else {
		for _, resolution := range plexResolutions {
			allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, resolution, filter)...)
		}
	}
	return allMovies
}
