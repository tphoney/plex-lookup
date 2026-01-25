package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tphoney/plex-lookup/types"
	"github.com/tphoney/plex-lookup/web/movies"
	"github.com/tphoney/plex-lookup/web/music"
	"github.com/tphoney/plex-lookup/web/settings"
	"github.com/tphoney/plex-lookup/web/tv"
)

var (
	//go:embed index.html
	indexPage string

	//go:embed static/*
	staticFS embed.FS

	port       string = "9090"
	config     *types.Configuration
	jobTracker *JobTracker
)

const (
	jobStatusRunning  = "running"
	jobStatusComplete = "complete"
	jobStatusCanceled = "canceled"
	cleanupInterval   = 5 * time.Minute
)

// JobProgress represents the state of a running or completed job.
type JobProgress struct {
	ID         string
	Type       string // "music", "movies", "tv"
	Status     string // "running", "complete", "canceled"
	Current    int
	Total      int
	Phase      string // e.g., "Searching artists", "Fetching albums"
	Results    any
	CreatedAt  time.Time
	CancelFunc context.CancelFunc
}

// JobTracker manages multiple concurrent jobs with progress tracking.
type JobTracker struct {
	mu         sync.RWMutex
	jobs       map[string]*JobProgress
	jobCounter atomic.Uint64
}

// NewJobTracker creates a new JobTracker instance.
func NewJobTracker() *JobTracker {
	return &JobTracker{
		jobs: make(map[string]*JobProgress),
	}
}

// CreateJob creates a new job and returns its ID and cancellable context.
func (jt *JobTracker) CreateJob(jobType string, total int) (string, context.Context) {
	id := fmt.Sprintf("%d", jt.jobCounter.Add(1))
	ctx, cancel := context.WithCancel(context.Background())

	job := &JobProgress{
		ID:         id,
		Type:       jobType,
		Status:     jobStatusRunning,
		Current:    0,
		Total:      total,
		Phase:      "",
		Results:    nil,
		CreatedAt:  time.Now(),
		CancelFunc: cancel,
	}

	jt.mu.Lock()
	jt.jobs[id] = job
	jt.mu.Unlock()

	return id, ctx
}

// UpdateProgress updates the current progress and optional phase description.
func (jt *JobTracker) UpdateProgress(jobID string, current int, phase string) {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	if job, exists := jt.jobs[jobID]; exists && job.Status == jobStatusRunning {
		job.Current = current
		if phase != "" {
			job.Phase = phase
		}
	}
}

// GetProgress retrieves the current progress for a job.
func (jt *JobTracker) GetProgress(jobID string) (*JobProgress, bool) {
	jt.mu.RLock()
	defer jt.mu.RUnlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	jobCopy := *job
	return &jobCopy, true
}

// MarkComplete marks a job as complete and stores its results.
func (jt *JobTracker) MarkComplete(jobID string, results any) {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	if job, exists := jt.jobs[jobID]; exists {
		job.Status = jobStatusComplete
		job.Current = job.Total
		job.Results = results
		job.Phase = ""
	}
}

// CancelJob cancels a running job.
func (jt *JobTracker) CancelJob(jobID string) bool {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	if job, exists := jt.jobs[jobID]; exists && job.Status == jobStatusRunning {
		job.Status = jobStatusCanceled
		if job.CancelFunc != nil {
			job.CancelFunc()
		}
		return true
	}
	return false
}

// CleanupOldJobs removes jobs older than 10 minutes.
func (jt *JobTracker) CleanupOldJobs() {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for id, job := range jt.jobs {
		if job.CreatedAt.Before(cutoff) {
			if job.CancelFunc != nil {
				job.CancelFunc()
			}
			delete(jt.jobs, id)
		}
	}
}

