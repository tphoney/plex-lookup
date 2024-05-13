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

type SimilarArtistsResponse struct {
	Artists []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
}

func SearchSpotifyArtist(plexArtist *types.PlexMusicArtist, clientID, clientSecret string, ch chan<- *types.SearchResults) {
	// context with a timeout of 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(lookupTimeout))
	defer cancel()
	searchResults := types.SearchResults{}
	searchResults.PlexMusicArtist = *plexArtist
	// get oauth token
	err := SpotifyOauthToken(ctx, clientID, clientSecret)
	if err != nil {
		fmt.Printf("SearchSpotifyArtist: unable to get oauth token: %s\n", err.Error())
		ch <- &searchResults
		return
	}
	urlEncodedArtist := url.QueryEscape(plexArtist.Name)
	artistURL := fmt.Sprintf("%s/search?q=%s&type=artist&limit=10", spotifyAPIURL, urlEncodedArtist)
	client := &http.Client{
		Timeout: time.Second * lookupTimeout,
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, artistURL, http.NoBody)
	bearer := fmt.Sprintf("Bearer %s", oauthToken)
	req.Header.Add("Authorization", bearer)
	var response *http.Response
	for {
		response, err = client.Do(req)
		if err != nil {
			response.Body.Close()
			fmt.Printf("lookupArtist: get failed from spotify: %s\n", err.Error())
			ch <- &searchResults
			return
		}
		if response.StatusCode == http.StatusTooManyRequests {
			// rate limited
			wait := response.Header.Get("Retry-After")
			waitSeconds, _ := strconv.Atoi(wait)
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

func SearchSpotifyAlbum(m *types.SearchResults, clientID, clientSecret string, ch chan<- *types.SearchResults) {
	// get oauth token
	err := SpotifyOauthToken(context.Background(), clientID, clientSecret)
	if err != nil {
		fmt.Printf("SearchSpotifyAlbums: unable to get oauth token: %s\n", err.Error())
		ch <- m
		return
	}
	if len(m.MusicSearchResults) == 0 {
		// no artist found for the plex artist
		fmt.Printf("SearchSpotifyAlbums: no artist found for %s\n", m.PlexMusicArtist.Name)
		ch <- m
		return
	}
	albumURL := fmt.Sprintf("%s/artists/%s/albums?include_groups=album&limit=50&", spotifyAPIURL, m.MusicSearchResults[0].ID)
	client := &http.Client{
		Timeout: time.Second * lookupTimeout,
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, albumURL, http.NoBody)
	bearer := fmt.Sprintf("Bearer %s", oauthToken)
	req.Header.Add("Authorization", bearer)
	var response *http.Response
	for {
		response, err = client.Do(req)
		if err != nil {
			response.Body.Close()
			fmt.Printf("lookupArtistAlbums: get failed from spotify: %s\n", err.Error())
			ch <- m
			return
		}
		if response.StatusCode == http.StatusTooManyRequests {
			wait := response.Header.Get("Retry-After")
			waitSeconds, _ := strconv.Atoi(wait)
			if waitSeconds > lookupTimeout {
				fmt.Printf("lookupArtistAlbums: rate limited for %d seconds\n", waitSeconds)
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

	m.MusicSearchResults[0].Albums = albums
	ch <- m
}

func FindSimilarArtists(ownedArtists []types.MusicArtistSearchResult, clientID, clientSecret string) (
	similar map[string]types.MusicSimilarArtistResult, err error) {
	// make a map of SearchSimilarArtists
	similar = make(map[string]types.MusicSimilarArtistResult, len(ownedArtists))
	// get oauth token
	err = SpotifyOauthToken(context.Background(), clientID, clientSecret)
	if err != nil {
		return similar, fmt.Errorf("FindSimilarArtists: unable to get oauth token: %s", err.Error())
	}
	for i := range ownedArtists {
		// check if the artist is in the similar map already
		artist, ok := similar[ownedArtists[i].ID]
		if !ok {
			// add the artist to the map
			similar[ownedArtists[i].ID] = types.MusicSimilarArtistResult{
				Name:            ownedArtists[i].Name,
				URL:             ownedArtists[i].URL,
				Owned:           true,
				SimilarityCount: 0,
			}
		} else {
			// set owned to true
			artist.Owned = true
			similar[ownedArtists[i].ID] = artist
		}
		// get the similar artists
		similarArtists, err := SearchSpotifySimilarArtist(ownedArtists[i].ID, clientID, clientSecret)
		if err != nil {
			fmt.Printf("FindSimilarArtists: unable to get similar artists: %s\n", err.Error())
			continue
		}
		// iterate through the similar artists, if they are not in the owned artists, add them to the similar artists
		for j := range similarArtists.Artists {
			if _, ok := similar[similarArtists.Artists[j].ID]; !ok {
				similar[similarArtists.Artists[j].ID] = types.MusicSimilarArtistResult{
					Name:            similarArtists.Artists[j].Name,
					URL:             fmt.Sprintf("https://open.spotify.com/artist/%s", similarArtists.Artists[j].ID),
					Owned:           false,
					SimilarityCount: 1,
				}
			} else {
				// increment the similarity count
				artist := similar[similarArtists.Artists[j].ID]
				artist.SimilarityCount++
				similar[similarArtists.Artists[j].ID] = artist
			}
		}
	}
	return similar, nil
}

func SearchSpotifySimilarArtist(artistID, clientID, clientSecret string) (similar SimilarArtistsResponse, err error) {
	err = SpotifyOauthToken(context.Background(), clientID, clientSecret)
	if err != nil {
		return similar, fmt.Errorf("spotifyLookupArtistSimilar: unable to get oauth token: %s", err.Error())
	}
	similarArtistURL := fmt.Sprintf("%s/artists/%s/related-artists", spotifyAPIURL, artistID)
	client := &http.Client{
		Timeout: time.Second * lookupTimeout,
	}
	req, httpErr := http.NewRequestWithContext(context.Background(), http.MethodGet, similarArtistURL, http.NoBody)
	if httpErr != nil {
		return similar, fmt.Errorf("spotifyLookupArtistSimilar: get failed from spotify: %s", httpErr.Error())
	}
	bearer := fmt.Sprintf("Bearer %s", oauthToken)
	req.Header.Add("Authorization", bearer)
	response, err := client.Do(req)
	if err != nil {
		return similar, fmt.Errorf("spotifyLookupArtistSimilar: get failed from spotify: %s", err.Error())
	}
	if response.StatusCode == http.StatusTooManyRequests {
		return similar, fmt.Errorf("lookupArtist: rate limited by spotify")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return similar, fmt.Errorf("spotifyLookupArtistSimilar: unable to parse response from spotify: %s", err.Error())
	}

	var similarArtistsResponse SimilarArtistsResponse
	jsonErr := json.Unmarshal(body, &similarArtistsResponse)
	if jsonErr != nil {
		return similar, fmt.Errorf("spotifyLookupArtistSimilar: unable to unmarshal response from spotify: %s", jsonErr.Error())
	}
	similar.Artists = similarArtistsResponse.Artists
	return similarArtistsResponse, nil
}

// function that gets an oauth token from spotify from the client id and secret
func SpotifyOauthToken(ctx context.Context, clientID, clientSecret string) (err error) {
	// get oauth token
	if oauthToken != "" {
		return nil
	}
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
		return fmt.Errorf("spotifyOauthToken: get failed from spotify: %s", err.Error())
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("spotifyOauthToken: unable to read response from spotify: %s", err.Error())
	}
	var oauthResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	err = json.Unmarshal(body, &oauthResponse)
	if err != nil {
		return fmt.Errorf("getOauthToken: unable to parse response from spotify: %s", err.Error())
	}
	oauthToken = oauthResponse.AccessToken
	return nil
}

func artistStringMatcher(dbName, webName string) bool {
	// check if the names are the same, ignoring case and punctuation
	dbName = strings.ToLower(dbName)
	webName = strings.ToLower(webName)
	dbName = strings.NewReplacer(".", "", " ", "", ",", "", "\"", "").Replace(dbName)
	webName = strings.NewReplacer(".", "", " ", "", ",", "", "\"", "").Replace(webName)

	return dbName == webName
}
