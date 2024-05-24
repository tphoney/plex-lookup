package plex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	types "github.com/tphoney/plex-lookup/types"
)

type MovieContainer struct {
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

type MovieDetailContainer struct {
	XMLName             xml.Name `xml:"MediaContainer"`
	Text                string   `xml:",chardata"`
	Size                string   `xml:"size,attr"`
	AllowSync           string   `xml:"allowSync,attr"`
	Identifier          string   `xml:"identifier,attr"`
	LibrarySectionID    string   `xml:"librarySectionID,attr"`
	LibrarySectionTitle string   `xml:"librarySectionTitle,attr"`
	LibrarySectionUUID  string   `xml:"librarySectionUUID,attr"`
	MediaTagPrefix      string   `xml:"mediaTagPrefix,attr"`
	MediaTagVersion     string   `xml:"mediaTagVersion,attr"`
	Video               struct {
		Text                  string `xml:",chardata"`
		RatingKey             string `xml:"ratingKey,attr"`
		Key                   string `xml:"key,attr"`
		AttrGUID              string `xml:"guid,attr"`
		Studio                string `xml:"studio,attr"`
		Type                  string `xml:"type,attr"`
		Title                 string `xml:"title,attr"`
		LibrarySectionTitle   string `xml:"librarySectionTitle,attr"`
		LibrarySectionID      string `xml:"librarySectionID,attr"`
		LibrarySectionKey     string `xml:"librarySectionKey,attr"`
		ContentRating         string `xml:"contentRating,attr"`
		Summary               string `xml:"summary,attr"`
		AttrRating            string `xml:"rating,attr"`
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
		ChapterSource         string `xml:"chapterSource,attr"`
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
				Stream       []struct {
					Text                 string `xml:",chardata"`
					ID                   string `xml:"id,attr"`
					StreamType           string `xml:"streamType,attr"`
					Default              string `xml:"default,attr"`
					Codec                string `xml:"codec,attr"`
					Index                string `xml:"index,attr"`
					Bitrate              string `xml:"bitrate,attr"`
					Language             string `xml:"language,attr"`
					LanguageTag          string `xml:"languageTag,attr"`
					LanguageCode         string `xml:"languageCode,attr"`
					BitDepth             string `xml:"bitDepth,attr"`
					ChromaLocation       string `xml:"chromaLocation,attr"`
					ChromaSubsampling    string `xml:"chromaSubsampling,attr"`
					CodedHeight          string `xml:"codedHeight,attr"`
					CodedWidth           string `xml:"codedWidth,attr"`
					FrameRate            string `xml:"frameRate,attr"`
					Height               string `xml:"height,attr"`
					Level                string `xml:"level,attr"`
					Profile              string `xml:"profile,attr"`
					RefFrames            string `xml:"refFrames,attr"`
					ScanType             string `xml:"scanType,attr"`
					Width                string `xml:"width,attr"`
					DisplayTitle         string `xml:"displayTitle,attr"`
					ExtendedDisplayTitle string `xml:"extendedDisplayTitle,attr"`
					Selected             string `xml:"selected,attr"`
					Channels             string `xml:"channels,attr"`
					AudioChannelLayout   string `xml:"audioChannelLayout,attr"`
					HeaderCompression    string `xml:"headerCompression,attr"`
					SamplingRate         string `xml:"samplingRate,attr"`
					Title                string `xml:"title,attr"`
				} `xml:"Stream"`
			} `xml:"Part"`
		} `xml:"Media"`
		Genre []struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Filter string `xml:"filter,attr"`
			Tag    string `xml:"tag,attr"`
		} `xml:"Genre"`
		Country []struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Filter string `xml:"filter,attr"`
			Tag    string `xml:"tag,attr"`
		} `xml:"Country"`
		GUID []struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"Guid"`
		Rating []struct {
			Text  string `xml:",chardata"`
			Image string `xml:"image,attr"`
			Value string `xml:"value,attr"`
			Type  string `xml:"type,attr"`
		} `xml:"Rating"`
		Director struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Filter string `xml:"filter,attr"`
			Tag    string `xml:"tag,attr"`
			TagKey string `xml:"tagKey,attr"`
			Thumb  string `xml:"thumb,attr"`
		} `xml:"Director"`
		Writer []struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Filter string `xml:"filter,attr"`
			Tag    string `xml:"tag,attr"`
			TagKey string `xml:"tagKey,attr"`
			Thumb  string `xml:"thumb,attr"`
		} `xml:"Writer"`
		Role []struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Filter string `xml:"filter,attr"`
			Tag    string `xml:"tag,attr"`
			TagKey string `xml:"tagKey,attr"`
			Role   string `xml:"role,attr"`
			Thumb  string `xml:"thumb,attr"`
		} `xml:"Role"`
		Producer struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Filter string `xml:"filter,attr"`
			Tag    string `xml:"tag,attr"`
			TagKey string `xml:"tagKey,attr"`
			Thumb  string `xml:"thumb,attr"`
		} `xml:"Producer"`
	} `xml:"Video"`
}

type TVContainer struct {
	XMLName             xml.Name `xml:"MediaContainer"`
	Text                string   `xml:",chardata"`
	Size                string   `xml:"size,attr"`
	AllowSync           string   `xml:"allowSync,attr"`
	Art                 string   `xml:"art,attr"`
	Content             string   `xml:"content,attr"`
	Identifier          string   `xml:"identifier,attr"`
	LibrarySectionID    string   `xml:"librarySectionID,attr"`
	LibrarySectionTitle string   `xml:"librarySectionTitle,attr"`
	LibrarySectionUUID  string   `xml:"librarySectionUUID,attr"`
	MediaTagPrefix      string   `xml:"mediaTagPrefix,attr"`
	MediaTagVersion     string   `xml:"mediaTagVersion,attr"`
	Nocache             string   `xml:"nocache,attr"`
	Thumb               string   `xml:"thumb,attr"`
	Title1              string   `xml:"title1,attr"`
	Title2              string   `xml:"title2,attr"`
	ViewGroup           string   `xml:"viewGroup,attr"`
	Directory           []struct {
		Text                  string `xml:",chardata"`
		RatingKey             string `xml:"ratingKey,attr"`
		Key                   string `xml:"key,attr"`
		GUID                  string `xml:"guid,attr"`
		Studio                string `xml:"studio,attr"`
		Type                  string `xml:"type,attr"`
		Title                 string `xml:"title,attr"`
		ContentRating         string `xml:"contentRating,attr"`
		Summary               string `xml:"summary,attr"`
		Index                 string `xml:"index,attr"`
		AudienceRating        string `xml:"audienceRating,attr"`
		ViewCount             string `xml:"viewCount,attr"`
		LastViewedAt          string `xml:"lastViewedAt,attr"`
		Year                  string `xml:"year,attr"`
		Thumb                 string `xml:"thumb,attr"`
		Art                   string `xml:"art,attr"`
		Theme                 string `xml:"theme,attr"`
		Duration              string `xml:"duration,attr"`
		OriginallyAvailableAt string `xml:"originallyAvailableAt,attr"`
		LeafCount             string `xml:"leafCount,attr"`
		ViewedLeafCount       string `xml:"viewedLeafCount,attr"`
		ChildCount            string `xml:"childCount,attr"`
		SeasonCount           string `xml:"seasonCount,attr"`
		AddedAt               string `xml:"addedAt,attr"`
		UpdatedAt             string `xml:"updatedAt,attr"`
		AudienceRatingImage   string `xml:"audienceRatingImage,attr"`
		PrimaryExtraKey       string `xml:"primaryExtraKey,attr"`
		SkipCount             string `xml:"skipCount,attr"`
		Tagline               string `xml:"tagline,attr"`
		TitleSort             string `xml:"titleSort,attr"`
		OriginalTitle         string `xml:"originalTitle,attr"`
		Slug                  string `xml:"slug,attr"`
		Genre                 []struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Genre"`
		Country []struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Country"`
		Role []struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Role"`
	} `xml:"Directory"`
}

