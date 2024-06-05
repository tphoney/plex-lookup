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
	lookupFilters           types.MovieLookupFilters
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

func (c MoviesConfig) PlaylistHTML(w http.ResponseWriter, _ *http.Request) {
	playlistHTML := `<fieldset id="playlist">
	 <label for="All">
		 <input type="radio" id="playlist" name="playlist" value="all" checked />
		 All: dont use a playlist. (SLOW, only use for small libraries)
	 </label>`
	playlists, _ := plex.GetPlaylists(c.Config.PlexIP, c.Config.PlexToken, c.Config.PlexMovieLibraryID)
	fmt.Println("Playlists:", len(playlists))
	for i := range playlists {
		playlistHTML += fmt.Sprintf(
			`<label for=%q>
			<input type="radio" id="playlist" name="playlist" value=%q/>
			%s</label>`,
			playlists[i].Title, playlists[i].RatingKey, playlists[i].Title)
	}

	playlistHTML += `</fieldset>`
	fmt.Fprint(w, playlistHTML)
}

func (c MoviesConfig) ProcessHTML(w http.ResponseWriter, r *http.Request) {
	playlist := r.FormValue("playlist")
	lookup = r.FormValue("lookup")
	// lookup filters
	lookupFilters.AudioLanguage = r.FormValue("language")
	lookupFilters.NewerVersion = r.FormValue("newerVersion") == types.StringTrue
	// fetch from plex
	if playlist == "all" {
		plexMovies = plex.AllMovies(c.Config.PlexIP, c.Config.PlexMovieLibraryID, c.Config.PlexToken)
	} else {
		plexMovies = plex.GetMoviesFromPlaylist(c.Config.PlexIP, c.Config.PlexToken, playlist)
	}
	//nolint: gocritic
	// plexMovies = plexMovies[:30]
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
			searchResults = cinemaparadiso.MoviesInParallel(plexMovies)
			if lookupFilters.NewerVersion {
				searchResults = cinemaparadiso.ScrapeMoviesParallel(searchResults)
			}
		} else {
			searchResults = amazon.MoviesInParallel(plexMovies, lookupFilters.AudioLanguage, c.Config.AmazonRegion)
			// if we are filtering by newer version, we need to search again
			if lookupFilters.NewerVersion {
				searchResults = amazon.ScrapeTitlesParallel(searchResults, c.Config.AmazonRegion, false)
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
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="string"><strong>Plex Audio</strong></th><th data-sort="string"><strong>Plex Resolution</strong></th><th data-sort="int"><strong>Blu-ray</strong></th><th data-sort="int"><strong>4K-ray</strong></th><th data-sort="string"><strong>New release</strong></th><th><strong>Available Discs</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		newRelease := "no"
		for j := range searchResults[i].MovieSearchResults {
			if searchResults[i].MovieSearchResults[j].NewRelease {
				newRelease = "yes"
				break
			}
		}
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]</a></td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td>`,
			searchResults[i].SearchURL, searchResults[i].PlexMovie.Title, searchResults[i].PlexMovie.Year, searchResults[i].PlexMovie.AudioLanguages,
			searchResults[i].PlexMovie.Resolution, searchResults[i].MatchesBluray, searchResults[i].Matches4k, newRelease)
		if searchResults[i].MatchesBluray+searchResults[i].Matches4k > 0 {
			tableRows += "<td>"
			for _, result := range searchResults[i].MovieSearchResults {
				if result.BestMatch && (result.Format == types.DiskBluray || result.Format == types.Disk4K) {
					tableRows += fmt.Sprintf(`<a href=%q target="_blank">%s - %s</a><br>`, result.URL, result.FoundTitle, result.Format)
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
