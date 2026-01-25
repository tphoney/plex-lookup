package music

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/tphoney/plex-lookup/musicbrainz"
	"github.com/tphoney/plex-lookup/plex"
	"github.com/tphoney/plex-lookup/spotify"
	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/utils"
)

var (
	//go:embed music.html
	musicPage string

	spotifyToken string
)

const (
	lookupTypeMusicBrainz = "musicbrainz"
	lookupTypeSpotify     = "spotify"
	minArtistsForEstimate = 50
)

type MusicConfig struct {
	Config     *types.Configuration
	JobTracker types.JobTracker
}

func MusicHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("music").Parse(musicPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render music page", http.StatusInternalServerError)
		return
	}
}

func (c MusicConfig) PlaylistHTML(w http.ResponseWriter, _ *http.Request) {
	playlistHTML := `<fieldset id="playlist">
	 <label for="All">
		 <input type="radio" id="playlist" name="playlist" value="all" checked />
		 All: dont use a playlist. (SLOW, only use for small libraries)
	 </label>`
	playlists, _ := plex.GetPlaylists(c.Config.PlexIP, c.Config.PlexToken, c.Config.PlexMusicLibraryID)
	fmt.Println("Playlists:", len(playlists))
	for i := range playlists {
		playlistHTML += fmt.Sprintf(
			`<label for=%q>
			<input type="radio" id="playlist" name="playlist" value=%q/>
			%s</label>`,
			playlists[i].Title, playlists[i].RatingKey, playlists[i].Title)
	}

	playlistHTML += `</fieldset>`
	fmt.Fprint(w, playlistHTML)
}

// validateLookupConfig checks if the lookup service is properly configured.
func (c MusicConfig) validateLookupConfig(w http.ResponseWriter, r *http.Request, lookup string) bool {
	if lookup == lookupTypeMusicBrainz {
		if c.Config.MusicBrainzURL == "" {
			fmt.Fprintf(w, `<div class="container"><b>MusicBrainz URL is not set</b>. Please set in <a href="/settings">settings.</a></div>`)
			return false
		}
	}
	if lookup == lookupTypeSpotify {
		if c.Config.SpotifyClientID == "" || c.Config.SpotifyClientSecret == "" {
			fmt.Fprintf(w, `<div class="container"><b>Spotify Client ID or Secret is not set</b>. Please set in <a href="/settings">settings.</a></div>`)
			return false
		}
		if spotifyToken == "" {
			var err error
			spotifyToken, err = spotify.SpotifyOAuthToken(r.Context(), c.Config.SpotifyClientID, c.Config.SpotifyClientSecret)
			if err != nil {
				fmt.Fprintf(w, `<div class="alert alert-danger" role="alert">Failed to get Spotify OAuth token<br>%s</div>`, err.Error())
				return false
			}
		}
	}
	return true
}

func (c MusicConfig) ProcessHTML(w http.ResponseWriter, r *http.Request) {
	tracker := c.JobTracker
	if tracker == nil {
		http.Error(w, "Job tracker not available", http.StatusInternalServerError)
		return
	}

	playlist := r.FormValue("playlist")
	lookup := r.FormValue("lookup")
	if !c.validateLookupConfig(w, r, lookup) {
		return
	}

	// Get artists from plex
	var plexMusic []types.PlexMusicArtist
	if playlist == "all" {
		plexMusic = plex.AllMusicArtists(c.Config.PlexIP, c.Config.PlexToken, c.Config.PlexMusicLibraryID)
	} else {
		plexMusic = plex.GetArtistsFromPlaylist(c.Config.PlexIP, c.Config.PlexToken, playlist)
	}

	// Limit for non-local musicbrainz
	totalArtists := len(plexMusic)
	if lookup == lookupTypeMusicBrainz && strings.Contains(c.Config.MusicBrainzURL, "musicbrainz.org") {
		if len(plexMusic) > minArtistsForEstimate {
			plexMusic = plexMusic[:minArtistsForEstimate]
			totalArtists = minArtistsForEstimate
		}
	}

	// Create job
	jobID, ctx := tracker.CreateJob("music", totalArtists)

	// Return initial progress bar
	fmt.Fprintf(w, `<div hx-get="/progress/%s" hx-trigger="every 250ms" class="container" id="progress">
		<progress value="0" max="%d"></progress></div>`, jobID, totalArtists)

	// Start processing in goroutine
	go func() {
		startTime := time.Now()
		var artistsSearchResults []types.MusicSearchResponse

		progressFunc := func(current int, phase string) {
			tracker.UpdateProgress(jobID, current, phase)
		}

		switch lookup {
		case "musicbrainz":
			for i := range plexMusic {
				// Check cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}
				fmt.Print(".")
				searchResult, _ := musicbrainz.SearchMusicBrainzArtist(ctx, &plexMusic[i], c.Config.MusicBrainzURL)
				artistsSearchResults = append(artistsSearchResults, searchResult)
				progressFunc(i+1, "Searching MusicBrainz")
			}
		default:
			// Search spotify
			artistsSearchResults = spotify.GetArtistsInParallel(ctx, progressFunc, plexMusic, spotifyToken)
			artistsSearchResults = spotify.GetAlbumsInParallel(ctx, progressFunc, artistsSearchResults, spotifyToken)
			// sanitise album titles
			artistsSearchResults = sanitizeAlbumTitles(artistsSearchResults)
		}

		// Generate results table HTML
		tableHTML := renderArtistAlbumsTable(artistsSearchResults)
		resultsHTML := fmt.Sprintf(`<table class="table-sortable" hx-boost="true">%s</tbody></table>
		<script>document.querySelector('.table-sortable').tsortable()</script>`, tableHTML)

		tracker.MarkComplete(jobID, resultsHTML)
		fmt.Printf("\nProcessed %d artists in %v\n", len(plexMusic), time.Since(startTime))
	}()
}

