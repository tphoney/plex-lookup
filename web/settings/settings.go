package settings

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"path"

	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed settings.html
	settingsPage string
)

type SettingsConfig struct {
	Config *types.Configuration
}

func SettingsHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("settings").Parse(settingsPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render settings page", http.StatusInternalServerError)
		return
	}
}

func ProcessPlexLibrariesHTML(w http.ResponseWriter, r *http.Request) {
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

// plexInformationOKHTML remove the warning in the html if the plex information is set
func (c SettingsConfig) PlexInformationOKHTML(w http.ResponseWriter, r *http.Request) {
	currentURL := r.Header.Get("hx-current-url")
	// get the last part of the url

	requestingPage := path.Base(currentURL)
	if c.Config.PlexIP == "" || c.Config.PlexToken == "" {
		fmt.Fprint(w, `<h1><a href="/settings"> Enter your plex token and plex ip</a></h1>`)
	} else {
		switch requestingPage {
		case "movies":
			if c.Config.PlexMovieLibraryID == "" {
				fmt.Fprint(w, `<h1><a href="/settings"> Enter your plex movie library section ID</a></h1>`)
			}
		case "tv":
			if c.Config.PlexTVLibraryID == "" {
				fmt.Fprint(w, `<h1><a href="/settings"> Enter your plex tv library section ID</a></h1>`)
			}
		case "music":
			if c.Config.PlexMusicLibraryID == "" {
				fmt.Fprint(w, `<h1><a href="/settings"> Enter your plex music library section ID</a></h1>`)
			}
		default:
			fmt.Fprint(w, ``)
		}
	}
}

func renderLibraries(libraries []types.PlexLibrary) string {
	html := `<h2 class="container">Libraries</h2><table><thead><tr><th>Title</th><th>Type</th><th>ID</th></tr></thead><tbody>`
	for _, library := range libraries {
		html += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`, library.Title, library.Type, library.ID)
	}
	html += "</tbody></table>"
	return html
}
