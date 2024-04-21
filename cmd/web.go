package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/web"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Starts the web server",
	Long:  `Starts the web server, that allows you to compare plex to amazon/cinema paradiso.`,
	Run: func(_ *cobra.Command, _ []string) {
		startServer()
	},
}

func startServer() {
	plexInformation := types.PlexInformation{}
	// read environment variables
	plexInformation.IP = os.Getenv("PLEX_IP")
	plexInformation.MovieLibraryID = os.Getenv("PLEX_MOVIE_LIBRARY_ID")
	plexInformation.Token = os.Getenv("PLEX_TOKEN")

	web.StartServer(plexInformation)
}
