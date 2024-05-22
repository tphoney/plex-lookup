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
	config := types.Configuration{}
	// read environment variables
	config.PlexIP = os.Getenv("PLEX_IP")
	config.PlexMovieLibraryID = os.Getenv("PLEX_MOVIE_LIBRARY_ID")
	config.PlexTVLibraryID = os.Getenv("PLEX_TV_LIBRARY_ID")
	config.PlexMusicLibraryID = os.Getenv("PLEX_MUSIC_LIBRARY_ID")
	config.PlexToken = os.Getenv("PLEX_TOKEN")
	config.AmazonRegion = os.Getenv("AMAZON_REGION")
	if config.AmazonRegion == "" {
		config.AmazonRegion = "uk"
	}
	config.MusicBrainzURL = os.Getenv("MUSICBRAINZ_URL")
	if config.MusicBrainzURL == "" {
		config.MusicBrainzURL = "https://musicbrainz.org/ws/2"
	}
	config.SpotifyClientID = os.Getenv("SPOTIFY_CLIENT_ID")
	config.SpotifyClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")

	web.StartServer(&config)
}
