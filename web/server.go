package web

import (
	_ "embed"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/tphoney/plex-lookup/types"
)

var (
	//go:embed index.html
	indexPage       string
	port            string = "9090"
	PlexInformation types.PlexInformation
)

func StartServer() {
	// find the local IP address
	ipAddress := GetOutboundIP()
	fmt.Printf("Starting server on http://%s:%s\n", ipAddress.String(), port)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/plex", plexHandler)
	http.HandleFunc("/plexlibraries", processPlexLibrariesHTML)
	http.HandleFunc("/saveplex", plexSaveHandler)

	http.HandleFunc("/movies", moviesHandler)
	http.HandleFunc("/processmovies", processMoviesHTML)
	http.HandleFunc("/progress", progressBarHTML)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil) //nolint: gosec
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
