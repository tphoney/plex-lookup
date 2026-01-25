package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tphoney/plex-lookup/types"
)

const (
	spotifyAPIURL  = "https://api.spotify.com/v1"
	lookupTimeout  = 10
	spotifyThreads = 2
)

var (
	numberOfArtistsProcessed int
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

type SimilarArtistsResponse struct {
	Artists []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
}

func GetArtistsInParallel(plexArtists []types.PlexMusicArtist, token string) []types.MusicSearchResponse {
	numberOfArtistsProcessed = 0
	ch := make(chan *types.MusicSearchResponse, len(plexArtists))
	semaphore := make(chan struct{}, spotifyThreads)
	for i := range len(plexArtists) {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchSpotifyArtist(&plexArtists[i], token, ch)
		}(i)
	}
	// gather results
	artistsSearchResults := make([]types.MusicSearchResponse, 0, len(plexArtists))
	for range len(plexArtists) {
		result := <-ch
		artistsSearchResults = append(artistsSearchResults, *result)
		fmt.Print(".")
		numberOfArtistsProcessed++
	}
	numberOfArtistsProcessed = 0
	return artistsSearchResults
}

func GetAlbumsInParallel(artistsSearchResults []types.MusicSearchResponse, token string) []types.MusicSearchResponse {
	numberOfArtistsProcessed = 0
	ch := make(chan *types.MusicSearchResponse, len(artistsSearchResults))
	semaphore := make(chan struct{}, spotifyThreads)
	for i := range artistsSearchResults {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			searchSpotifyAlbum(&artistsSearchResults[i], token, ch)
		}(i)
	}
	// gather results
	enrichedArtistSearchResults := make([]types.MusicSearchResponse, 0)
	for range artistsSearchResults {
		result := <-ch
		enrichedArtistSearchResults = append(enrichedArtistSearchResults, *result)
		fmt.Print(".")
		numberOfArtistsProcessed++
	}
	numberOfArtistsProcessed = 0
	return enrichedArtistSearchResults
}

func GetJobProgress() int {
	return numberOfArtistsProcessed
}

func searchSpotifyArtist(plexArtist *types.PlexMusicArtist, token string, ch chan<- *types.MusicSearchResponse) {
	// context with a timeout of 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(lookupTimeout))
	defer cancel()
	searchResults := types.MusicSearchResponse{}
	searchResults.PlexMusicArtist = *plexArtist
	urlEncodedArtist := url.QueryEscape(plexArtist.Name)
	artistURL := fmt.Sprintf("%s/search?q=%s&type=artist&limit=10", spotifyAPIURL, urlEncodedArtist)
	body, err := makeRequest(artistURL, token, ctx)
	if err != nil {
		fmt.Printf("lookupArtist: unable to read response from spotify: %s\n", err.Error())
		ch <- &searchResults
		return
	}
	var artistResponse ArtistResponse
	jsonErr := json.Unmarshal(body, &artistResponse)
	if jsonErr != nil {
		fmt.Printf("lookupArtist: unable to parse response from spotify: %s\n", jsonErr.Error())
		ch <- &searchResults
		return
	}
	for i := range artistResponse.Artists.Items {
		if artistStringMatcher(plexArtist.Name, artistResponse.Artists.Items[i].Name) {
			// only get the first match
			searchResults.MusicSearchResults = append(searchResults.MusicSearchResults, types.MusicArtistSearchResult{
				Name: artistResponse.Artists.Items[i].Name,
				ID:   artistResponse.Artists.Items[i].ID,
				URL:  artistResponse.Artists.Items[i].ExternalUrls.Spotify,
			})
			// only get the first match
			break
		}
	}
	ch <- &searchResults
}

func searchSpotifyAlbum(m *types.MusicSearchResponse, token string, ch chan<- *types.MusicSearchResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(lookupTimeout))
	defer cancel()
	// get oauth token
	if len(m.MusicSearchResults) == 0 {
		// no artist found for the plex artist
		fmt.Printf("SearchSpotifyAlbums: no artist found for %v\n", m.PlexMusicArtist)
		ch <- m
		return
	}
	albumURL := fmt.Sprintf("%s/artists/%s/albums?include_groups=album&limit=50&", spotifyAPIURL, m.MusicSearchResults[0].ID)
	body, err := makeRequest(albumURL, token, ctx)
	if err != nil {
		fmt.Printf("lookupArtistAlbums: unable to parse response from spotify: %s\n", err.Error())
		ch <- m
		return
	}
	var albumsResponse AlbumsResponse
	_ = json.Unmarshal(body, &albumsResponse)

	albums := make([]types.MusicAlbumSearchResult, 0)
	for i := range albumsResponse.Items {
		// convert "2022-06-03" to "2022"
		year := strings.Split(albumsResponse.Items[i].ReleaseDate, "-")[0]
		albums = append(albums, types.MusicAlbumSearchResult{
			Title: albumsResponse.Items[i].Name,
			ID:    albumsResponse.Items[i].ID,
			URL:   albumsResponse.Items[i].ExternalUrls.Spotify,
			Year:  year,
		})
	}

	m.MusicSearchResults[0].FoundAlbums = albums
	ch <- m
}

// function that gets an oauth token from spotify from the client id and secret
func SpotifyOAuthToken(ctx context.Context, clientID, clientSecret string) (token string, err error) {
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

func makeRequest(inputURL, token string, ctx context.Context) (rawResponse []byte, err error) {
	client := &http.Client{
		Timeout: time.Second * lookupTimeout,
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, inputURL, http.NoBody)
	bearer := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", bearer)
	var response *http.Response
	for {
		response, err = client.Do(req)
		if err != nil {
			response.Body.Close()
			return nil, err
		}
		if response.StatusCode == http.StatusTooManyRequests {
			wait := response.Header.Get("Retry-After")
			waitSeconds, _ := strconv.Atoi(wait)
			if waitSeconds > lookupTimeout {
				fmt.Printf("spotify: rate limited for %d seconds\n", waitSeconds)
			}
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}
		if response.StatusCode == http.StatusOK {
			break
		}
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// check for a 200 status code
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify: status code not OK: %d", response.StatusCode)
	}
	return body, nil
}

func artistStringMatcher(dbName, webName string) bool {
	// check if the names are the same, ignoring case and punctuation
	dbName = strings.ToLower(dbName)
	webName = strings.ToLower(webName)
	dbName = strings.NewReplacer(".", "", " ", "", ",", "", "\"", "").Replace(dbName)
	webName = strings.NewReplacer(".", "", " ", "", ",", "", "\"", "").Replace(webName)

	return dbName == webName
}