type SeasonContainer struct {
	XMLName             xml.Name `xml:"MediaContainer"`
	Text                string   `xml:",chardata"`
	Size                string   `xml:"size,attr"`
	AllowSync           string   `xml:"allowSync,attr"`
	Art                 string   `xml:"art,attr"`
	Identifier          string   `xml:"identifier,attr"`
	Key                 string   `xml:"key,attr"`
	LibrarySectionID    string   `xml:"librarySectionID,attr"`
	LibrarySectionTitle string   `xml:"librarySectionTitle,attr"`
	LibrarySectionUUID  string   `xml:"librarySectionUUID,attr"`
	MediaTagPrefix      string   `xml:"mediaTagPrefix,attr"`
	MediaTagVersion     string   `xml:"mediaTagVersion,attr"`
	Nocache             string   `xml:"nocache,attr"`
	ParentIndex         string   `xml:"parentIndex,attr"`
	ParentTitle         string   `xml:"parentTitle,attr"`
	ParentYear          string   `xml:"parentYear,attr"`
	Summary             string   `xml:"summary,attr"`
	Theme               string   `xml:"theme,attr"`
	Thumb               string   `xml:"thumb,attr"`
	Title1              string   `xml:"title1,attr"`
	Title2              string   `xml:"title2,attr"`
	ViewGroup           string   `xml:"viewGroup,attr"`
	Directory           []struct {
		Text            string `xml:",chardata"`
		LeafCount       string `xml:"leafCount,attr"`
		Thumb           string `xml:"thumb,attr"`
		ViewedLeafCount string `xml:"viewedLeafCount,attr"`
		Key             string `xml:"key,attr"`
		Title           string `xml:"title,attr"`
		RatingKey       string `xml:"ratingKey,attr"`
		ParentRatingKey string `xml:"parentRatingKey,attr"`
		GUID            string `xml:"guid,attr"`
		ParentGUID      string `xml:"parentGuid,attr"`
		ParentStudio    string `xml:"parentStudio,attr"`
		Type            string `xml:"type,attr"`
		ParentKey       string `xml:"parentKey,attr"`
		ParentTitle     string `xml:"parentTitle,attr"`
		Summary         string `xml:"summary,attr"`
		Index           string `xml:"index,attr"`
		ParentIndex     string `xml:"parentIndex,attr"`
		ParentYear      string `xml:"parentYear,attr"`
		Art             string `xml:"art,attr"`
		ParentThumb     string `xml:"parentThumb,attr"`
		ParentTheme     string `xml:"parentTheme,attr"`
		AddedAt         string `xml:"addedAt,attr"`
		UpdatedAt       string `xml:"updatedAt,attr"`
		ViewCount       string `xml:"viewCount,attr"`
		LastViewedAt    string `xml:"lastViewedAt,attr"`
	} `xml:"Directory"`
}

