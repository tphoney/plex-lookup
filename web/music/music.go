package music

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
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

	numberOfArtistsProcessed int  = 0
	artistsJobRunning        bool = false
	totalArtists             int  = 0

	plexMusic             []types.PlexMusicArtist
	artistsSearchResults  []types.SearchResult
	similarArtistsResults map[string]types.MusicSimilarArtistResult
	spotifyToken          string
	lookup                string
	lookupType            string
)

const (
	albumReleaseYearCutoff int    = 5
	spotifyString          string = "spotify"
)

type MusicConfig struct {
	Config *types.Configuration
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

// nolint: lll, nolintlint
func (c MusicConfig) ProcessHTML(w http.ResponseWriter, r *http.Request) {
	playlist := r.FormValue("playlist")
	lookup = r.FormValue("lookup")
	if lookup == "musicbrainz" {
		if c.Config.MusicBrainzURL == "" {
			fmt.Fprintf(w, `<div class="container"><b>MusicBrainz URL is not set</b>. Please set in <a href="/settings">settings.</a></div>`)
			return
		}
	}
	if lookup == spotifyString {
		if c.Config.SpotifyClientID == "" || c.Config.SpotifyClientSecret == "" {
			fmt.Fprintf(w, `<div class="container"><b>Spotify Client ID or Secret is not set</b>. Please set in <a href="/settings">settings.</a></div>`)
			return
		}
		if spotifyToken == "" {
			var err error
			spotifyToken, err = spotify.SpotifyOAuthToken(r.Context(), c.Config.SpotifyClientID, c.Config.SpotifyClientSecret)
			if err != nil {
				fmt.Fprintf(w, `<div class="alert alert-danger" role="alert">Failed to get Spotify OAuth token<br>%s</div>`, err.Error())
				return
			}
		}
	}
	lookupType = r.FormValue("lookuptype")
	// only get the artists from plex once
	if playlist == "all" {
		plexMusic = plex.AllMusicArtists(c.Config.PlexIP, c.Config.PlexToken, c.Config.PlexMusicLibraryID)
	} else {
		plexMusic = plex.GetArtistsFromPlaylist(c.Config.PlexIP, c.Config.PlexToken, playlist)
	}
	//nolint: gocritic
	// plexMusic = plexMusic[:30]
	//lint: gocritic
	var searchResult types.SearchResult
	artistsJobRunning = true
	numberOfArtistsProcessed = 0
	totalArtists = len(plexMusic) - 1

	fmt.Fprintf(w, `<div hx-get="/musicprogress" hx-trigger="every 100ms" hx-boost="true" class="container" id="progress">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)
	startTime := time.Now()
	if lookupType == "missingalbums" {
		switch lookup {
		case "musicbrainz":
			// limit the number of artists to 50 for nonlocal musicbrainz instances
			if strings.Contains(c.Config.MusicBrainzURL, "musicbrainz.org") {
				plexMusic = plexMusic[:50]
				totalArtists = len(plexMusic) - 1
			}
			go func() {
				for i := range plexMusic {
					fmt.Print(".")
					searchResult, _ = musicbrainz.SearchMusicBrainzArtist(&plexMusic[i], c.Config.MusicBrainzURL)
					artistsSearchResults = append(artistsSearchResults, searchResult)
					numberOfArtistsProcessed = i
				}
				artistsJobRunning = false
			}()
		default:
			// search spotify
			go func() {
				artistsSearchResults = spotify.GetArtistsInParallel(plexMusic, spotifyToken)
				artistsSearchResults = spotify.GetAlbumsInParallel(artistsSearchResults, spotifyToken)
				// sanitize album titles
				artistsSearchResults = sanitizeAlbumTitles(artistsSearchResults)
				artistsJobRunning = false
			}()
		}
	} else {
		switch lookup {
		case spotifyString:
			go func() {
				artistsSearchResults = spotify.GetArtistsInParallel(plexMusic, spotifyToken)
				similarArtistsResults = spotify.GetSimilarArtistsInParallel(artistsSearchResults, spotifyToken)
				totalArtists = numberOfArtistsProcessed
				artistsJobRunning = false
			}()
			fmt.Println("Searching Spotify for similar artists")
		default:
			fmt.Fprintf(w, `<div class="alert alert-danger" role="alert">Similar Artist search is not available for this lookup provider</div>`)
		}
	}
	fmt.Printf("Processed %d artists in %v\n", totalArtists, time.Since(startTime))
}

func ProgressBarHTML(w http.ResponseWriter, _ *http.Request) {
	if lookup == spotifyString {
		numberOfArtistsProcessed = spotify.GetJobProgress()
	}
	if artistsJobRunning {
		fmt.Fprintf(w, `<div hx-get="/musicprogress" hx-trigger="every 100ms" class="container" hx-boost="true" id="progress" hx-swap="outerHTML">
		<progress value="%d" max= "%d"/></div>`, numberOfArtistsProcessed, totalArtists)
	} else {
		tableContents := ""
		if lookupType == "missingalbums" {
			tableContents = renderArtistAlbumsTable()
		} else {
			tableContents = renderSimilarArtistsTable()
		}
		fmt.Fprintf(w,
			`<table class="table-sortable" hx-boost="true">%s</tbody></table>
		</script><script>document.querySelector('.table-sortable').tsortable()</script>`,
			tableContents)
		// reset variables
		numberOfArtistsProcessed = 0
		totalArtists = 0
	}
}

func renderArtistAlbumsTable() (tableRows string) {
	searchResults := filterMusicSearchResults(artistsSearchResults)
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Artist</strong></th><th data-sort="int"><strong>Owned Albums</strong></th><th data-sort="int"><strong>Wanted Albums</strong></th><th><strong>Album</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			// use accordians https://picocss.com/docs/accordion
			tableRows += fmt.Sprintf(`<tr><td><a href=%q target="_blank">%s</a></td><td>%s</td><td>%d</td><td><ul>`,
				searchResults[i].MusicSearchResults[0].URL,
				searchResults[i].Name,
				renderAccordian(searchResults[i].MusicSearchResults[0].OwnedAlbums),
				len(searchResults[i].MusicSearchResults[0].FoundAlbums))
			for j := range searchResults[i].MusicSearchResults[0].FoundAlbums {
				tableRows += fmt.Sprintf(`<li><a href=%q target="_blank">%s</a> (%s)</li>`,
					searchResults[i].MusicSearchResults[0].FoundAlbums[j].URL,
					searchResults[i].MusicSearchResults[0].FoundAlbums[j].Title,
					searchResults[i].MusicSearchResults[0].FoundAlbums[j].Year)
			}
			tableRows += "</ul></td></tr>"
		}
	}
	return tableRows // Return the generated HTML for table rows
}

func renderAccordian(s []string) string {
	retval := fmt.Sprintf(`<details><summary>%d</summary>`, len(s))
	for _, item := range s {
		retval += fmt.Sprintf(`<p>%s</p>`, item)
	}
	retval += `</details>`
	return retval
}

func renderSimilarArtistsTable() (tableRows string) {
	tableRows = `<thead><tr><th data-sort="string"><strong>Plex Artist</strong></th><th data-sort="string"><strong>Owned</strong></th><th data-sort="int"><strong>Similarity Count</strong></th></tr></thead><tbody>` //nolint: lll
	for i := range similarArtistsResults {
		ownedString := "No"
		if similarArtistsResults[i].Owned {
			ownedString = "Yes"
		}
		tableRows += fmt.Sprintf(`<tr><td><a href=%q target="_blank">%s</a></td><td>%s</td><td>%d</td></tr>`,
			similarArtistsResults[i].URL,
			similarArtistsResults[i].Name,
			ownedString,
			similarArtistsResults[i].SimilarityCount)
	}
	return tableRows // Return the generated HTML for table rows
}

func sanitizeAlbumTitles(artistsSearchResults []types.SearchResult) []types.SearchResult {
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

func filterMusicSearchResults(searchResults []types.SearchResult) []types.SearchResult {
	searchResults = markOwnedAlbumsInSearchResult(searchResults)
	searchResults = removeOlderSearchedAlbums(searchResults)
	return searchResults
}

func removeOlderSearchedAlbums(searchResults []types.SearchResult) []types.SearchResult {
	cutoffYear := time.Now().Year() - albumReleaseYearCutoff
	filteredResults := make([]types.SearchResult, 0)
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			filteredAlbums := make([]types.MusicAlbumSearchResult, 0)
			for _, album := range searchResults[i].MusicSearchResults[0].FoundAlbums {
				albumYear, _ := strconv.Atoi(album.Year)
				if albumYear >= cutoffYear {
					filteredAlbums = append(filteredAlbums, album)
				}
			}
			searchResults[i].MusicSearchResults[0].FoundAlbums = filteredAlbums
			filteredResults = append(filteredResults, searchResults[i])
		}
	}
	return filteredResults
}

func markOwnedAlbumsInSearchResult(searchResults []types.SearchResult) []types.SearchResult {
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			// iterate over plex albums
			for _, plexAlbum := range searchResults[i].Albums {
				searchResults[i].MusicSearchResults[0].OwnedAlbums = append(searchResults[i].MusicSearchResults[0].OwnedAlbums, plexAlbum.Title+" ("+plexAlbum.Year+")")
				// make a deep copy of the albums in the search results
				copy := make([]types.MusicAlbumSearchResult, len(searchResults[i].MusicSearchResults[0].FoundAlbums))
				copy = append(copy, searchResults[i].MusicSearchResults[0].FoundAlbums...)
				removeMatchingAlbumFromSearch(plexAlbum, copy)
			}
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
			// remove any albums that are not within one year of the plex album
		}
	}
	return searchResults
}

func removeMatchingAlbumFromSearch(plexAlbum types.PlexMusicAlbum, original []types.MusicAlbumSearchResult) (foundID string) {
	for j := range original {
		if utils.WithinOneYear(plexAlbum.Year, original[j].Year) {
			// if the album is owned, mark it as such
			original[j].WithinOneYear = true
		}
	}
	// iterate over the search albums and remove any that are not within one year
	for j := len(original) - 1; j >= 0; j-- {
		if !original[j].WithinOneYear {
			original = append(original[:j], original[j+1:]...)
		}
	}
	sanitizedAlbumTitles := make([]string, 0)
	for _, searchAlbum := range original {
		sanitizedAlbumTitles = append(sanitizedAlbumTitles, searchAlbum.SanitizedTitle)
	}
	matches := fuzzy.RankFind(utils.SanitizedAlbumTitle(plexAlbum.Title), sanitizedAlbumTitles)
	sort.Sort(matches)

	for _, match := range matches {
		if match.Distance < 0 {
			continue // Skip negative scores
		}
		// Find the index of the matched album in the original slice
		for j := range original {
			if original[j].SanitizedTitle == match.Target {
				foundID = original[j].ID
				break
			}
		}
	}
	fmt.Printf("Found ID: %s for album: %s\n", foundID, plexAlbum.Title)
	return foundID
}

func cleanAlbums(original, toRemove []types.MusicAlbumSearchResult) []types.MusicAlbumSearchResult {
	cleaned := make([]types.MusicAlbumSearchResult, 0)
	for _, album := range original {
		found := false
		for _, remove := range toRemove {
			if album.Title == remove.Title && album.Year == remove.Year {
				found = true
				break
			}
		}
		if !found {
			cleaned = append(cleaned, album)
		}
	}
	return cleaned
}
