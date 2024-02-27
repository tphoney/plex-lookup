//go:generate templ generate
//go:generate tailwindcss -i assets/input.css -o assets/dist.css --minify

package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/tphoney/plex-lookup/amazon"
	"github.com/tphoney/plex-lookup/cinemaparadiso"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed index.html
	indexHTML string
	port      string = "9090"
)

func StartServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		tmpl := template.Must(template.New("index").Parse(indexHTML))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to render index", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve form fields (replace with proper values)
		plexIP := r.FormValue("plexIP")
		plexLibraryID := r.FormValue("plexLibraryID")
		plexToken := r.FormValue("plexToken")
		lookup := r.FormValue("lookup")

		// Prepare table data
		data := fetchPlexMovies(plexIP, plexLibraryID, plexToken)
		var spax []pooPoo

		if lookup == "cinemaParadiso" {
			spax = cinemaParadisoLookup(data)
		} else {
			spax = amazonLookup(data)
		}

		// Render table with HTMX
		fmt.Fprintf(w, `<table>%s</table>`, renderTable(spax))
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil) //nolint: gosec
	if err != nil {
		fmt.Printf("Failed to start server on port %s: %s\n", port, err)
		panic(err)
	}
}

func renderTable(spax []pooPoo) (tableRows string) {
	tableRows = `<h2 class="container">Results</h2>`
	tableRows += `<tr><th>Title</th><th>Year</th><th>Hit</th><th>Formats</th></tr>`
	for _, movie := range spax {
		tableRows += fmt.Sprintf(`<tr><td>%s</td><td>%v</td><td>%v</td><td><a href=%q target="_blank">%v</a></td></tr>`,
			movie.Title, movie.Year, movie.hit, movie.url, movie.formats)
	}
	return tableRows // Return the generated HTML for table rows
}

func fetchPlexMovies(plexIP, plexLibraryID, plexToken string) (allMovies []types.Movie) {
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "sd", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "480", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "576", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "720", plexToken)...)
	return allMovies
}

type pooPoo struct {
	types.Movie
	hit     bool
	url     string
	formats []string
}

func cinemaParadisoLookup(allMovies []types.Movie) (spax []pooPoo) {
	for _, movie := range allMovies {
		hit, url, formats := cinemaparadiso.SearchCinemaParadiso(movie.Title, movie.Year)
		spax = append(spax, pooPoo{movie, hit, url, formats})
	}
	return spax
}

func amazonLookup(allMovies []types.Movie) (spax []pooPoo) {
	for _, movie := range allMovies {
		hit, url, formats := amazon.SearchAmazon(movie.Title, movie.Year)
		spax = append(spax, pooPoo{movie, hit, url, formats})
	}
	return spax
}
