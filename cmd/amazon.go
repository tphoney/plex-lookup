package cmd

import (
	"fmt"

	"github.com/tphoney/plex-lookup/amazon"
	"github.com/tphoney/plex-lookup/types"

	"github.com/spf13/cobra"
)

var amazonCmd = &cobra.Command{
	Use:   "amazon",
	Short: "Compare Movies/TV in your plex library with amazon",
	Long: `This command will compare movies in your plex library with amazon and print out the 
movies that of higher quality than DVD.`,
	Run: func(_ *cobra.Command, _ []string) {
		performAmazonLookup()
	},
}

func performAmazonLookup() {
	initializeFlags()
	if libraryType == types.PlexMovieType {
		plexMovies := initializePlexMovies()
		// lets search movies in amazon
		for _, movie := range plexMovies {
			movieResult, err := amazon.SearchAmazonMovie(movie, "")
			if err != nil {
				fmt.Printf("Error searching for movie %s: %s\n", movieResult.PlexMovie.Title, err)
				continue
			}
			// if hit, and contains any format that isnt dvd, print the movie
			for _, individualResult := range movieResult.MovieSearchResults {
				if individualResult.BestMatch && (individualResult.Format == types.DiskBluray || individualResult.Format == types.Disk4K) {
					fmt.Printf("%s %v: %s\n", movieResult.PlexMovie.Title, movieResult.PlexMovie.Year, individualResult.URL)
				}
			}
		}
	}
}
