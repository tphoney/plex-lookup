package musicbrainz

import (
	"fmt"

	"github.com/michiwend/gomusicbrainz"
)

// library docs https://github.com/michiwend/gomusicbrainz/blob/master/release_group.go
// api docs https://musicbrainz.org/doc/Release_Group/Type#Secondary_types
// example artist https://musicbrainz.org/artist/83d91898-7763-47d7-b03b-b92132375c47

const (
	// MusicBrainzURL is the URL for the MusicBrainz API
	musicBrainzURL = "https://musicbrainz.org/ws/2"
	agent          = "plex-lookup"
	agentVersion   = "0.0.1"
	lookupLimit    = 100
)

type MusicBrainzArtist struct {
	Name string
	ID   string
}

type MusicBrainzAlbum struct {
	Title string
	ID    string
	Year  string
}

func SearchMusicBrainzArtist(artistName string) (artist MusicBrainzArtist, err error) {
	client, err := gomusicbrainz.NewWS2Client(
		musicBrainzURL, agent, agentVersion, "")

	if err != nil {
		return artist, err
	}
	resp, _ := client.SearchArtist(artistName, -1, -1)

	if len(resp.Artists) > 0 {
		artist = MusicBrainzArtist{
			Name: resp.Artists[0].Name,
			ID:   fmt.Sprintf("%v", resp.Artists[0].ID),
		}
	} else {
		err = fmt.Errorf("artist not found")
	}
	return artist, err
}

func SearchMusicBrainzAlbums(artistID string) (albums []MusicBrainzAlbum, err error) {
	client, err := gomusicbrainz.NewWS2Client(
		musicBrainzURL, agent, agentVersion, "")

	if err != nil {
		return albums, err
	}

	queryURL := fmt.Sprintf("arid:%v AND primarytype:album AND status:official AND type:album",
		artistID)
	resp, _ := client.SearchReleaseGroup(queryURL, lookupLimit, -1)
	for _, release := range resp.ReleaseGroups {
		if release.Type == "Album" {
			year := release.FirstReleaseDate.Year()
			albums = append(albums, MusicBrainzAlbum{
				Title: release.Title,
				ID:    fmt.Sprintf("%v", release.ID),
				Year:  fmt.Sprintf("%v", year),
			})
		}
	}

	return albums, err
}
