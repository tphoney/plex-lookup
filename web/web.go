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
	indexHTML               string
	port                    string = "9090"
	numberOfMoviesProcessed int    = 0
	jobRunning              bool   = false
	totalMovies             int    = 0
)

func StartServer() {
	// find the local IP address
	ipAddress := GetOutboundIP()
	fmt.Printf("Starting server on http://%s:%s\n", ipAddress.String(), port)
	http.HandleFunc("/", index)

	http.HandleFunc("/process", processMoviesHTML)

	http.HandleFunc("/progress", progressBarHTML)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil) //nolint: gosec
	if err != nil {
		fmt.Printf("Failed to start server on port %s: %s\n", port, err)
		panic(err)
	}
}

func index(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render index", http.StatusInternalServerError)
		return
	}
}

// nolint: lll, nolintlint
func processMoviesHTML(w http.ResponseWriter, r *http.Request) {
	// Retrieve form fields (replace with proper values)
	plexIP := r.FormValue("plexIP")
	plexLibraryID := r.FormValue("plexLibraryID")
	plexToken := r.FormValue("plexToken")
	lookup := r.FormValue("lookup")
	german := r.FormValue("german")

	// Prepare table data
	data := fetchPlexMovies(plexIP, plexLibraryID, plexToken, german)
	var movieResults []types.MovieSearchResults
	var movieResult types.MovieSearchResults
	jobRunning = true
	numberOfMoviesProcessed = 0
	totalMovies = len(data)
	for i, movie := range data {
		fmt.Print(".")
		if lookup == "cinemaParadiso" {
			movieResult, _ = cinemaparadiso.SearchCinemaParadiso(movie.Title, movie.Year)
		} else {
			if german == "true" {
				movieResult, _ = amazon.SearchAmazon(movie.Title, movie.Year, "&audio=german")
			} else {
				movieResult, _ = amazon.SearchAmazon(movie.Title, movie.Year, "")
			}
		}
		movieResults = append(movieResults, movieResult)
		numberOfMoviesProcessed = i
	}
	jobRunning = false
	fmt.Fprintf(w,
		`<h2 class="container">Results</h2><table class="table-sortable">%s</tbody></table>
<script>function getCellIndex(t){var a=t.parentNode,r=Array.from(a.parentNode.children).indexOf(a);let s=0;for(let e=0;e<a.cells.length;e++){var l=a.cells[e].colSpan;if(s+=l,0===r){if(e===t.cellIndex)return s-1}else if(!isNaN(parseInt(t.dataset.sortCol)))return parseInt(t.dataset.sortCol)}return s-1}let is_sorting_process_on=!1,delay=100;
function tablesort(e){if(is_sorting_process_on)return!1;is_sorting_process_on=!0;var t=e.currentTarget.closest("table"),a=getCellIndex(e.currentTarget),r=e.currentTarget.dataset.sort,s=t.querySelector("th[data-dir]"),s=(s&&s!==e.currentTarget&&delete s.dataset.dir,e.currentTarget.dataset.dir?"asc"===e.currentTarget.dataset.dir?"desc":"asc":e.currentTarget.dataset.sortDefault||"asc"),l=(e.currentTarget.dataset.dir=s,[]),o=t.querySelectorAll("tbody tr");let n,u,c,d,v;for(j=0,jj=o.length;j<jj;j++)for(n=o[j],l.push({tr:n,values:[]}),v=l[j],c=n.querySelectorAll("th, td"),i=0,ii=c.length;i<ii;i++)u=c[i],d=u.dataset.sortValue||u.innerText,"int"===r?d=parseInt(d):"float"===r?d=parseFloat(d):"date"===r&&(d=new Date(d)),v.values.push(d);l.sort("string"===r?"asc"===s?(e,t)=>(""+e.values[a]).localeCompare(t.values[a]):(e,t)=>-(""+e.values[a]).localeCompare(t.values[a]):"asc"===s?(e,t)=>isNaN(e.values[a])||isNaN(t.values[a])?isNaN(e.values[a])?isNaN(t.values[a])?0:-1:1:e.values[a]<t.values[a]?-1:e.values[a]>t.values[a]?1:0:(e,t)=>isNaN(e.values[a])||isNaN(t.values[a])?isNaN(e.values[a])?isNaN(t.values[a])?0:1:-1:e.values[a]<t.values[a]?1:e.values[a]>t.values[a]?-1:0);const N=document.createDocumentFragment();return l.forEach(e=>N.appendChild(e.tr)),t.querySelector("tbody").replaceChildren(N),setTimeout(()=>is_sorting_process_on=!1,delay),!0}Node.prototype.tsortable=function(){this.querySelectorAll("thead th[data-sort], thead td[data-sort]").forEach(e=>e.onclick=tablesort)};
</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
		renderTable(movieResults))
}

func progressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if jobRunning {
		fmt.Fprintf(w, `<progress value="%d" max= "%d"/>`, numberOfMoviesProcessed, totalMovies)
	}
}

func renderTable(movieCollection []types.MovieSearchResults) (tableRows string) {
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="int"><strong>Blu-ray</strong></th><th data-sort="int"><strong>4K-ray</strong></th><th><strong>Disc</strong></th></tr></thead><tbody>` //nolint: lll
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

func fetchPlexMovies(plexIP, plexLibraryID, plexToken, german string) (allMovies []types.Movie) {
	if german == "true" {
		filter := []plex.Filter{
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
		allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "", plexToken, filter)...)
	} else {
		allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "sd", plexToken, nil)...)
		allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "480", plexToken, nil)...)
		allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "576", plexToken, nil)...)
		allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, "720", plexToken, nil)...)
	}
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
