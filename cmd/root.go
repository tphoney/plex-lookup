package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	plexIP        string
	plexLibraryID string
	plexToken     string

	rootCmd = &cobra.Command{
		Use:   "plex-lookup",
		Short: "A tool to compare your plex library",
		Long:  `A tool to compare your plex librarys with other physical media rental / purchasing services.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() { //nolint: gochecknoinits
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVar(&plexIP, "plexIP", "", "Plex IP Address")
	rootCmd.PersistentFlags().StringVar(&plexLibraryID, "plexLibraryID", "", "Plex Library ID")
	rootCmd.PersistentFlags().StringVar(&plexToken, "plexToken", "", "Plex Token")

	rootCmd.AddCommand(amazonCmd)
	rootCmd.AddCommand(cinemaParadisoCmd)
	rootCmd.AddCommand(plexCmd)
	rootCmd.AddCommand(versionCmd)
}
