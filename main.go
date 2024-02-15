package main

import (
	"fmt"
	"os"
	"slices"
	"tphoney/plex-lookup/cinemaparadiso"
	"tphoney/plex-lookup/plex"
	"tphoney/plex-lookup/types"
)

func main() {
	// get environment variables
	ipAddress := os.Getenv("PLEX_IP")
	libraryId := os.Getenv("PLEX_LIBRARY_ID")
	plexToken := os.Getenv("PLEX_TOKEN")
	var allMovies []types.Movie
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryId, "sd", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryId, "480", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryId, "576", plexToken)...)
	allMovies = append(allMovies, plex.GetPlexMovies(ipAddress, libraryId, "720", plexToken)...)

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
