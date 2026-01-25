package cmd

import (
	"fmt"

	"github.com/tphoney/plex-lookup/cinemaparadiso"
	"github.com/tphoney/plex-lookup/types"

	"github.com/spf13/cobra"
)

var cinemaParadisoCmd = &cobra.Command{
	Use:   "cinema-paradiso",
	Short: "Compare movies/TV in your plex library with cinema paradiso",
	Long: `This command will compare movies in your plex library with cinema paradiso and print out the 
movies that of higher quality than DVD.`,
	Run: func(_ *cobra.Command, _ []string) {
		performCinemaParadisoLookup()
	},
}

func performCinemaParadisoLookup() {
	initializeFlags()
	if libraryType == types.PlexMovieType {
		plexMovies := initializePlexMovies()
		// lets search movies in cinemaparadiso
		searchResults := cinemaparadiso.MoviesInParallel(plexMovies)
		// if hit, and contains any format that isnt dvd, print the movie
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
