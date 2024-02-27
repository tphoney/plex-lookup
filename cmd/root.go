package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"
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

func initializePlexMovies() []types.Movie {
	ipAddress := rootCmd.PersistentFlags().Lookup("plexIP").Value.String()
	libraryID := rootCmd.PersistentFlags().Lookup("plexLibraryID").Value.String()
	plexToken := rootCmd.PersistentFlags().Lookup("plexToken").Value.String()

	if ipAddress == "" {
		panic("plexIP Address is required")
	}
	if libraryID == "" {
		panic("plexLibraryID is required")
	}
	if plexToken == "" {
		panic("plexToken is required")
	}

	var allMovies []types.Movie
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "sd", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "480", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "576", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "720", plexToken)...)

	fmt.Printf("\nThere are a total of %d movies in the library.\n\nMovies available:\n", len(allMovies))
	return allMovies
}
