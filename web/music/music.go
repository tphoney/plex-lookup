package music

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	artistsSearchResults  []types.SearchResults
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
	var searchResult types.SearchResults
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
			tableRows += fmt.Sprintf(`<tr><td><a href=%q target="_blank">%s</a></td><td>%d</td><td>%d</td><td><ul>`,
				searchResults[i].MusicSearchResults[0].URL,
				searchResults[i].Name,
				searchResults[i].MusicSearchResults[0].OwnedAlbums,
				len(searchResults[i].MusicSearchResults[0].Albums))
			for j := range searchResults[i].MusicSearchResults[0].Albums {
				tableRows += fmt.Sprintf(`<li><a href=%q target="_blank">%s</a> (%s)</li>`,
					searchResults[i].MusicSearchResults[0].Albums[j].URL,
					searchResults[i].MusicSearchResults[0].Albums[j].Title,
					searchResults[i].MusicSearchResults[0].Albums[j].Year)
			}
			tableRows += "</ul></td></tr>"
		}
	}
	return tableRows // Return the generated HTML for table rows
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

func filterMusicSearchResults(searchResults []types.SearchResults) []types.SearchResults {
	searchResults = removeOwnedAlbums(searchResults)
	searchResults = removeOlderSearchedAlbums(searchResults)
	return searchResults
}

func removeOlderSearchedAlbums(searchResults []types.SearchResults) []types.SearchResults {
	cutoffYear := time.Now().Year() - albumReleaseYearCutoff
	filteredResults := make([]types.SearchResults, 0)
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			filteredAlbums := make([]types.MusicAlbumSearchResult, 0)
			for _, album := range searchResults[i].MusicSearchResults[0].Albums {
				albumYear, _ := strconv.Atoi(album.Year)
				if albumYear >= cutoffYear {
					filteredAlbums = append(filteredAlbums, album)
				}
			}
			searchResults[i].MusicSearchResults[0].Albums = filteredAlbums
			filteredResults = append(filteredResults, searchResults[i])
		}
	}
	return filteredResults
}

func removeOwnedAlbums(searchResults []types.SearchResults) []types.SearchResults {
	for i := range searchResults {
		if len(searchResults[i].MusicSearchResults) > 0 {
			albumsToRemove := make([]types.MusicAlbumSearchResult, 0)
			// set the number of owned albums
			searchResults[i].MusicSearchResults[0].OwnedAlbums = len(searchResults[i].Albums)
			// iterate over plex albums
			for _, plexAlbum := range searchResults[i].Albums {
				// iterate over search results
				for _, album := range searchResults[i].MusicSearchResults[0].Albums {
					if utils.CompareAlbumTitles(plexAlbum.Title, album.Title) {
						albumsToRemove = append(albumsToRemove, album)
					}
				}
			}
			searchResults[i].MusicSearchResults[0].Albums = cleanAlbums(searchResults[i].MusicSearchResults[0].Albums, albumsToRemove)
		}
	}
	return searchResults
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
