package types

const (
	DiskBluray = "Blu-ray"
	DiskDVD    = "DVD"
	Disk4K     = "4K Blu-ray"
)

type Movie struct {
	Title string
	Year  string
}

type MovieSearchResults struct {
	Movie
	SearchURL     string
	SearchResults []SearchResult
}

type SearchResult struct {
	FoundTitle string
	BestMatch  bool
	URL        string
	Format     string
	Year       string
}

type PlexLibrary struct {
	Title string
	Type  string
	ID    string
}
