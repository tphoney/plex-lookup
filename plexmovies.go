package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
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
		Guid                  string `xml:"guid,attr"`
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

func getMovies(ipAddress, libraryId, resolution, plexToken string) (movieList []Movie) {
	url := fmt.Sprintf("http://%s:32400/library/sections/%s/resolution/%s", ipAddress, libraryId, resolution)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("X-Plex-Token", plexToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	movieList = extractMovies(string(body))
	fmt.Printf("Movies at resolution %s: %v\n", resolution, movieList)
	return movieList
}

func extractMovies(xmlString string) (movieList []Movie) {
	var container MediaContainer
	err := xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	for _, video := range container.Video {
		movieList = append(movieList, Movie{Title: video.Title, Year: video.Year})
	}
	return movieList
}
