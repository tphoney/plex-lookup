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

const (
	stringTrue = "true"
)

var (
	//go:embed movies.html
	moviesPage string

	numberOfMoviesProcessed int  = 0
	jobRunning              bool = false
	totalMovies             int  = 0
)

func moviesHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("movies").Parse(moviesPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render plex page", http.StatusInternalServerError)
		return
	}
}

// nolint: lll, nolintlint
func processMoviesHTML(w http.ResponseWriter, r *http.Request) {
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
	newerVersion := r.FormValue("newerVersion")
	// Prepare table plexMovies
	plexMovies := fetchPlexMovies(PlexInformation.IP, PlexInformation.MovieLibraryID, PlexInformation.Token, filteredResolutions, german)
	var searchResults []types.MovieSearchResults
	var searchResult types.MovieSearchResults
	jobRunning = true
	numberOfMoviesProcessed = 0
	totalMovies = len(plexMovies)
	for i, movie := range plexMovies {
		fmt.Print(".")
		if lookup == "cinemaParadiso" {
			searchResult, _ = cinemaparadiso.SearchCinemaParadiso(movie)
		} else {
			if german == stringTrue {
				searchResult, _ = amazon.SearchAmazon(movie, "&audio=german")
			} else {
				searchResult, _ = amazon.SearchAmazon(movie, "")
			}
			// if we are filtering by newer version, we need to search again
			if newerVersion == stringTrue {
				_, _ = amazon.ScrapeMovies(&searchResult)
			}
		}
		searchResults = append(searchResults, searchResult)
		numberOfMoviesProcessed = i
	}
	jobRunning = false
	fmt.Fprintf(w,
		`<h2 class="container">Results</h2><table class="table-sortable">%s</tbody></table>
<script>function getCellIndex(t){var a=t.parentNode,r=Array.from(a.parentNode.children).indexOf(a);let s=0;for(let e=0;e<a.cells.length;e++){var l=a.cells[e].colSpan;if(s+=l,0===r){if(e===t.cellIndex)return s-1}else if(!isNaN(parseInt(t.dataset.sortCol)))return parseInt(t.dataset.sortCol)}return s-1}let is_sorting_process_on=!1,delay=100;
function tablesort(e){if(is_sorting_process_on)return!1;is_sorting_process_on=!0;var t=e.currentTarget.closest("table"),a=getCellIndex(e.currentTarget),r=e.currentTarget.dataset.sort,s=t.querySelector("th[data-dir]"),s=(s&&s!==e.currentTarget&&delete s.dataset.dir,e.currentTarget.dataset.dir?"asc"===e.currentTarget.dataset.dir?"desc":"asc":e.currentTarget.dataset.sortDefault||"asc"),l=(e.currentTarget.dataset.dir=s,[]),o=t.querySelectorAll("tbody tr");let n,u,c,d,v;for(j=0,jj=o.length;j<jj;j++)for(n=o[j],l.push({tr:n,values:[]}),v=l[j],c=n.querySelectorAll("th, td"),i=0,ii=c.length;i<ii;i++)u=c[i],d=u.dataset.sortValue||u.innerText,"int"===r?d=parseInt(d):"float"===r?d=parseFloat(d):"date"===r&&(d=new Date(d)),v.values.push(d);l.sort("string"===r?"asc"===s?(e,t)=>(""+e.values[a]).localeCompare(t.values[a]):(e,t)=>-(""+e.values[a]).localeCompare(t.values[a]):"asc"===s?(e,t)=>isNaN(e.values[a])||isNaN(t.values[a])?isNaN(e.values[a])?isNaN(t.values[a])?0:-1:1:e.values[a]<t.values[a]?-1:e.values[a]>t.values[a]?1:0:(e,t)=>isNaN(e.values[a])||isNaN(t.values[a])?isNaN(e.values[a])?isNaN(t.values[a])?0:1:-1:e.values[a]<t.values[a]?1:e.values[a]>t.values[a]?-1:0);const N=document.createDocumentFragment();return l.forEach(e=>N.appendChild(e.tr)),t.querySelector("tbody").replaceChildren(N),setTimeout(()=>is_sorting_process_on=!1,delay),!0}Node.prototype.tsortable=function(){this.querySelectorAll("thead th[data-sort], thead td[data-sort]").forEach(e=>e.onclick=tablesort)};
</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
		renderTable(searchResults))
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

func fetchPlexMovies(plexIP, plexLibraryID, plexToken string, plexResolutions []string, german string) (allMovies []types.PlexMovie) {
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
		allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, plexToken, "", filter)...)
	} else {
		for _, resolution := range plexResolutions {
			allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexLibraryID, plexToken, resolution, filter)...)
		}
	}
	return allMovies
}
