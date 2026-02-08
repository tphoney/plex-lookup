package web

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestJobTracker_CreateJob(t *testing.T) {
	jt := NewJobTracker()

	jobID, ctx := jt.CreateJob("movies", 100)

	if jobID == "" {
		t.Error("Expected non-empty job ID")
	}

	if ctx == nil {
		t.Error("Expected non-nil context")
	}

	job, exists := jt.GetProgress(jobID)
	if !exists {
		t.Fatal("Expected job to exist")
	}

	if job.Type != "movies" {
		t.Errorf("Expected job type 'movies', got %s", job.Type)
	}

	if job.Total != 100 {
		t.Errorf("Expected total 100, got %d", job.Total)
	}

	if job.Status != jobStatusRunning {
		t.Errorf("Expected status %s, got %s", jobStatusRunning, job.Status)
	}

	if job.Current != 0 {
		t.Errorf("Expected current 0, got %d", job.Current)
	}
}

func TestJobTracker_UpdateProgress(t *testing.T) {
	jt := NewJobTracker()
	jobID, _ := jt.CreateJob("tv", 50)

	jt.UpdateProgress(jobID, 25, "Processing episode 25")

	job, exists := jt.GetProgress(jobID)
	if !exists {
		t.Fatal("Expected job to exist")
	}

	if job.Current != 25 {
		t.Errorf("Expected current 25, got %d", job.Current)
	}

	if job.Phase != "Processing episode 25" {
		t.Errorf("Expected phase 'Processing episode 25', got %s", job.Phase)
	}
}

func TestJobTracker_UpdateProgress_NonExistentJob(t *testing.T) {
	jt := NewJobTracker()

	// Should not panic when updating non-existent job
	jt.UpdateProgress("nonexistent", 10, "test")

	_, exists := jt.GetProgress("nonexistent")
	if exists {
		t.Error("Expected non-existent job to remain non-existent")
	}
}

func TestJobTracker_MarkComplete(t *testing.T) {
	jt := NewJobTracker()
	jobID, _ := jt.CreateJob("music", 10)

	results := "<div>Results here</div>"
	jt.MarkComplete(jobID, results)

	job, exists := jt.GetProgress(jobID)
	if !exists {
		t.Fatal("Expected job to exist")
	}

	if job.Status != jobStatusComplete {
		t.Errorf("Expected status %s, got %s", jobStatusComplete, job.Status)
	}

	if job.Current != job.Total {
		t.Errorf("Expected current %d to equal total %d", job.Current, job.Total)
	}

	if job.Results.(string) != results {
		t.Errorf("Expected results %s, got %s", results, job.Results)
	}

	if job.Phase != "" {
		t.Errorf("Expected empty phase, got %s", job.Phase)
	}
}

func TestJobTracker_CancelJob(t *testing.T) {
	jt := NewJobTracker()
	jobID, ctx := jt.CreateJob("movies", 100)

	// Verify context is initially active
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be cancelled initially")
	default:
		// Expected
	}

	success := jt.CancelJob(jobID)
	if !success {
		t.Error("Expected CancelJob to return true")
	}

	job, exists := jt.GetProgress(jobID)
	if !exists {
		t.Fatal("Expected job to exist")
	}

	if job.Status != jobStatusCancelled {
		t.Errorf("Expected status %s, got %s", jobStatusCancelled, job.Status)
	}

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled")
	}
}

func TestJobTracker_CancelJob_AlreadyComplete(t *testing.T) {
	jt := NewJobTracker()
	jobID, _ := jt.CreateJob("tv", 10)

	jt.MarkComplete(jobID, "results")

	success := jt.CancelJob(jobID)
	if success {
		t.Error("Expected CancelJob to return false for completed job")
	}

	job, _ := jt.GetProgress(jobID)
	if job.Status != jobStatusComplete {
		t.Errorf("Expected status to remain %s, got %s", jobStatusComplete, job.Status)
	}
}

func TestJobTracker_CancelJob_NonExistent(t *testing.T) {
	jt := NewJobTracker()

	success := jt.CancelJob("nonexistent")
	if success {
		t.Error("Expected CancelJob to return false for non-existent job")
	}
}

func TestJobTracker_CleanupOldJobs(t *testing.T) {
	jt := NewJobTracker()

	// Create an old job by manually setting CreatedAt
	jobID, ctx := jt.CreateJob("movies", 10)
	jt.mu.Lock()
	jt.jobs[jobID].CreatedAt = time.Now().Add(-15 * time.Minute)
	jt.mu.Unlock()

	// Create a recent job
	recentID, recentCtx := jt.CreateJob("tv", 20)

	// Run cleanup
	jt.CleanupOldJobs()

	// Old job should be removed
	_, exists := jt.GetProgress(jobID)
	if exists {
		t.Error("Expected old job to be cleaned up")
	}

	// Verify old job context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Old job context should be cancelled")
	}

	// Recent job should still exist
	_, exists = jt.GetProgress(recentID)
	if !exists {
		t.Error("Expected recent job to still exist")
	}

	// Verify recent job context is still active
	select {
	case <-recentCtx.Done():
		t.Error("Recent job context should not be cancelled")
	default:
		// Expected
	}
}

