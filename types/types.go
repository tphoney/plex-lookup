package types

import "time"

const (
	DiskBluray       = "Blu-ray"
	DiskDVD          = "DVD"
	Disk4K           = "4K Blu-ray"
	PlexMovieType    = "Movie"
	ConcurrencyLimit = 10
)

type MovieSearchResults struct {
	PlexMovie
	SearchURL     string
	Matches4k     int
	MatchesBluray int
	SearchResults []SearchResult
}

type SearchResult struct {
	FoundTitle  string
	UITitle     string
	BestMatch   bool
	URL         string
	Format      string
	Year        string
	ReleaseDate time.Time
	NewRelease  bool
}

type PlexMovie struct {
	Title     string
	Year      string
	DateAdded time.Time
}

type PlexTVShow struct {
	Title     string
	Year      string
	RatingKey string
	DateAdded time.Time
	Seasons   []PlexTVSeason
}

type PlexTVSeason struct {
	Title     string
	RatingKey string
	Episodes  []PlexTVEpisode
}

type PlexTVEpisode struct {
	Title      string
	Index      string
	Resolution string
}

type PlexLibrary struct {
	Title string
	Type  string
	ID    string
}

type PlexInformation struct {
	IP             string
	Token          string
	MovieLibraryID string
	TVLibraryID    string
}
