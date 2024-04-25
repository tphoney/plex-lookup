package web

import (
	"embed"

	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed index.html
	indexPage string

	//go:embed static/*
	staticFS embed.FS

	port            string = "9090"
	PlexInformation types.PlexInformation
)

func StartServer(info types.PlexInformation) {
	PlexInformation = info
	// find the local IP address
	ipAddress := GetOutboundIP()
	fmt.Printf("Starting server on http://%s:%s\n", ipAddress.String(), port)
	mux := http.NewServeMux()

	// serve static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	mux.HandleFunc("/plex", plexHandler)
	mux.HandleFunc("/plexlibraries", processPlexLibrariesHTML)
	mux.HandleFunc("/plexinfook", plexInformationOKHTML)
	mux.HandleFunc("/plexsave", plexSaveHandler)

	mux.HandleFunc("/movies", moviesHandler)
	mux.HandleFunc("/processmovies", processMoviesHTML)
	mux.HandleFunc("/progress", progressBarHTML)

	mux.HandleFunc("/tv", tvHandler)
	mux.HandleFunc("/processtv", processTVHTML)
	mux.HandleFunc("/progresstv", tvProgressBarHTML)

	mux.HandleFunc("/", indexHandler)
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
