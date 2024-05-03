package types

import "time"

const (
	DiskBluray       = "Blu-ray"
	DiskDVD          = "DVD"
	Disk4K           = "4K Blu-ray"
	PlexMovieType    = "Movie"
	ConcurrencyLimit = 10
)

type SearchResults struct {
	PlexMovie
	PlexTVShow
	PlexMusicArtist
	SearchURL          string
	Matches4k          int
	MatchesBluray      int
	MovieSearchResults []MovieSearchResult
	TVSearchResults    []TVSearchResult
	MusicSearchResults []MusicSearchResult
}

type PlexInformation struct {
	IP             string
	Token          string
	MovieLibraryID string
	TVLibraryID    string
	MusicLibraryID string
}

// ==============================================================================================================
type PlexMovie struct {
	Title     string
	Year      string
	DateAdded time.Time
}

type MovieSearchResult struct {
	FoundTitle  string
	UITitle     string
	BestMatch   bool
	URL         string
	Format      string
	Year        string
	ReleaseDate time.Time
	NewRelease  bool
}

// ==============================================================================================================
type PlexTVShow struct {
	Title     string
	Year      string
	RatingKey string
	DateAdded time.Time
	Seasons   []PlexTVSeason
}

type PlexTVSeason struct {
	Title     string
	Number    int
	RatingKey string
	Episodes  []PlexTVEpisode
}

type PlexTVEpisode struct {
	Title      string
	Index      string
	Resolution string
}

type TVSearchResult struct {
	FoundTitle  string
	UITitle     string
	BestMatch   bool
	URL         string
	Format      []string
	Year        string
	ReleaseDate time.Time
	NewRelease  bool
	BoxSet      bool
	Series      []TVSeriesResult
}

type TVSeriesResult struct {
	Number      int
	URL         string
	Format      []string
	ReleaseDate time.Time
}

// ==============================================================================================================
type PlexMusicArtist struct {
	Name      string
	RatingKey string
	DateAdded time.Time
	Albums    []PlexMusicAlbum
}

type PlexMusicAlbum struct {
	Title     string
	RatingKey string
	Year      string
	DateAdded time.Time
}

type MusicSearchResult struct {
	Name   string
	ID     string
	URL    string
	Albums []MusicSearchAlbumResult
}

type MusicSearchAlbumResult struct {
	Title string
	ID    string
	URL   string
	Year  string
}

// ==============================================================================================================
type PlexLibrary struct {
	Title string
	Type  string
	ID    string
}
