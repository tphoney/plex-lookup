package cmd

import (
	"fmt"
	"slices"

	"github.com/tphoney/plex-lookup/cinemaparadiso"

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
	allMovies := initializePlexMovies()
	// lets search movies in cinemaparadiso
	for _, movie := range allMovies {
		hit, url, formats := cinemaparadiso.SearchCinemaParadiso(movie.Title, movie.Year)
		// if hit, and contains any format that isnt dvd, print the movie
		if hit && (slices.Contains(formats, "Blu-ray") || slices.Contains(formats, "4K Blu-ray")) {
			fmt.Printf("%s %v: %s\n", movie.Title, formats, url)
		}
	}
}