type EpisodeContainer struct {
	XMLName                  xml.Name `xml:"MediaContainer"`
	Text                     string   `xml:",chardata"`
	Size                     string   `xml:"size,attr"`
	AllowSync                string   `xml:"allowSync,attr"`
	Art                      string   `xml:"art,attr"`
	GrandparentContentRating string   `xml:"grandparentContentRating,attr"`
	GrandparentRatingKey     string   `xml:"grandparentRatingKey,attr"`
	GrandparentStudio        string   `xml:"grandparentStudio,attr"`
	GrandparentTheme         string   `xml:"grandparentTheme,attr"`
	GrandparentThumb         string   `xml:"grandparentThumb,attr"`
	GrandparentTitle         string   `xml:"grandparentTitle,attr"`
	Identifier               string   `xml:"identifier,attr"`
	Key                      string   `xml:"key,attr"`
	LibrarySectionID         string   `xml:"librarySectionID,attr"`
	LibrarySectionTitle      string   `xml:"librarySectionTitle,attr"`
	LibrarySectionUUID       string   `xml:"librarySectionUUID,attr"`
	MediaTagPrefix           string   `xml:"mediaTagPrefix,attr"`
	MediaTagVersion          string   `xml:"mediaTagVersion,attr"`
	Nocache                  string   `xml:"nocache,attr"`
	ParentIndex              string   `xml:"parentIndex,attr"`
	ParentTitle              string   `xml:"parentTitle,attr"`
	Summary                  string   `xml:"summary,attr"`
	Theme                    string   `xml:"theme,attr"`
	Thumb                    string   `xml:"thumb,attr"`
	Title1                   string   `xml:"title1,attr"`
	Title2                   string   `xml:"title2,attr"`
	ViewGroup                string   `xml:"viewGroup,attr"`
	Video                    []struct {
		Text                  string `xml:",chardata"`
		RatingKey             string `xml:"ratingKey,attr"`
		Key                   string `xml:"key,attr"`
		ParentRatingKey       string `xml:"parentRatingKey,attr"`
		GrandparentRatingKey  string `xml:"grandparentRatingKey,attr"`
		GUID                  string `xml:"guid,attr"`
		ParentGUID            string `xml:"parentGuid,attr"`
		GrandparentGUID       string `xml:"grandparentGuid,attr"`
		Type                  string `xml:"type,attr"`
		Title                 string `xml:"title,attr"`
		GrandparentKey        string `xml:"grandparentKey,attr"`
		ParentKey             string `xml:"parentKey,attr"`
		GrandparentTitle      string `xml:"grandparentTitle,attr"`
		ParentTitle           string `xml:"parentTitle,attr"`
		ContentRating         string `xml:"contentRating,attr"`
		Summary               string `xml:"summary,attr"`
		Index                 string `xml:"index,attr"`
		ParentIndex           string `xml:"parentIndex,attr"`
		AudienceRating        string `xml:"audienceRating,attr"`
		ViewCount             string `xml:"viewCount,attr"`
		LastViewedAt          string `xml:"lastViewedAt,attr"`
		Year                  string `xml:"year,attr"`
		Thumb                 string `xml:"thumb,attr"`
		Art                   string `xml:"art,attr"`
		ParentThumb           string `xml:"parentThumb,attr"`
		GrandparentThumb      string `xml:"grandparentThumb,attr"`
		GrandparentArt        string `xml:"grandparentArt,attr"`
		GrandparentTheme      string `xml:"grandparentTheme,attr"`
		Duration              string `xml:"duration,attr"`
		OriginallyAvailableAt string `xml:"originallyAvailableAt,attr"`
		AddedAt               string `xml:"addedAt,attr"`
		UpdatedAt             string `xml:"updatedAt,attr"`
		AudienceRatingImage   string `xml:"audienceRatingImage,attr"`
		TitleSort             string `xml:"titleSort,attr"`
		Media                 struct {
			Text                  string `xml:",chardata"`
			ID                    string `xml:"id,attr"`
			Duration              string `xml:"duration,attr"`
			Bitrate               string `xml:"bitrate,attr"`
			Width                 string `xml:"width,attr"`
			Height                string `xml:"height,attr"`
			AspectRatio           string `xml:"aspectRatio,attr"`
			AudioChannels         string `xml:"audioChannels,attr"`
			AudioCodec            string `xml:"audioCodec,attr"`
			VideoCodec            string `xml:"videoCodec,attr"`
			VideoResolution       string `xml:"videoResolution,attr"`
			Container             string `xml:"container,attr"`
			VideoFrameRate        string `xml:"videoFrameRate,attr"`
			OptimizedForStreaming string `xml:"optimizedForStreaming,attr"`
			AudioProfile          string `xml:"audioProfile,attr"`
			Has64bitOffsets       string `xml:"has64bitOffsets,attr"`
			VideoProfile          string `xml:"videoProfile,attr"`
			Part                  struct {
				Text                  string `xml:",chardata"`
				ID                    string `xml:"id,attr"`
				Key                   string `xml:"key,attr"`
				Duration              string `xml:"duration,attr"`
				File                  string `xml:"file,attr"`
				Size                  string `xml:"size,attr"`
				AudioProfile          string `xml:"audioProfile,attr"`
				Container             string `xml:"container,attr"`
				Has64bitOffsets       string `xml:"has64bitOffsets,attr"`
				HasThumbnail          string `xml:"hasThumbnail,attr"`
				OptimizedForStreaming string `xml:"optimizedForStreaming,attr"`
				VideoProfile          string `xml:"videoProfile,attr"`
			} `xml:"Part"`
		} `xml:"Media"`
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

type ArtistContainer struct {
	XMLName             xml.Name `xml:"MediaContainer"`
	Text                string   `xml:",chardata"`
	Size                string   `xml:"size,attr"`
	AllowSync           string   `xml:"allowSync,attr"`
	Art                 string   `xml:"art,attr"`
	Content             string   `xml:"content,attr"`
	Identifier          string   `xml:"identifier,attr"`
	LibrarySectionID    string   `xml:"librarySectionID,attr"`
	LibrarySectionTitle string   `xml:"librarySectionTitle,attr"`
	LibrarySectionUUID  string   `xml:"librarySectionUUID,attr"`
	MediaTagPrefix      string   `xml:"mediaTagPrefix,attr"`
	MediaTagVersion     string   `xml:"mediaTagVersion,attr"`
	Nocache             string   `xml:"nocache,attr"`
	Thumb               string   `xml:"thumb,attr"`
	Title1              string   `xml:"title1,attr"`
	Title2              string   `xml:"title2,attr"`
	ViewGroup           string   `xml:"viewGroup,attr"`
	Directory           []struct {
		Text         string `xml:",chardata"`
		RatingKey    string `xml:"ratingKey,attr"`
		Key          string `xml:"key,attr"`
		GUID         string `xml:"guid,attr"`
		Type         string `xml:"type,attr"`
		Title        string `xml:"title,attr"`
		Summary      string `xml:"summary,attr"`
		Index        string `xml:"index,attr"`
		ViewCount    string `xml:"viewCount,attr"`
		LastViewedAt string `xml:"lastViewedAt,attr"`
		Thumb        string `xml:"thumb,attr"`
		Art          string `xml:"art,attr"`
		AddedAt      string `xml:"addedAt,attr"`
		UpdatedAt    string `xml:"updatedAt,attr"`
		TitleSort    string `xml:"titleSort,attr"`
		SkipCount    string `xml:"skipCount,attr"`
		Genre        []struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Genre"`
		Country struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Country"`
	} `xml:"Directory"`
}

type AlbumContainer struct {
	XMLName             xml.Name `xml:"MediaContainer"`
	Text                string   `xml:",chardata"`
	Size                string   `xml:"size,attr"`
	AllowSync           string   `xml:"allowSync,attr"`
	Art                 string   `xml:"art,attr"`
	Content             string   `xml:"content,attr"`
	Identifier          string   `xml:"identifier,attr"`
	LibrarySectionID    string   `xml:"librarySectionID,attr"`
	LibrarySectionTitle string   `xml:"librarySectionTitle,attr"`
	LibrarySectionUUID  string   `xml:"librarySectionUUID,attr"`
	MediaTagPrefix      string   `xml:"mediaTagPrefix,attr"`
	MediaTagVersion     string   `xml:"mediaTagVersion,attr"`
	Nocache             string   `xml:"nocache,attr"`
	Thumb               string   `xml:"thumb,attr"`
	Title1              string   `xml:"title1,attr"`
	Title2              string   `xml:"title2,attr"`
	ViewGroup           string   `xml:"viewGroup,attr"`
	Directory           []struct {
		Text                    string `xml:",chardata"`
		RatingKey               string `xml:"ratingKey,attr"`
		Key                     string `xml:"key,attr"`
		ParentRatingKey         string `xml:"parentRatingKey,attr"`
		GUID                    string `xml:"guid,attr"`
		ParentGUID              string `xml:"parentGuid,attr"`
		Studio                  string `xml:"studio,attr"`
		Type                    string `xml:"type,attr"`
		Title                   string `xml:"title,attr"`
		ParentKey               string `xml:"parentKey,attr"`
		ParentTitle             string `xml:"parentTitle,attr"`
		Summary                 string `xml:"summary,attr"`
		Index                   string `xml:"index,attr"`
		Rating                  string `xml:"rating,attr"`
		ViewCount               string `xml:"viewCount,attr"`
		SkipCount               string `xml:"skipCount,attr"`
		LastViewedAt            string `xml:"lastViewedAt,attr"`
		Year                    string `xml:"year,attr"`
		Thumb                   string `xml:"thumb,attr"`
		Art                     string `xml:"art,attr"`
		ParentThumb             string `xml:"parentThumb,attr"`
		OriginallyAvailableAt   string `xml:"originallyAvailableAt,attr"`
		AddedAt                 string `xml:"addedAt,attr"`
		UpdatedAt               string `xml:"updatedAt,attr"`
		LoudnessAnalysisVersion string `xml:"loudnessAnalysisVersion,attr"`
		Genre                   struct {
			Text string `xml:",chardata"`
			Tag  string `xml:"tag,attr"`
		} `xml:"Genre"`
	} `xml:"Directory"`
}

type Filter struct {
	Name     string
	Value    string
	Modifier string
}

func GetPlexMovies(ipAddress, libraryID, plexToken string, filters []Filter) (movieList []types.PlexMovie) {
	url := fmt.Sprintf("http://%s:32400/library/sections/%s/all", ipAddress, libraryID)
	//nolint:gocritic
	// EG
	// filter = []plex.Filter{
	// 	// does not have german audio
	// 	{
	// 		Name:     "audioLanguage",
	// 		Value:    "de",
	// 		Modifier: "\u0021=",
	// 	},
	// 	// has german audio
	// 	{
	// 		Name:     "audioLanguage",
	// 		Value:    "de",
	// 		Modifier: "=",
	// 	},
	// }

	for i := range filters {
		if i == 0 {
			url += "?"
		} else {
			url += "&"
		}
		url += fmt.Sprintf("%s%s%s", filters[i].Name, filters[i].Modifier, filters[i].Value)
	}

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("GetPlexMovies: Error making request:", err)
		return movieList
	}

	movieList = extractMovies(response)
	// we need to make an API request for each movie to get audio languages
	ch := make(chan types.PlexMovie, len(movieList))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)

	for i := range movieList {
		semaphore <- struct{}{}
		go func(i int) {
			defer func() { <-semaphore }()
			getPlexMovieDetails(ipAddress, plexToken, &movieList[i], ch)
		}(i)
	}
	detailedMovies := make([]types.PlexMovie, len(movieList))
	// wait for all of the go routines to finish
	for i := range movieList {
		detailedMovies[i] = <-ch
	}
	fmt.Printf("Plex movies: %d.\n", len(detailedMovies))
	return detailedMovies
}