func TestJobTracker_GetProgress_ReturnsCopy(t *testing.T) {
	jt := NewJobTracker()
	jobID, _ := jt.CreateJob("music", 100)

	job1, _ := jt.GetProgress(jobID)
	job2, _ := jt.GetProgress(jobID)

	// Modifying job1 should not affect job2
	job1.Current = 999

	if job2.Current == 999 {
		t.Error("GetProgress should return a copy, not a reference")
	}
}

func TestJobTracker_ConcurrentAccess(t *testing.T) {
	jt := NewJobTracker()
	const numGoroutines = 50
	var wg sync.WaitGroup

	// Test concurrent job creation
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			defer wg.Done()
			jobID, _ := jt.CreateJob("test", n)
			jt.UpdateProgress(jobID, n/2, "halfway")
			if n%2 == 0 {
				jt.MarkComplete(jobID, "done")
			} else {
				jt.CancelJob(jobID)
			}
		}(i)
	}

	wg.Wait()

	// Verify no jobs are in running state after all operations
	jt.mu.RLock()
	for _, job := range jt.jobs {
		if job.Status == jobStatusRunning {
			t.Error("Found job still in running state after concurrent operations")
		}
	}
	jt.mu.RUnlock()
}

func TestJobTracker_UniqueJobIDs(t *testing.T) {
	jt := NewJobTracker()
	ids := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		jobID, _ := jt.CreateJob("test", 10)
		if ids[jobID] {
			t.Errorf("Duplicate job ID generated: %s", jobID)
		}
		ids[jobID] = true
	}
}

func TestJobTracker_UpdateProgress_OnlyUpdatesRunningJobs(t *testing.T) {
	jt := NewJobTracker()
	jobID, _ := jt.CreateJob("movies", 100)

	// Mark as complete
	jt.MarkComplete(jobID, "results")

	// Try to update progress
	jt.UpdateProgress(jobID, 50, "should not update")

	job, _ := jt.GetProgress(jobID)
	if job.Current != 100 {
		t.Errorf("Expected current to remain 100, got %d", job.Current)
	}
	if job.Phase != "" {
		t.Errorf("Expected empty phase, got %s", job.Phase)
	}
}

func TestGetJobTracker(t *testing.T) {
	// Initialise global tracker
	jobTracker = NewJobTracker()

	tracker := GetJobTracker()
	if tracker == nil {
		t.Error("Expected non-nil job tracker")
	}

	if tracker != jobTracker {
		t.Error("Expected GetJobTracker to return global jobTracker instance")
	}
}

func TestJobTracker_ContextCancellation(t *testing.T) {
	jt := NewJobTracker()
	jobID, ctx := jt.CreateJob("movies", 100)

	// Start a goroutine that listens to context
	done := make(chan bool)
	go func() {
		<-ctx.Done()
		done <- true
	}()

	// Cancel the job
	jt.CancelJob(jobID)

	// Wait for context cancellation
	select {
	case <-done:
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Context should be cancelled when job is cancelled")
	}
}

func TestJobTracker_EmptyPhase(t *testing.T) {
	jt := NewJobTracker()
	jobID, _ := jt.CreateJob("tv", 50)

	// Set initial phase
	jt.UpdateProgress(jobID, 10, "initial phase")

	job, _ := jt.GetProgress(jobID)
	if job.Phase != "initial phase" {
		t.Errorf("Expected phase 'initial phase', got %s", job.Phase)
	}

	// Update progress with empty phase - should keep old phase
	jt.UpdateProgress(jobID, 20, "")

	job, _ = jt.GetProgress(jobID)
	if job.Current != 20 {
		t.Errorf("Expected current 20, got %d", job.Current)
	}
	if job.Phase != "initial phase" {
		t.Errorf("Expected phase to remain 'initial phase', got %s", job.Phase)
	}
}

func TestCleanupGoroutineStops(t *testing.T) {
	// Create a temporary context for this test
	ctx, cancel := context.WithCancel(context.Background())
	jt := NewJobTracker()

	var wg sync.WaitGroup
	wg.Add(1)

	// Start cleanup goroutine
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				jt.CleanupOldJobs()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Let it run a bit
	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for goroutine to stop
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Expected - goroutine stopped
	case <-time.After(1 * time.Second):
		t.Error("Cleanup goroutine should stop when context is cancelled")
	}
}
