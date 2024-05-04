package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tphoney/plex-lookup/types"
)

const (
	spotifyAPIURL = "https://api.spotify.com/v1"
	lookupTimeout = 10
)

var (
	oauthToken string
)

type ArtistResponse struct {
	Artists struct {
		Href  string `json:"href"`
		Items []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Followers struct {
				Total int64 `json:"total"`
				Href  any   `json:"href"`
			} `json:"followers"`
			Genres []string `json:"genres"`
			Href   string   `json:"href"`
			ID     string   `json:"id"`
			Images []struct {
				Height int64  `json:"height"`
				URL    string `json:"url"`
				Width  int64  `json:"width"`
			} `json:"images"`
			Name       string `json:"name"`
			Popularity int64  `json:"popularity"`
			Type       string `json:"type"`
			URI        string `json:"uri"`
		} `json:"items"`
		Limit    int64  `json:"limit"`
		Next     string `json:"next"`
		Offset   int64  `json:"offset"`
		Previous any    `json:"previous"`
		Total    int64  `json:"total"`
	} `json:"artists"`
}

type AlbumsResponse struct {
	Href  string `json:"href"`
	Items []struct {
		AlbumGroup string `json:"album_group"`
		AlbumType  string `json:"album_type"`
		Artists    []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height int64  `json:"height"`
			URL    string `json:"url"`
			Width  int64  `json:"width"`
		} `json:"images"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		TotalTracks          int64  `json:"total_tracks"`
		Type                 string `json:"type"`
		URI                  string `json:"uri"`
	} `json:"items"`
	Limit    int64 `json:"limit"`
	Next     any   `json:"next"`
	Offset   int64 `json:"offset"`
	Previous any   `json:"previous"`
	Total    int64 `json:"total"`
}

func SearchSpotifyArtist(plexArtist *types.PlexMusicArtist, clientID, clientSecret string) (artist types.SearchResults, err error) {
	artist.PlexMusicArtist = *plexArtist
	// get oauth token
	if oauthToken == "" {
		oauthToken, err = spotifyOauthToken(context.Background(), clientID, clientSecret)
		if err != nil {
			return artist, fmt.Errorf("SearchSpotifyArtist: unable to get oauth token: %s", err.Error())
		}
	}
	urlEncodedArtist := url.QueryEscape(plexArtist.Name)
	artistURL := fmt.Sprintf("%s/search?q=%s&type=artist&limit=10", spotifyAPIURL, urlEncodedArtist)
	client := &http.Client{
		Timeout: time.Second * lookupTimeout,
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, artistURL, http.NoBody)
	bearer := fmt.Sprintf("Bearer %s", oauthToken)
	req.Header.Add("Authorization", bearer)
	response, err := client.Do(req)
	if err != nil {
		return artist, fmt.Errorf("lookupArtist: get failed from spotify: %s", err.Error())
	}
	if response.StatusCode == http.StatusTooManyRequests {
		return artist, fmt.Errorf("lookupArtist: rate limited by spotify")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return artist, fmt.Errorf("lookupArtist: unable to parse response from spotify: %s", err.Error())
	}
	var artistResponse ArtistResponse
	jsonErr := json.Unmarshal(body, &artistResponse)
	if jsonErr != nil {
		return artist, fmt.Errorf("lookupArtist: unable to parse response from spotify: %s", jsonErr.Error())
	}
	for i := range artistResponse.Artists.Items {
		if artistStringMatcher(plexArtist.Name, artistResponse.Artists.Items[i].Name) {
			// only get the first match
			artist.MusicSearchResults = append(artist.MusicSearchResults, types.MusicSearchResult{
				Name: artistResponse.Artists.Items[i].Name,
				ID:   artistResponse.Artists.Items[i].ID,
				URL:  artistResponse.Artists.Items[i].ExternalUrls.Spotify,
			})
			// only get the first match
			break
		}
	}
	if len(artist.MusicSearchResults) == 0 {
		return artist, err
	}
	// get the albums
	artist.MusicSearchResults[0].Albums, err = SearchSpotifyAlbums(artist.MusicSearchResults[0].ID, clientID, clientSecret)
	return artist, nil
}

func SearchSpotifyAlbums(artistID, clientID, clientSecret string) (albums []types.MusicSearchAlbumResult, err error) {
	if oauthToken == "" {
		oauthToken, err = spotifyOauthToken(context.Background(), clientID, clientSecret)
		if err != nil {
			return albums, fmt.Errorf("SearchSpotifyAlbums: unable to get oauth token: %s", err.Error())
		}
	}
	albumURL := fmt.Sprintf("%s/artists/%s/albums?include_groups=album&limit=50&", spotifyAPIURL, artistID)
	client := &http.Client{
		Timeout: time.Second * lookupTimeout,
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, albumURL, http.NoBody)
	bearer := fmt.Sprintf("Bearer %s", oauthToken)
	req.Header.Add("Authorization", bearer)
	response, err := client.Do(req)
	if err != nil {
		return albums, fmt.Errorf("lookupArtistAlbums: get failed from spotify: %s", err.Error())
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return albums, fmt.Errorf("lookupArtistAlbums: unable to parse response from spotify: %s", err.Error())
	}
	var albumsResponse AlbumsResponse
	_ = json.Unmarshal(body, &albumsResponse)

	for i := range albumsResponse.Items {
		albums = append(albums, types.MusicSearchAlbumResult{
			Title: albumsResponse.Items[i].Name,
			ID:    albumsResponse.Items[i].ID,
			URL:   albumsResponse.Items[i].ExternalUrls.Spotify,
			Year:  albumsResponse.Items[i].ReleaseDate,
		})
	}

	return albums, err
}

// function that gets an oauth token from spotify from the client id and secret
func spotifyOauthToken(ctx context.Context, clientID, clientSecret string) (oauth string, err error) {
	// get oauth token
	oauthURL := "https://accounts.spotify.com/api/token"
	client := &http.Client{
		Timeout: time.Second * lookupTimeout,
	}
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, oauthURL, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("spotifyOauthToken: get failed from spotify: %s", err.Error())
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("spotifyOauthToken: unable to read response from spotify: %s", err.Error())
	}
	var oauthResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	err = json.Unmarshal(body, &oauthResponse)
	if err != nil {
		return "", fmt.Errorf("getOauthToken: unable to parse response from spotify: %s", err.Error())
	}
	return oauthResponse.AccessToken, nil
}

func artistStringMatcher(dbName, webName string) bool {
	// check if the names are the same, ignoring case and punctuation
	dbName = strings.ToLower(dbName)
	webName = strings.ToLower(webName)
	dbName = strings.NewReplacer(".", "", " ", "", ",", "", "\"", "").Replace(dbName)
	webName = strings.NewReplacer(".", "", " ", "", ",", "", "\"", "").Replace(webName)

	return dbName == webName
}