func extractMovies(xmlString string) (movieList []types.PlexMovie) {
	var container MovieContainer
	err := xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	for i := range container.Video {
		movieList = append(movieList, types.PlexMovie{
			Title:      container.Video[i].Title,
			Year:       container.Video[i].Year,
			RatingKey:  container.Video[i].RatingKey,
			Resolution: container.Video[i].Media.VideoResolution,
			DateAdded:  parsePlexDate(container.Video[i].AddedAt)})
	}
	return movieList
}

func getPlexMovieDetails(ipAddress, plexToken string, movie *types.PlexMovie, ch chan<- types.PlexMovie) {
	url := fmt.Sprintf("http://%s:32400/library/metadata/%s", ipAddress, movie.RatingKey)

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("getPlexMovieDetails: Error making request:", err)
		ch <- *movie
		return
	}

	var container MovieDetailContainer
	err = xml.Unmarshal([]byte(response), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		ch <- *movie
		return
	}

	languages := make(map[string]string)
	for i := range container.Video.Media.Part.Stream {
		if container.Video.Media.Part.Stream[i].StreamType == "2" {
			languages[container.Video.Media.Part.Stream[i].Language] = container.Video.Media.Part.Stream[i].Language
		}
	}
	// convert map to slice
	for _, value := range languages {
		movie.AudioLanguages = append(movie.AudioLanguages, value)
	}
	ch <- *movie
}

