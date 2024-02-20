package cmd

import (
	"fmt"
	"slices"
	"tphoney/plex-lookup/cinemaparadiso"
	"tphoney/plex-lookup/plex"
	"tphoney/plex-lookup/types"

	"github.com/spf13/cobra"
)

var cinemaParadisoCmd = &cobra.Command{
	Use:   "cinema-paradiso",
	Short: "Compare movies in your plex library with cinema paradiso",
	Long: `This command will compare movies in your plex library with cinema paradiso and print out the 
movies that of higher quality than DVD.`,
	Run: func(_ *cobra.Command, _ []string) {
		performCinemaParadisoLookup()
	},
}

func performCinemaParadisoLookup() {
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
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "sd", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "480", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "576", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryID, "720", plexToken)...)

	fmt.Printf("There are a total of %d movies in the library.", len(allMovies))

	// lets search movies in cinemaparadiso
	for _, movie := range allMovies {
		hit, url, formats := cinemaparadiso.SearchCinemaParadiso(movie.Title, movie.Year)
		// if hit, and contains any format that isnt dvd, print the movie
		if hit && !slices.Contains(formats, "DVD") {
			fmt.Printf("\n%s %v: %s\n", movie.Title, formats, url)
		}
	}
}
