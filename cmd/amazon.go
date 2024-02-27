package cmd

import (
	"fmt"
	"slices"

	"github.com/tphoney/plex-lookup/amazon"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/types"

	"github.com/spf13/cobra"
)

var amazonCmd = &cobra.Command{
	Use:   "amazon",
	Short: "Compare movies in your plex library with amazon",
	Long: `This command will compare movies in your plex library with amazon and print out the 
movies that of higher quality than DVD.`,
	Run: func(_ *cobra.Command, _ []string) {
		performAmazonLookup()
	},
}

func performAmazonLookup() {
	ipAddress := rootCmd.PersistentFlags().Lookup("plexIP").Value.String()
	libraryID := rootCmd.PersistentFlags().Lookup("plexLibraryID").Value.String()
	plexToken := rootCmd.PersistentFlags().Lookup("plexToken").Value.String()

	// validate the inputs
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
	//allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "sd", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "480", plexToken)...)
	//allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "576", plexToken)...)
	//allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "720", plexToken)...)

	fmt.Printf("\nThere are a total of %d movies in the library.\n\nMovies available:\n", len(allMovies))

	// lets search movies in amazon
	for _, movie := range allMovies {
		hit, url, formats := amazon.SearchAmazon(movie.Title, movie.Year)
		// if hit, and contains any format that isnt dvd, print the movie
		if hit && (slices.Contains(formats, "Blu-ray") || slices.Contains(formats, "4K Blu-ray")) {
			fmt.Printf("%s %v: %s\n", movie.Title, formats, url)
		}
	}
}