func FilterPlexMovies(movies []types.PlexMovie, filters types.PlexLookupFilters) []types.PlexMovie {
	// filter resolutions first
	var filteredMovies []types.PlexMovie
	for i := range movies {
		if slices.Contains(filters.MatchesResolutions, movies[i].Resolution) {
			filteredMovies = append(filteredMovies, movies[i])
		}
	}
	return filteredMovies
}

// =================================================================================================
func GetPlexTV(ipAddress, libraryID, plexToken string) (tvShowList []types.PlexTVShow) {
	url := fmt.Sprintf("http://%s:32400/library/sections/%s/all", ipAddress, libraryID)

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("GetPlexTV: Error making request:", err)
		return tvShowList
	}

	tvShowList = extractTVShows(response)
	// now we need to get the episodes for each TV show
	for i := range tvShowList {
		tvShowList[i].Seasons = getPlexTVSeasons(ipAddress, plexToken, tvShowList[i].RatingKey)
	}
	// remove TV shows with no seasons
	var filteredTVShows []types.PlexTVShow
	for i := range tvShowList {
		if len(tvShowList[i].Seasons) > 0 {
			filteredTVShows = append(filteredTVShows, tvShowList[i])
		}
	}
	fmt.Printf("Plex TV shows: %d.\n", len(filteredTVShows))
	return filteredTVShows
}

