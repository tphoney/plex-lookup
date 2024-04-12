package types

const (
	DiskBluray = "Blu-ray"
	DiskDVD    = "DVD"
	Disk4K     = "4K Blu-ray"
)

type MovieSearchResults struct {
	PlexMovie
	SearchURL     string
	Matches4k     int
	MatchesBluray int
	SearchResults []SearchResult
}

type SearchResult struct {
	FoundTitle string
	UITitle    string
	BestMatch  bool
	URL        string
	Format     string
	Year       string
}

type PlexMovie struct {
	Title     string
	Year      string
	DateAdded string
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
}
