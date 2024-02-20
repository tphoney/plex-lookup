package cmd

import (
	"tphoney/plex-lookup/plex"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(plexCmd)
}

var plexCmd = &cobra.Command{
	Use:   "plex-libraries",
	Short: "list out the libraries in your plex server",
	Long:  `This command will list out the libraries in your plex server.`,
	Run: func(cmd *cobra.Command, args []string) {
		getPlexLibraries()
	},
}

func getPlexLibraries() {
	ipAddress := rootCmd.PersistentFlags().Lookup("plexIP").Value.String()
	plexToken := rootCmd.PersistentFlags().Lookup("plexToken").Value.String()

	plex.GetPlexLibraries(ipAddress, plexToken)
}