func getPlexTVSeasons(ipAddress, plexToken, ratingKey string) (seasonList []types.PlexTVSeason) {
	url := fmt.Sprintf("http://%s:32400/library/metadata/%s/children?", ipAddress, ratingKey)

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("GetPlexTVSeasons: Error making request:", err)
		return seasonList
	}

	seasonList = extractTVSeasons(response)
	// os.WriteFile("seasons.xml", body, 0644)
	// now we need to get the episodes for each TV show
	ch := make(chan types.PlexTVSeason, len(seasonList))
	semaphore := make(chan struct{}, types.ConcurrencyLimit)
	for i := range seasonList {
		semaphore <- struct{}{}
		go func(i int) {
			defer func() { <-semaphore }()
			getPlexTVEpisodes(ipAddress, plexToken, &seasonList[i], ch)
		}(i)
	}

	detailedSeasons := make([]types.PlexTVSeason, len(seasonList))
	for i := range seasonList {
		detailedSeasons[i] = <-ch
	}
	// remove seasons with no episodes
	var filteredSeasons []types.PlexTVSeason
	for i := range detailedSeasons {
		if len(detailedSeasons[i].Episodes) < 1 {
			continue
		}
		// lets add all of the resolutions for the episodes
		var listOfResolutions []string
		for j := range detailedSeasons[i].Episodes {
			listOfResolutions = append(listOfResolutions, detailedSeasons[i].Episodes[j].Resolution)
		}
		// now we have all of the resolutions for the episodes
		detailedSeasons[i].LowestResolution = findLowestResolution(listOfResolutions)
		// get the last episode added date
		detailedSeasons[i].LastEpisodeAdded = detailedSeasons[i].Episodes[len(detailedSeasons[i].Episodes)-1].DateAdded
		filteredSeasons = append(filteredSeasons, detailedSeasons[i])
	}
	return filteredSeasons
}

