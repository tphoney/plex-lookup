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
	"sync/atomic"
	"time"

	"github.com/sourcegraph/conc/iter"
	"github.com/tphoney/plex-lookup/types"
)

const (
	spotifyAPIURL      = "https://api.spotify.com/v1"
	lookupTimeout      = 10
	spotifyConcurrency = 2
)

var (
	numberOfArtistsProcessed atomic.Int32
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
	numberOfArtistsProcessed.Store(0)
	mapper := iter.Mapper[types.PlexMusicArtist, types.MusicSearchResponse]{
		MaxGoroutines: spotifyConcurrency,
	}
	artistsSearchResults := mapper.Map(plexArtists, func(artist *types.PlexMusicArtist) types.MusicSearchResponse {
		result := searchSpotifyArtistValue(artist, token)
		numberOfArtistsProcessed.Add(1)
		fmt.Print(".")
		return result
	})
	numberOfArtistsProcessed.Store(0)
	return artistsSearchResults
}

func GetAlbumsInParallel(artistsSearchResults []types.MusicSearchResponse, token string) []types.MusicSearchResponse {
	numberOfArtistsProcessed.Store(0)
	mapper := iter.Mapper[types.MusicSearchResponse, types.MusicSearchResponse]{
		MaxGoroutines: spotifyConcurrency,
	}
	enrichedArtistSearchResults := mapper.Map(artistsSearchResults, func(result *types.MusicSearchResponse) types.MusicSearchResponse {
		res := searchSpotifyAlbumValue(result, token)
		numberOfArtistsProcessed.Add(1)
		fmt.Print(".")
		return res
	})
	numberOfArtistsProcessed.Store(0)
	return enrichedArtistSearchResults
}

// searchSpotifyArtistValue is a value-returning version for use with iter.Map
func searchSpotifyArtistValue(plexArtist *types.PlexMusicArtist, token string) types.MusicSearchResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(lookupTimeout))
	defer cancel()
	searchResults := types.MusicSearchResponse{}
	searchResults.PlexMusicArtist = *plexArtist
	urlEncodedArtist := url.QueryEscape(plexArtist.Name)
	artistURL := fmt.Sprintf("%s/search?q=%s&type=artist&limit=10", spotifyAPIURL, urlEncodedArtist)
	body, err := makeRequest(artistURL, token, ctx)
	if err != nil {
		fmt.Printf("lookupArtist: unable to read response from spotify: %s\n", err.Error())
		return searchResults
	}
	var artistResponse ArtistResponse
	jsonErr := json.Unmarshal(body, &artistResponse)
	if jsonErr != nil {
		fmt.Printf("lookupArtist: unable to parse response from spotify: %s\n", jsonErr.Error())
		return searchResults
	}
	for i := range artistResponse.Artists.Items {
		if artistStringMatcher(plexArtist.Name, artistResponse.Artists.Items[i].Name) {
			searchResults.MusicSearchResults = append(searchResults.MusicSearchResults, types.MusicArtistSearchResult{
				Name: artistResponse.Artists.Items[i].Name,
				ID:   artistResponse.Artists.Items[i].ID,
				URL:  artistResponse.Artists.Items[i].ExternalUrls.Spotify,
			})
			break
		}
	}
	return searchResults
}

// searchSpotifyAlbumValue is a value-returning version for use with iter.Map
func searchSpotifyAlbumValue(m *types.MusicSearchResponse, token string) types.MusicSearchResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(lookupTimeout))
	defer cancel()
	result := *m
	if len(result.MusicSearchResults) == 0 {
		fmt.Printf("SearchSpotifyAlbums: no artist found for %v\n", result.PlexMusicArtist)
		return result
	}
	albumURL := fmt.Sprintf("%s/artists/%s/albums?include_groups=album&limit=50&", spotifyAPIURL, result.MusicSearchResults[0].ID)
	body, err := makeRequest(albumURL, token, ctx)
	if err != nil {
		fmt.Printf("lookupArtistAlbums: unable to parse response from spotify: %s\n", err.Error())
		return result
	}
	var albumsResponse AlbumsResponse
	_ = json.Unmarshal(body, &albumsResponse)
	albums := make([]types.MusicAlbumSearchResult, 0)
	for i := range albumsResponse.Items {
		year := strings.Split(albumsResponse.Items[i].ReleaseDate, "-")[0]
		albums = append(albums, types.MusicAlbumSearchResult{
			Title: albumsResponse.Items[i].Name,
			ID:    albumsResponse.Items[i].ID,
			URL:   albumsResponse.Items[i].ExternalUrls.Spotify,
			Year:  year,
		})
	}
	result.MusicSearchResults[0].FoundAlbums = albums
	return result
}

func GetJobProgress() int {
	return int(numberOfArtistsProcessed.Load())
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
