package cmd

import (
	"context"
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
		searchResults := amazon.MoviesInParallel(context.Background(), nil, plexMovies, "", amazonRegion)
		for i := range searchResults {
			for _, individualResult := range searchResults[i].MovieSearchResults {
				if individualResult.BestMatch && (individualResult.Format == types.DiskBluray || individualResult.Format == types.Disk4K) {
					fmt.Printf("%s - %s (%s): %s\n", searchResults[i].Title, individualResult.Format,
						searchResults[i].Year, individualResult.URL)
				}
			}
		}
	}
}