func getPlexTVEpisodes(ipAddress, plexToken string, season *types.PlexTVSeason, ch chan<- types.PlexTVSeason) {
	url := fmt.Sprintf("http://%s:32400/library/metadata/%s/children?", ipAddress, season.RatingKey)

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("GetPlexTVEpisodes: Error making request:", err)
		ch <- *season
		return
	}

	showList := extractTVEpisodes(response)
	if len(showList) > 0 {
		season.Episodes = showList
	}
	ch <- *season
}

func extractTVShows(xmlString string) (showList []types.PlexTVShow) {
	var container TVContainer
	err := xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	for i := range container.Directory {
		showList = append(showList, types.PlexTVShow{
			Title: container.Directory[i].Title, Year: container.Directory[i].Year,
			DateAdded: parsePlexDate(container.Directory[i].AddedAt), RatingKey: container.Directory[i].RatingKey})
	}
	return showList
}

func extractTVSeasons(xmlString string) (seasonList []types.PlexTVSeason) {
	var container SeasonContainer
	err := xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	for i := range container.Directory {
		if strings.HasPrefix(container.Directory[i].Title, "Season") {
			seasonNumber, _ := strconv.Atoi(container.Directory[i].Index)
			seasonList = append(seasonList, types.PlexTVSeason{
				Title: container.Directory[i].Title, RatingKey: container.Directory[i].RatingKey, Number: seasonNumber})
		}
	}
	return seasonList
}

func extractTVEpisodes(xmlString string) (episodeList []types.PlexTVEpisode) {
	var container EpisodeContainer
	err := xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	for i := range container.Video {
		intTime, err := strconv.ParseInt(container.Video[i].AddedAt, 10, 64)
		var parsedDate time.Time
		if err != nil {
			parsedDate = time.Time{}
		} else {
			parsedDate = time.Unix(intTime, 0)
		}
		episodeList = append(episodeList, types.PlexTVEpisode{
			Title: container.Video[i].Title, Resolution: container.Video[i].Media.VideoResolution,
			Index: container.Video[i].Index, DateAdded: parsedDate})
	}
	return episodeList
}

