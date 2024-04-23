package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed plex.html
	plexPage string
)

func plexHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("plex").Parse(plexPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render plex page", http.StatusInternalServerError)
		return
	}
}

func processPlexLibrariesHTML(w http.ResponseWriter, r *http.Request) {
	// Retrieve form fields (replace with proper values)
	plexIP := r.FormValue("plexIP")
	plexToken := r.FormValue("plexToken")

	libraries, err := plex.GetPlexLibraries(plexIP, plexToken)
	if err != nil {
		http.Error(w, "Failed to get plex libraries", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, renderLibraries(libraries))
}

func plexInformationOKHTML(w http.ResponseWriter, _ *http.Request) {
	// remove the warning in the html if the plex information is set
	if PlexInformation.IP == "" || PlexInformation.Token == "" || PlexInformation.MovieLibraryID == "" {
		fmt.Fprint(w, `<h1><a href="/plex"> Enter your plex information here</a></h1>`)
	} else {
		fmt.Fprint(w, ``)
	}
}

func plexSaveHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve form fields (replace with proper values)
	plexIP := r.FormValue("plexIP")
	plexToken := r.FormValue("plexToken")
	plexMovieLibraryID := r.FormValue("plexMovieLibraryID")
	plexTVLibraryID := r.FormValue("plexTVLibraryID")
	// validate form fields
	PlexInformation = types.PlexInformation{
		IP:             plexIP,
		Token:          plexToken,
		MovieLibraryID: plexMovieLibraryID,
		TVLibraryID:    plexTVLibraryID,
	}
	fmt.Fprint(w, `<h2>Saved!</h2><a href="/">Back</a>`)
	fmt.Printf("Saved plex information: %+v\n", PlexInformation)
}

func renderLibraries(libraries []types.PlexLibrary) string {
	html := `<h2 class="container">Libraries</h2><table><thead><tr><th>Title</th><th>Type</th><th>ID</th></tr></thead><tbody>`
	for _, library := range libraries {
		html += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`, library.Title, library.Type, library.ID)
	}
	html += `</tbody></table>
	<input type="text" placeholder="Plex Movie Library Section ID" name="plexMovieLibraryID"id="plexMovieLibraryID">
	<input type="text" placeholder="Plex TV Series Library Section ID" name="plexTVLibraryID"id="plexTVLibraryID">
	<button hx-post="/plexsave" hx-include="#plexMovieLibraryID,#plexIP,#plexToken" hx-swap="outerHTML">Save</button>`
	return html
}
