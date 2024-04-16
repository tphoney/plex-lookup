package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"
)

var (
	// Used for flags.
	plexIP             string
	plexMovieLibraryID string
	plexToken          string
	libraryType        string

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
	rootCmd.PersistentFlags().StringVar(&plexMovieLibraryID, "plexMovieLibraryID", "", "Plex Library ID")
	rootCmd.PersistentFlags().StringVar(&plexToken, "plexToken", "", "Plex Token")
	// add modifier flags
	rootCmd.PersistentFlags().StringVar(&libraryType, "type", types.PlexMovieType, "Library Type (Movie, TV)")

	rootCmd.AddCommand(amazonCmd)
	rootCmd.AddCommand(cinemaParadisoCmd)
	rootCmd.AddCommand(plexCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(webCmd)
}

func initializeFlags() {
	plexIP = rootCmd.PersistentFlags().Lookup("plexIP").Value.String()
	plexMovieLibraryID = rootCmd.PersistentFlags().Lookup("plexMovieLibraryID").Value.String()
	plexToken = rootCmd.PersistentFlags().Lookup("plexToken").Value.String()
	libraryType = rootCmd.PersistentFlags().Lookup("type").Value.String()

	if plexIP == "" {
		panic("plexIP Address is required")
	}
	if plexMovieLibraryID == "" {
		panic("plexMovieLibraryID is required")
	}
	if plexToken == "" {
		panic("plexToken is required")
	}
	if libraryType != types.PlexMovieType && libraryType != "TV" {
		panic("type of library must be Movie or TV")
	}
}

func initializePlexMovies() []types.PlexMovie {
	var allMovies []types.PlexMovie
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, "sd", nil)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, "480", nil)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, "576", nil)...)
	allMovies = append(allMovies, plex.GetPlexMovies(plexIP, plexMovieLibraryID, plexToken, "720", nil)...)

	fmt.Printf("\nThere are a total of %d movies in the library.\n\nMovies available:\n", len(allMovies))
	return allMovies
}
