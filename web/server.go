package web

import (
	"embed"

	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/web/movies"
	"github.com/tphoney/plex-lookup/web/music"
	"github.com/tphoney/plex-lookup/web/settings"
	"github.com/tphoney/plex-lookup/web/tv"
)

var (
	//go:embed index.html
	indexPage string

	//go:embed static/*
	staticFS embed.FS

	port   string = "9090"
	config *types.Configuration
)

func StartServer(startingConfig *types.Configuration) {
	config = startingConfig
	// find the local IP address
	ipAddress := GetOutboundIP()
	fmt.Printf("Starting server on http://%s:%s\n", ipAddress.String(), port)
	mux := http.NewServeMux()

	// serve static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	mux.HandleFunc("/settings", settings.SettingsHandler)
	mux.HandleFunc("/settings/plexlibraries", settings.ProcessPlexLibrariesHTML)
	mux.HandleFunc("/settings/plexinfook", settings.SettingsConfig{Config: config}.PlexInformationOKHTML)

	mux.HandleFunc("/movies", movies.MoviesHandler)
	mux.HandleFunc("/moviesprocess", movies.MoviesConfig{Config: config}.ProcessHTML)
	mux.HandleFunc("/moviesprogress", movies.ProgressBarHTML)

	mux.HandleFunc("/tv", tv.TVHandler)
	mux.HandleFunc("/tvprocess", tv.TVConfig{Config: config}.ProcessHTML)
	mux.HandleFunc("/tvprogress", tv.ProgressBarHTML)

	mux.HandleFunc("/music", music.MusicHandler)
	mux.HandleFunc("/musicprocess", music.MusicConfig{Config: config}.ProcessHTML)
	mux.HandleFunc("/musicprogress", music.ProgressBarHTML)

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/settings/save", settingsSaveHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux) //nolint: gosec
	if err != nil {
		fmt.Printf("Failed to start server on port %s: %s\n", port, err)
		panic(err)
	}
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("index").Parse(indexPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render index", http.StatusInternalServerError)
		return
	}
}

func settingsSaveHandler(w http.ResponseWriter, r *http.Request) {
	oldConfig := config
	// Retrieve form fields (replace with proper values)
	config.PlexIP = r.FormValue("plexIP")
	config.PlexToken = r.FormValue("plexToken")
	config.PlexMovieLibraryID = r.FormValue("plexMovieLibraryID")
	config.PlexTVLibraryID = r.FormValue("plexTVLibraryID")
	config.PlexMusicLibraryID = r.FormValue("plexMusicLibraryID")
	config.AmazonRegion = r.FormValue("amazonRegion")
	config.MusicBrainzURL = r.FormValue("musicBrainzURL")
	config.SpotifyClientID = r.FormValue("spotifyClientID")
	config.SpotifyClientSecret = r.FormValue("spotifyClientSecret")
	fmt.Fprint(w, `<h2>Saved!</h2><a href="/">Back</a>`)
	fmt.Printf("Saved Settings\nold\n%+v\nnew\n%+v\n", oldConfig, config)
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
