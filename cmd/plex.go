package cmd

import (
	"fmt"

	"github.com/tphoney/plex-lookup/plex"

	"github.com/spf13/cobra"
)

var plexCmd = &cobra.Command{
	Use:   "plex-libraries",
	Short: "List out the libraries in your plex server",
	Long:  `This command will list out the libraries in your plex server.`,
	Run: func(_ *cobra.Command, _ []string) {
		getPlexLibraries()
	},
}

func getPlexLibraries() {
	ipAddress := rootCmd.PersistentFlags().Lookup("plexIP").Value.String()
	plexToken := rootCmd.PersistentFlags().Lookup("plexToken").Value.String()
	// validate the input
	if ipAddress == "" {
		panic("plexIP Address is required")
	}
	if plexToken == "" {
		panic("plexToken is required")
	}

	libraries, err := plex.GetPlexLibraries(ipAddress, plexToken)
	if err != nil {
		panic(err)
	}
	for _, library := range libraries {
		fmt.Printf("Title: %s\n", library.Title)
		fmt.Printf("Type: %s\n", library.Type)
		fmt.Printf("ID: %s\n", library.ID)
		fmt.Println()
	}
}
