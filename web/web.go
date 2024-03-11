//go:generate templ generate
//go:generate tailwindcss -i assets/input.css -o assets/dist.css --minify

package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net"
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
	numberOfMovies := 0
	jobRunning := false
	totalMovies := 0

	// find the local IP address
	ipAddress := GetOutboundIP()
	fmt.Printf("Starting server on http://%s:%s\n", ipAddress.String(), port)
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
		var movieResults []types.MovieSearchResults
		var movieResult types.MovieSearchResults
		jobRunning = true
		totalMovies = len(data)
		for i, movie := range data {
			fmt.Print(".")
			if lookup == "cinemaParadiso" {
				movieResult, _ = cinemaparadiso.SearchCinemaParadiso(movie.Title, movie.Year)
			} else {
				movieResult, _ = amazon.SearchAmazon(movie.Title, movie.Year)
			}
			movieResults = append(movieResults, movieResult)
			numberOfMovies = i
		}
		jobRunning = false
		fmt.Fprintf(w, `<table id="movielist">%s</table>`, renderTable(movieResults))
	})

	http.HandleFunc("/progress", func(w http.ResponseWriter, _ *http.Request) {
		if jobRunning {
			fmt.Fprintf(w, "Processing %d of %d", numberOfMovies, totalMovies)
		}
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil) //nolint: gosec
	if err != nil {
		fmt.Printf("Failed to start server on port %s: %s\n", port, err)
		panic(err)
	}
}

func renderTable(movieCollection []types.MovieSearchResults) (tableRows string) {
	tableRows = `<h2 class="container">Results</h2>`
	tableRows += `<tr><th onclick="sortTable(0,false)"><strong>Plex Title</strong></th><th onclick="sortTable(1,true)"><strong>Blu-ray</strong></th><th onclick="sortTable(2,true)"><strong>4K-ray</strong></th><th><strong>Disc</strong></th></tr>` //nolint: lll
	for _, movie := range movieCollection {
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]</a></td><td>%d</td><td>%d</td>`,
			movie.SearchURL, movie.Title, movie.Year, movie.MatchesBluray, movie.Matches4k)
		if movie.MatchesBluray+movie.Matches4k > 0 {
			tableRows += "<td>"
			for _, result := range movie.SearchResults {
				if result.BestMatch && (result.Format == types.DiskBluray || result.Format == types.Disk4K) {
					tableRows += fmt.Sprintf(
						`<a href=%q target="_blank">%v</a> `,
						result.URL, result.UITitle)
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

func fetchPlexMovies(plexIP, plexLibraryID, plexToken string) (allMovies []types.Movie) {
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "sd", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "480", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "576", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "720", plexToken)...)
	return allMovies
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Failed to get local IP address")
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
