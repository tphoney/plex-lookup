package plex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	types "github.com/tphoney/plex-lookup/types"
)

type MediaContainer struct {
	XMLName             xml.Name `xml:"MediaContainer"`
	Text                string   `xml:",chardata"`
	Size                string   `xml:"size,attr"`
	AllowSync           string   `xml:"allowSync,attr"`
	Art                 string   `xml:"art,attr"`
	Identifier          string   `xml:"identifier,attr"`
	LibrarySectionID    string   `xml:"librarySectionID,attr"`
	LibrarySectionTitle string   `xml:"librarySectionTitle,attr"`
	LibrarySectionUUID  string   `xml:"librarySectionUUID,attr"`
	MediaTagPrefix      string   `xml:"mediaTagPrefix,attr"`
	MediaTagVersion     string   `xml:"mediaTagVersion,attr"`
	Thumb               string   `xml:"thumb,attr"`
	Title1              string   `xml:"title1,attr"`
	Title2              string   `xml:"title2,attr"`
	ViewGroup           string   `xml:"viewGroup,attr"`
	ViewMode            string   `xml:"viewMode,attr"`
	Video               []struct {
		Text                  string `xml:",chardata"`
		RatingKey             string `xml:"ratingKey,attr"`
		Key                   string `xml:"key,attr"`
		GUID                  string `xml:"guid,attr"`
		Studio                string `xml:"studio,attr"`
		Type                  string `xml:"type,attr"`
		Title                 string `xml:"title,attr"`
		ContentRating         string `xml:"contentRating,attr"`
		Summary               string `xml:"summary,attr"`
		Rating                string `xml:"rating,attr"`
		AudienceRating        string `xml:"audienceRating,attr"`
		ViewCount             string `xml:"viewCount,attr"`
		LastViewedAt          string `xml:"lastViewedAt,attr"`
		Year                  string `xml:"year,attr"`
		Tagline               string `xml:"tagline,attr"`
		Thumb                 string `xml:"thumb,attr"`
		Art                   string `xml:"art,attr"`
		Duration              string `xml:"duration,attr"`
		OriginallyAvailableAt string `xml:"originallyAvailableAt,attr"`
		AddedAt               string `xml:"addedAt,attr"`
		UpdatedAt             string `xml:"updatedAt,attr"`
		AudienceRatingImage   string `xml:"audienceRatingImage,attr"`
		PrimaryExtraKey       string `xml:"primaryExtraKey,attr"`
		RatingImage           string `xml:"ratingImage,attr"`
		Media                 struct {
			Text            string `xml:",chardata"`
			ID              string `xml:"id,attr"`
			Duration        string `xml:"duration,attr"`
			Bitrate         string `xml:"bitrate,attr"`
			Width           string `xml:"width,attr"`
			Height          string `xml:"height,attr"`
			AspectRatio     string `xml:"aspectRatio,attr"`
			AudioChannels   string `xml:"audioChannels,attr"`
			AudioCodec      string `xml:"audioCodec,attr"`
			VideoCodec      string `xml:"videoCodec,attr"`
			VideoResolution string `xml:"videoResolution,attr"`
			Container       string `xml:"container,attr"`
			VideoFrameRate  string `xml:"videoFrameRate,attr"`
			VideoProfile    string `xml:"videoProfile,attr"`
			Part            struct {
				Text         string `xml:",chardata"`
				ID           string `xml:"id,attr"`
				Key          string `xml:"key,attr"`
				Duration     string `xml:"duration,attr"`
				File         string `xml:"file,attr"`
				Size         string `xml:"size,attr"`
				Container    string `xml:"container,attr"`
				VideoProfile string `xml:"videoProfile,attr"`
			} `xml:"Part"`
		} `xml:"Media"`
		Genre []struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Genre"`
		Country struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Country"`
		Director struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Director"`
		Writer []struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Writer"`
		Role []struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Role"`
	} `xml:"Video"`
}

type LibraryContainer struct {
	XMLName   xml.Name `xml:"MediaContainer"`
	Text      string   `xml:",chardata"`
	Size      string   `xml:"size,attr"`
	AllowSync string   `xml:"allowSync,attr"`
	Title1    string   `xml:"title1,attr"`
	Directory []struct {
		Text             string `xml:",chardata"`
		AllowSync        string `xml:"allowSync,attr"`
		Art              string `xml:"art,attr"`
		Composite        string `xml:"composite,attr"`
		Filters          string `xml:"filters,attr"`
		Refreshing       string `xml:"refreshing,attr"`
		Thumb            string `xml:"thumb,attr"`
		Key              string `xml:"key,attr"`
		Type             string `xml:"type,attr"`
		Title            string `xml:"title,attr"`
		Agent            string `xml:"agent,attr"`
		Scanner          string `xml:"scanner,attr"`
		Language         string `xml:"language,attr"`
		UUID             string `xml:"uuid,attr"`
		UpdatedAt        string `xml:"updatedAt,attr"`
		CreatedAt        string `xml:"createdAt,attr"`
		ScannedAt        string `xml:"scannedAt,attr"`
		Content          string `xml:"content,attr"`
		Directory        string `xml:"directory,attr"`
		ContentChangedAt string `xml:"contentChangedAt,attr"`
		Hidden           string `xml:"hidden,attr"`
		Location         []struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Path string `xml:"path,attr"`
		} `xml:"Location"`
	} `xml:"Directory"`
}

type Filter struct {
	Name     string
	Value    string
	Modifier string
}

func GetPlexMovies(ipAddress, libraryID, plexToken, resolution string, filters []Filter) (movieList []types.Movie) {
	url := fmt.Sprintf("http://%s:32400/library/sections/%s", ipAddress, libraryID)
	if resolution == "" {
		url += "/all"
	} else {
		url += fmt.Sprintf("/resolution/%s", resolution)
	}

	for i := range filters {
		if i == 0 {
			url += "?"
		} else {
			url += "&"
		}
		url += fmt.Sprintf("%s%s%s", filters[i].Name, filters[i].Modifier, filters[i].Value)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, http.NoBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return movieList
	}

	req.Header.Set("X-Plex-Token", plexToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return movieList
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return movieList
	}

	movieList = extractMovies(string(body))
	fmt.Printf("Movies at resolution %s: %v\n", resolution, movieList)
	return movieList
}

func GetPlexLibraries(ipAddress, plexToken string) (libraryList []types.PlexLibrary, err error) {
	url := fmt.Sprintf("http://%s:32400/library/sections", ipAddress)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, http.NoBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return libraryList, err
	}

	req.Header.Set("X-Plex-Token", plexToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return libraryList, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return libraryList, err
	}

	libraryList, err = extractLibraries(string(body))
	return libraryList, err
}

func extractLibraries(xmlString string) (libraryList []types.PlexLibrary, err error) {
	var container LibraryContainer
	err = xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return libraryList, err
	}

	for i := range container.Directory {
		libraryList = append(libraryList, types.PlexLibrary{Title: container.Directory[i].Title,
			ID: container.Directory[i].Key, Type: container.Directory[i].Type})
	}
	return libraryList, nil
}

func extractMovies(xmlString string) (movieList []types.Movie) {
	var container MediaContainer
	err := xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	for i := range container.Video {
		movieList = append(movieList, types.Movie{
			Title: container.Video[i].Title, Year: container.Video[i].Year, DateAdded: container.Video[i].AddedAt})
	}
	return movieList
}
