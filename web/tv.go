package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"time"

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
	// german := r.FormValue("german")
	// newerVersion := r.FormValue("newerVersion")
	// Prepare table plexTV
	plexTV := plex.GetPlexTV(PlexInformation.IP, PlexInformation.TVLibraryID, PlexInformation.Token, filteredResolutions)
	var searchResults []types.SearchResults
	var searchResult types.SearchResults
	tvJobRunning = true

	numberOfTVProcessed = 0
	totalTV = len(plexTV)
	startTime := time.Now()
	for i := range plexTV {
		fmt.Print(".")
		if lookup == "cinemaParadiso" {
			searchResult, _ = cinemaparadiso.SearchCinemaParadisoTV(&plexTV[i])
			// } else {
			// 	if german == stringTrue {
			// 		searchResult, _ = amazon.SearchAmazon(plexTV[i], "&audio=german")
			// 	} else {
			// 		searchResult, _ = amazon.SearchAmazon(plexTV[i], "")
			// 	}
			// 	// if we are filtering by newer version, we need to search again
			// 	if newerVersion == stringTrue {
			// 		scrapedResults := amazon.ScrapeMovies(&searchResult)
			// 		searchResult.MovieSearchResults = scrapedResults
			// 	}
		}
		searchResults = append(searchResults, searchResult)
		numberOfTVProcessed = i
	}
	tvJobRunning = false
	fmt.Printf("\nProcessed %d TV Shows in %v\n", totalTV, time.Since(startTime))
	fmt.Fprintf(w,
		`<table class="table-sortable">%s</tbody></table>
<script>function getCellIndex(t){var a=t.parentNode,r=Array.from(a.parentNode.children).indexOf(a);let s=0;for(let e=0;e<a.cells.length;e++){var l=a.cells[e].colSpan;if(s+=l,0===r){if(e===t.cellIndex)return s-1}else if(!isNaN(parseInt(t.dataset.sortCol)))return parseInt(t.dataset.sortCol)}return s-1}let is_sorting_process_on=!1,delay=100;
function tablesort(e){if(is_sorting_process_on)return!1;is_sorting_process_on=!0;var t=e.currentTarget.closest("table"),a=getCellIndex(e.currentTarget),r=e.currentTarget.dataset.sort,s=t.querySelector("th[data-dir]"),s=(s&&s!==e.currentTarget&&delete s.dataset.dir,e.currentTarget.dataset.dir?"asc"===e.currentTarget.dataset.dir?"desc":"asc":e.currentTarget.dataset.sortDefault||"asc"),l=(e.currentTarget.dataset.dir=s,[]),o=t.querySelectorAll("tbody tr");let n,u,c,d,v;for(j=0,jj=o.length;j<jj;j++)for(n=o[j],l.push({tr:n,values:[]}),v=l[j],c=n.querySelectorAll("th, td"),i=0,ii=c.length;i<ii;i++)u=c[i],d=u.dataset.sortValue||u.innerText,"int"===r?d=parseInt(d):"float"===r?d=parseFloat(d):"date"===r&&(d=new Date(d)),v.values.push(d);l.sort("string"===r?"asc"===s?(e,t)=>(""+e.values[a]).localeCompare(t.values[a]):(e,t)=>-(""+e.values[a]).localeCompare(t.values[a]):"asc"===s?(e,t)=>isNaN(e.values[a])||isNaN(t.values[a])?isNaN(e.values[a])?isNaN(t.values[a])?0:-1:1:e.values[a]<t.values[a]?-1:e.values[a]>t.values[a]?1:0:(e,t)=>isNaN(e.values[a])||isNaN(t.values[a])?isNaN(e.values[a])?isNaN(t.values[a])?0:1:-1:e.values[a]<t.values[a]?1:e.values[a]>t.values[a]?-1:0);const N=document.createDocumentFragment();return l.forEach(e=>N.appendChild(e.tr)),t.querySelector("tbody").replaceChildren(N),setTimeout(()=>is_sorting_process_on=!1,delay),!0}Node.prototype.tsortable=function(){this.querySelectorAll("thead th[data-sort], thead td[data-sort]").forEach(e=>e.onclick=tablesort)};
</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
		renderTVTable(searchResults))
}

func tvProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if tvJobRunning {
		fmt.Fprintf(w, `<progress value="%d" max= "%d"/>`, numberOfTVProcessed, totalTV)
	}
}

func renderTVTable(collection []types.SearchResults) (tableRows string) {
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Title</strong></th><th data-sort="int"><strong>Blu-ray Seasons</strong></th><th data-sort="int"><strong>4K-ray Seasons</strong></th><th><strong>Disc</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range collection {
		tableRows += fmt.Sprintf(
			`<tr><td><a href=%q target="_blank">%s [%v]</a></td><td>%d</td><td>%d</td>`,
			collection[i].SearchURL, collection[i].PlexTVShow.Title, collection[i].PlexTVShow.Year,
			collection[i].MatchesBluray, collection[i].Matches4k)
		if collection[i].MatchesBluray+collection[i].Matches4k > 0 {
			tableRows += "<td>"
			for j := range collection[i].TVSearchResults {
				for _, series := range collection[i].TVSearchResults[j].Series {
					if slices.Contains(series.Format, types.DiskBluray) || slices.Contains(series.Format, types.Disk4K) {
						tableRows += fmt.Sprintf(
							`<a href=%q target="_blank">Season %d:%v`,
							series.URL, series.Number, series.Format)
						tableRows += "</a><br>"
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