func StartServer(startingConfig *types.Configuration) {
	config = startingConfig
	jobTracker = NewJobTracker()

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			jobTracker.CleanupOldJobs()
			fmt.Println("Cleaned up old jobs")
		}
	}()

	// find the local IP address
	ipAddress := GetOutboundIP()
	fmt.Printf("Starting server on http://%s:%s\n", ipAddress.String(), port)
	mux := http.NewServeMux()

	// serve static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	mux.HandleFunc("/settings", settings.SettingsHandler)
	mux.HandleFunc("/settings/plexlibraries", settings.ProcessPlexLibrariesHTML)
	mux.HandleFunc("/settings/plexinfook", settings.SettingsConfig{Config: config}.PlexInformationOKHTML)

	mux.HandleFunc("/movies", movies.MoviesHandler)
	mux.HandleFunc("/moviesprocess", movies.MoviesConfig{Config: config, JobTracker: jobTracker}.ProcessHTML)
	mux.HandleFunc("/moviesplaylists", movies.MoviesConfig{Config: config}.PlaylistHTML)

	mux.HandleFunc("/tv", tv.TVHandler)
	mux.HandleFunc("/tvprocess", tv.TVConfig{Config: config, JobTracker: jobTracker}.ProcessHTML)
	mux.HandleFunc("/tvplaylists", tv.TVConfig{Config: config}.PlaylistHTML)

	mux.HandleFunc("/music", music.MusicHandler)
	mux.HandleFunc("/musicprocess", music.MusicConfig{Config: config, JobTracker: jobTracker}.ProcessHTML)
	mux.HandleFunc("/musicplaylists", music.MusicConfig{Config: config}.PlaylistHTML)

	// Job management endpoints
	mux.HandleFunc("/progress/", progressHandler)
	mux.HandleFunc("/cancel/", cancelHandler)

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/settings/save", settingsSaveHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux) //nolint: gosec
	if err != nil {
		fmt.Printf("Failed to start server on port %s: %s\n", port, err)
		panic(err)
	}
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("index").Parse(indexPage))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render index", http.StatusInternalServerError)
		return
	}
}

func settingsSaveHandler(w http.ResponseWriter, r *http.Request) {
	oldConfig := config
	// Retrieve form fields (replace with proper values)
	config.PlexIP = r.FormValue("plexIP")
	config.PlexToken = r.FormValue("plexToken")
	config.PlexMovieLibraryID = r.FormValue("plexMovieLibraryID")
	config.PlexTVLibraryID = r.FormValue("plexTVLibraryID")
	config.PlexMusicLibraryID = r.FormValue("plexMusicLibraryID")
	config.AmazonRegion = r.FormValue("amazonRegion")
	config.MusicBrainzURL = r.FormValue("musicBrainzURL")
	config.SpotifyClientID = r.FormValue("spotifyClientID")
	config.SpotifyClientSecret = r.FormValue("spotifyClientSecret")
	fmt.Fprint(w, `<h2>Saved!</h2><a href="/">Back</a>`)
	fmt.Printf("Saved Settings\nold\n%+v\nnew\n%+v\n", oldConfig, config)
}

func GetOutboundIP() net.IP {
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Failed to get local IP address")
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// GetJobTracker returns the global job tracker instance.
func GetJobTracker() *JobTracker {
	return jobTracker
}

func progressHandler(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL path /progress/{jobID}
	jobID := r.URL.Path[len("/progress/"):]
	if jobID == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	job, exists := jobTracker.GetProgress(jobID)
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// If job is still running, return progress bar
	if job.Status == jobStatusRunning {
		phaseText := ""
		if job.Phase != "" {
			phaseText = fmt.Sprintf("<p>%s</p>", job.Phase)
		}
		fmt.Fprintf(w,
			`<div hx-get="/progress/%s" hx-trigger="every 100ms" class="container" id="progress" hx-swap="outerHTML">
			%s<progress value="%d" max="%d"></progress>
			<button hx-post="/cancel/%s" hx-swap="outerHTML" hx-target="#progress">Cancel</button>
			</div>`,
			jobID, phaseText, job.Current, job.Total, jobID)
		return
	}

	// If job is canceled
	if job.Status == "canceled" {
		fmt.Fprint(w, `<div class="container" id="progress"><p>Job canceled</p></div>`)
		return
	}

	// Job is complete - render results based on job type
	switch job.Type {
	case "music":
		if results, ok := job.Results.(string); ok {
			fmt.Fprintf(w, `<div id="progress">%s</div>`, results)
		}
	case "movies":
		if results, ok := job.Results.(string); ok {
			fmt.Fprintf(w, `<div id="progress">%s</div>`, results)
		}
	case "tv":
		if results, ok := job.Results.(string); ok {
			fmt.Fprintf(w, `<div id="progress">%s</div>`, results)
		}
	}
}

func cancelHandler(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL path /cancel/{jobID}
	jobID := r.URL.Path[len("/cancel/"):]
	if jobID == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	if jobTracker.CancelJob(jobID) {
		fmt.Fprint(w, `<div class="container" id="progress"><p>Job canceled successfully</p></div>`)
	} else {
		fmt.Fprint(w, `<div class="container" id="progress"><p>Job not found or already complete</p></div>`)
	}
}
