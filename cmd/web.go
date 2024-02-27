package cmd

import (
	"github.com/spf13/cobra"
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
	web.StartServer()
}
