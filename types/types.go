package types

import "time"

const (
	DiskBluray         = "Blu-ray"
	DiskDVD            = "DVD"
	Disk4K             = "4K Blu-ray"
	PlexMovieType      = "Movie"
	PlexResolutionSD   = "sd"
	PlexResolution240  = "240"
	PlexResolution480  = "480"
	PlexResolution576  = "576"
	PlexResolution720  = "720"
	PlexResolution1080 = "1080"
	PlexResolution4K   = "4k"
	ConcurrencyLimit   = 10
	StringTrue         = "true"
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
	MusicSearchResults []MusicArtistSearchResult
}

type Configuration struct {
	PlexIP              string
	PlexToken           string
	PlexMovieLibraryID  string
	PlexTVLibraryID     string
	PlexMusicLibraryID  string
	SpotifyClientID     string
	SpotifyClientSecret string
}

type FilteringOptions struct {
	AudioLanguage string
	NewerVersion  bool
}

// ==============================================================================================================
type PlexMovie struct {
	Title      string
	Year       string
	RatingKey  string
	Resolution string
	DateAdded  time.Time
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
	Title            string
	Number           int
	RatingKey        string
	LowestResolution string
	LastEpisodeAdded time.Time
	Episodes         []PlexTVEpisode
}

type PlexTVEpisode struct {
	Title      string
	Index      string
	Resolution string
	DateAdded  time.Time
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
	Seasons     []TVSeasonResult
}

type TVSeasonResult struct {
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

type MusicArtistSearchResult struct {
	Name        string
	ID          string
	URL         string
	OwnedAlbums int
	Albums      []MusicAlbumSearchResult
}

type MusicAlbumSearchResult struct {
	Title string
	ID    string
	URL   string
	Year  string
}

type MusicSimilarArtistResult struct {
	Name            string
	URL             string
	Owned           bool
	SimilarityCount int
}

// ==============================================================================================================
type PlexLibrary struct {
	Title string
	Type  string
	ID    string
}
