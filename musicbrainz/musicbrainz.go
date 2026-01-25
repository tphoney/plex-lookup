package musicbrainz

import (
	"fmt"
	"strings"
	"time"

	"github.com/michiwend/gomusicbrainz"
	"github.com/tphoney/plex-lookup/types"
)

// library docs https://github.com/michiwend/gomusicbrainz/blob/master/release_group.go
// api docs https://musicbrainz.org/doc/Release_Group/Type#Secondary_types
// example artist https://musicbrainz.org/artist/83d91898-7763-47d7-b03b-b92132375c47

const (
	agent         = "plex-lookup"
	agentVersion  = "0.0.1"
	lookupLimit   = 100
	lookupTimeout = 2
)

func SearchMusicBrainzArtist(plexArtist *types.PlexMusicArtist, musicBrainzURL string) (artist types.MusicSearchResponse, err error) {
	artist.PlexMusicArtist = *plexArtist
	client, err := gomusicbrainz.NewWS2Client(
		musicBrainzURL, agent, agentVersion, "")

	if err != nil {
		return artist, err
	}
	// encode the artist name according to lucene query syntax
	r := strings.NewReplacer(
		"+", `\`,
		"-", `\`,
		"&&", `\`,
		"||", `\`,
		"!", `\`,
		"(", `\`,
		")", `\`,
		"{", `\`,
		"}", `\`,
		"[", `\`,
		"]", `\`,
		"^", `\`,
		`"`, `\`,
		"~", `\`,
		"*", `\`,
		"?", `\`,
		":", `\`,
		`\`, `\`,
		"/", `\`)

	encodedArtist := r.Replace(plexArtist.Name)
	resp, err := client.SearchArtist(encodedArtist, -1, -1)

	if err != nil {
		// check for a 503 error
		if err.Error() == "EOF" {
			fmt.Printf("!")
			time.Sleep(lookupTimeout * time.Second)
			return SearchMusicBrainzArtist(plexArtist, musicBrainzURL)
		}
	}

	for i := range resp.Artists {
		if resp.Artists[i].Name != plexArtist.Name {
			continue
		}
		found := types.MusicArtistSearchResult{
			Name: resp.Artists[i].Name,
			ID:   fmt.Sprintf("%v", resp.Artists[i].ID),
		}
		url := fmt.Sprintf("https://musicbrainz.org/artist/%v", found.ID)
		found.URL = url
		// get the albums
		found.FoundAlbums, _ = SearchMusicBrainzAlbums(found.ID, musicBrainzURL)
		artist.MusicSearchResults = append(artist.MusicSearchResults, found)
		break
	}
	if len(artist.MusicSearchResults) == 0 {
		err = fmt.Errorf("artist not found")
	}
	return artist, err
}

func SearchMusicBrainzAlbums(artistID, musicBrainzURL string) (albums []types.MusicAlbumSearchResult, err error) {
	client, err := gomusicbrainz.NewWS2Client(
		musicBrainzURL, agent, agentVersion, "")

	if err != nil {
		return albums, err
	}

	queryURL := fmt.Sprintf("arid:%v AND status:official AND primarytype:album AND -secondarytype:*",
		artistID)
	resp, err := client.SearchReleaseGroup(queryURL, lookupLimit, -1)
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Printf("!")
			time.Sleep(lookupTimeout * time.Second)
			return SearchMusicBrainzAlbums(artistID, musicBrainzURL)
		}
	}
	for i := range resp.ReleaseGroups {
		if resp.ReleaseGroups[i].Type == "Album" {
			year := resp.ReleaseGroups[i].FirstReleaseDate.Year()
			albums = append(albums, types.MusicAlbumSearchResult{
				Title: resp.ReleaseGroups[i].Title,
				ID:    fmt.Sprintf("%v", resp.ReleaseGroups[i].ID),
				Year:  fmt.Sprintf("%v", year),
				URL:   fmt.Sprintf("https://musicbrainz.org/release-group/%v", resp.ReleaseGroups[i].ID),
			})
		}
	}

	return albums, err
}
