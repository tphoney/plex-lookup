package cmd

import (
	"fmt"
	"slices"

	"github.com/tphoney/plex-lookup/amazon"

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
	// validate the inputs
	allMovies := initializePlexMovies()

	// lets search movies in amazon
	for _, movie := range allMovies {
		hit, url, formats := amazon.SearchAmazon(movie.Title, movie.Year)
		// if hit, and contains any format that isnt dvd, print the movie
		if hit && (slices.Contains(formats, "Blu-ray") || slices.Contains(formats, "4K Blu-ray")) {
			fmt.Printf("%s %v: %s\n", movie.Title, formats, url)
		}
	}
}