func renderArtistAlbumsTable(artistsSearchResults []types.MusicSearchResponse) (tableRows string) {
	searchResults := filterMusicSearchResults(artistsSearchResults)
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Artist</strong></th><th data-sort="int">First album</th><th data-sort="int">Last album</th><th data-sort="int"><strong>Owned Albums</strong></th><th data-sort="int"><strong>Wanted Albums</strong></th></tr></thead><tbody>`
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			tableRows += fmt.Sprintf(`<tr><td><a href=%q target="_blank">%s</a></td><td>%d</td><td>%d</td><td>%s</td><td>%s</td></tr>`,
				searchResults[i].MusicSearchResults[0].URL,
				searchResults[i].Name,
				searchResults[i].MusicSearchResults[0].FirstAlbumYear,
				searchResults[i].MusicSearchResults[0].LastAlbumYear,
				renderAccordian(searchResults[i].MusicSearchResults[0].OwnedAlbums),
				renderAccordian(stringsFromFoundAlbums(searchResults[i].MusicSearchResults[0].FoundAlbums)))
		}
	}
	return tableRows // Return the generated HTML for table rows
}

func stringsFromFoundAlbums(albums []types.MusicAlbumSearchResult) []string {
	var titles []string
	for _, album := range albums {
		entry := fmt.Sprintf("<a href=%q target=\"_blank\">%s (%s)</a>", album.URL, album.Title, album.Year)
		titles = append(titles, entry)
	}
	return titles
}

func renderAccordian(s []string) string {
	retval := fmt.Sprintf(`<details><summary>%d</summary><ul>`, len(s))
	for _, item := range s {
		retval += fmt.Sprintf(`<li>%s</li>`, item)
	}
	retval += `</ul></details>`
	return retval
}

func sanitizeAlbumTitles(artistsSearchResults []types.MusicSearchResponse) []types.MusicSearchResponse {
	for i := range artistsSearchResults {
		if len(artistsSearchResults[i].MusicSearchResults) > 0 {
			for j := range artistsSearchResults[i].MusicSearchResults[0].FoundAlbums {
				artistsSearchResults[i].MusicSearchResults[0].FoundAlbums[j].SanitizedTitle =
					utils.SanitizedAlbumTitle(artistsSearchResults[i].MusicSearchResults[0].FoundAlbums[j].Title)
			}
		}
	}
	return artistsSearchResults
}

func filterMusicSearchResults(searchResults []types.MusicSearchResponse) []types.MusicSearchResponse {
	searchResults = markOwnedAlbumsInSearchResult(searchResults)
	searchResults = removeOlderSearchedAlbums(searchResults)
	return searchResults
}

func removeOlderSearchedAlbums(searchResults []types.MusicSearchResponse) []types.MusicSearchResponse {
	filteredResults := make([]types.MusicSearchResponse, 0)
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			filteredAlbums := make([]types.MusicAlbumSearchResult, 0)
			filteredAlbums = append(filteredAlbums, searchResults[i].MusicSearchResults[0].FoundAlbums...)
			searchResults[i].MusicSearchResults[0].FoundAlbums = filteredAlbums
			filteredResults = append(filteredResults, searchResults[i])
		}
	}
	return filteredResults
}

func markOwnedAlbumsInSearchResult(searchResults []types.MusicSearchResponse) []types.MusicSearchResponse {
	for i := range searchResults {
		var searchIDsToRemove []string
		if len(searchResults[i].MusicSearchResults) > 0 {
			// iterate over plex albums
			for _, plexAlbum := range searchResults[i].Albums {
				searchResults[i].MusicSearchResults[0].OwnedAlbums =
					append(searchResults[i].MusicSearchResults[0].OwnedAlbums, plexAlbum.Title+" ("+plexAlbum.Year+")")
				// make a deep copy of the albums in the search results
				albumsCopy := append([]types.MusicAlbumSearchResult(nil), searchResults[i].MusicSearchResults[0].FoundAlbums...)
				searchIDsToRemove = append(searchIDsToRemove, findMatchingAlbumFromSearch(plexAlbum, albumsCopy)...)
			}
			searchResults[i].MusicSearchResults[0].FoundAlbums = removeOwnedFromSearchResults(searchResults[i].MusicSearchResults[0].FoundAlbums, searchIDsToRemove)
		}
	}
	// sort the owned albums by year
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			sort.Slice(searchResults[i].MusicSearchResults[0].OwnedAlbums, func(a, b int) bool {
				yearA := strings.Split(searchResults[i].MusicSearchResults[0].OwnedAlbums[a], " (")
				yearB := strings.Split(searchResults[i].MusicSearchResults[0].OwnedAlbums[b], " (")
				yearAInt, _ := strconv.Atoi(strings.TrimSuffix(yearA[1], ")"))
				yearBInt, _ := strconv.Atoi(strings.TrimSuffix(yearB[1], ")"))
				return yearAInt > yearBInt // Sort by year descending
			})
		}
		// Calculate the first and last album year for each artist
		// NB we are sorting by the owned plex albums.
		if len(searchResults[i].MusicSearchResults) > 0 {
			youngestAlbumYear := 9999
			oldestAlbumYear := 0
			for j := range searchResults[i].Albums {
				year, err := strconv.Atoi(searchResults[i].Albums[j].Year)
				if err != nil {
					continue // Skip albums with invalid year
				}
				if year < youngestAlbumYear {
					youngestAlbumYear = year
				}
				if year > oldestAlbumYear {
					oldestAlbumYear = year
				}
			}
			searchResults[i].MusicSearchResults[0].FirstAlbumYear = youngestAlbumYear
			searchResults[i].MusicSearchResults[0].LastAlbumYear = oldestAlbumYear
		}
	}
	return searchResults
}

func findMatchingAlbumFromSearch(plexAlbum types.PlexMusicAlbum, original []types.MusicAlbumSearchResult) (foundIDs []string) {
	plexSanitizedTitle := utils.SanitizedAlbumTitle(plexAlbum.Title)
	sanitizedAlbumTitles := make([]string, 0)
	for _, searchAlbum := range original {
		sanitizedAlbumTitles = append(sanitizedAlbumTitles, searchAlbum.SanitizedTitle)
	}
	matches := fuzzy.RankFind(plexSanitizedTitle, sanitizedAlbumTitles)
	sort.Sort(matches)

	for _, match := range matches {
		if match.Distance < 0 {
			continue // Skip negative scores
		}
		// Find the index of the matched album in the original slice
		for j := range original {
			if original[j].SanitizedTitle == match.Target {
				foundIDs = append(foundIDs, original[j].ID)
			}
		}
	}
	keys := make(map[string]bool)
	cleaned := []string{}

	for _, entry := range foundIDs {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			cleaned = append(cleaned, entry)
		}
	}
	return cleaned
}

func removeOwnedFromSearchResults(original []types.MusicAlbumSearchResult, toRemove []string) []types.MusicAlbumSearchResult {
	if len(toRemove) == 0 {
		return original
	}
	cleaned := make([]types.MusicAlbumSearchResult, 0, len(original))
	// Iterate over the original search results and remove any albums that match the IDs in toRemove
	for _, album := range original {
		// If the album ID is not in toRemove, keep the search result
		if !slices.Contains(toRemove, album.ID) {
			// Add the album to the cleaned search result
			cleaned = append(cleaned, album)
		}
	}
	// print the number of albums removed
	fmt.Printf(" %d albums removed from search results\n", len(original)-len(cleaned))
	return cleaned
}