// =================================================================================================
func GetPlexMusicArtists(ipAddress, libraryID, plexToken string) (artists []types.PlexMusicArtist) {
	url := fmt.Sprintf("http://%s:32400/library/sections/%s/all", ipAddress, libraryID)

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("GetPlexMusicArtists: Error making request:", err)
		return artists
	}

	artists, err = extractMusicArtists(response)

	if err != nil {
		fmt.Println("Error extracting plex artists:", err)
		return artists
	}
	// now we need to get the albums for each artist
	for i := range artists {
		artists[i].Albums = GetPlexMusicAlbums(ipAddress, plexToken, libraryID, artists[i].RatingKey)
	}

	fmt.Printf("Plex music artists: %d.\n", len(artists))
	return artists
}

func GetPlexMusicAlbums(ipAddress, plexToken, libraryID, ratingKey string) (albums []types.PlexMusicAlbum) {
	url := fmt.Sprintf("http://%s:32400/library/sections/%s/all?artist.id=%s&type=9", ipAddress, libraryID, ratingKey)

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("GetPlexMusicAlbums: Error making request:", err)
		return albums
	}
	albums, _ = extractMusicAlbums(response)

	return albums
}

func extractMusicArtists(xmlString string) (artists []types.PlexMusicArtist, err error) {
	var container ArtistContainer
	err = xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return artists, err
	}

	for i := range container.Directory {
		artists = append(artists, types.PlexMusicArtist{
			Name:      container.Directory[i].Title,
			RatingKey: container.Directory[i].RatingKey,
			DateAdded: parsePlexDate(container.Directory[i].AddedAt)})
	}
	return artists, nil
}

func extractMusicAlbums(xmlString string) (albums []types.PlexMusicAlbum, err error) {
	var container AlbumContainer
	err = xml.Unmarshal([]byte(xmlString), &container)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return albums, err
	}

	for i := range container.Directory {
		albums = append(albums, types.PlexMusicAlbum{
			Title:     container.Directory[i].Title,
			Year:      container.Directory[i].Year,
			DateAdded: parsePlexDate(container.Directory[i].AddedAt),
			RatingKey: container.Directory[i].RatingKey})
	}
	return albums, nil
}

func GetPlexLibraries(ipAddress, plexToken string) (libraryList []types.PlexLibrary, err error) {
	url := fmt.Sprintf("http://%s:32400/library/sections", ipAddress)

	response, err := makePlexAPIRequest(url, plexToken)
	if err != nil {
		fmt.Println("GetPlexLibraries: Error making request:", err)
		return libraryList, err
	}

	libraryList, err = extractLibraries(response)
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

// =================================================================================================

func makePlexAPIRequest(inputURL, plexToken string) (response string, err error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", inputURL, http.NoBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("X-Plex-Token", plexToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}
	return string(body), nil
}

func findLowestResolution(resolutions []string) (lowestResolution string) {
	if slices.Contains(resolutions, types.PlexResolutionSD) {
		return types.PlexResolutionSD
	}
	if slices.Contains(resolutions, types.PlexResolution240) {
		return types.PlexResolution240
	}
	if slices.Contains(resolutions, types.PlexResolution480) {
		return types.PlexResolution480
	}
	if slices.Contains(resolutions, types.PlexResolution576) {
		return types.PlexResolution576
	}
	if slices.Contains(resolutions, types.PlexResolution720) {
		return types.PlexResolution720
	}
	if slices.Contains(resolutions, types.PlexResolution1080) {
		return types.PlexResolution1080
	}
	if slices.Contains(resolutions, types.PlexResolution4K) {
		return types.PlexResolution4K
	}
	return ""
}

func parsePlexDate(plexDate string) (parsedDate time.Time) {
	intTime, err := strconv.ParseInt(plexDate, 10, 64)
	if err != nil {
		parsedDate = time.Time{}
	} else {
		parsedDate = time.Unix(intTime, 0)
	}
	return parsedDate
}
