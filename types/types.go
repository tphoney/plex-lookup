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

// TVSearchResponse is the new dedicated struct for TV search results.
type TVSearchResponse struct {
	PlexTVShow
	SearchURL       string
	TVSearchResults []TVSearchResult
	Matches4k       int
	MatchesBluray   int
	MatchesDVD      int
}

// MusicSearchResponse is the new dedicated struct for music search results.
type MusicSearchResponse struct {
	PlexMusicArtist
	SearchURL          string
	MusicSearchResults []MusicArtistSearchResult
}

// MovieSearchResponse is the new dedicated struct for movie search results.
type MovieSearchResponse struct {
	PlexMovie
	SearchURL          string
	Matches4k          int
	MatchesBluray      int
	MatchesDVD         int
	MovieSearchResults []MovieSearchResult
}

type Configuration struct {
	PlexIP              string
	PlexToken           string
	PlexMovieLibraryID  string
	PlexTVLibraryID     string
	PlexMusicLibraryID  string
	AmazonRegion        string
	MusicBrainzURL      string
	SpotifyClientID     string
	SpotifyClientSecret string
}

type MovieLookupFilters struct {
	AudioLanguage string
	NewerVersion  bool
}

type PlexLookupFilters struct {
	MissingAudioLanguage string
	MatchesResolutions   []string
}

// ==============================================================================================================
type PlexMovie struct {
	Title          string
	Year           string
	RatingKey      string
	Resolution     string
	AudioLanguages []string
	DateAdded      time.Time
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
	Title             string
	Year              string
	RatingKey         string
	DateAdded         time.Time
	FirstEpisodeAired time.Time
	LastEpisodeAired  time.Time
	Seasons           []PlexTVSeason
}

type PlexTVSeason struct {
	Number            int
	RatingKey         string
	LowestResolution  string
	LastEpisodeAdded  time.Time
	FirstEpisodeAired time.Time
	LastEpisodeAired  time.Time
	Episodes          []PlexTVEpisode
}

type PlexTVEpisode struct {
	Title           string
	Index           string
	Resolution      string
	DateAdded       time.Time
	OriginallyAired time.Time
}

type TVSearchResult struct {
	FoundTitle     string
	UITitle        string
	BestMatch      bool
	URL            string
	Format         []string
	FirstAiredYear string
	ReleaseDate    time.Time
	NewRelease     bool
	Seasons        []TVSeasonResult
}

type TVSeasonResult struct {
	Number      int
	BoxSetName  string
	URL         string
	Format      string
	BoxSet      bool
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
	Name           string
	ID             string
	URL            string
	FirstAlbumYear int
	LastAlbumYear  int

	OwnedAlbums []string
	FoundAlbums []MusicAlbumSearchResult
}

type MusicAlbumSearchResult struct {
	Title          string
	SanitizedTitle string
	ID             string
	URL            string
	Year           string
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

type PlexPlaylist struct {
	Title     string
	Type      string
	RatingKey string
}
