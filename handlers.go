package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	c "miniflux.app/v2/client"
)

// healthzHandler returns the health status of the application.
func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check connection to Miniflux API
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Simple connectivity test
	_, err := s.client.Me()
	if err != nil {
		log.Printf("Health check failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"message": "Cannot connect to Miniflux API",
			"error":   err.Error(),
		})
		return
	}

	select {
	case <-ctx.Done():
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"message": "Health check timeout",
		})
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// processEntriesHandler processes unread entries.
func (s *Server) processEntriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Processing unread entries...")
	timer := prometheus.NewTimer(ProcessDurationSeconds)
	defer timer.ObserveDuration()

	unreadFilter := &c.Filter{
		Status: c.EntryStatusUnread,
	}
	processed, errors, entries := s.Process(unreadFilter)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"processed": processed,
		"errors":    errors,
		"total":     entries,
	})
}

// Process retrieves unread entries and processes them.
// It is safe to call concurrently: if a run is already in progress, the call
// returns immediately with zero counts.
func (s *Server) Process(unreadFilter *c.Filter) (int, int, int) {
	if !s.processing.CompareAndSwap(false, true) {
		log.Println("Process already running, skipping")
		return 0, 0, 0
	}
	defer s.processing.Store(false)

	start := time.Now()
	entries, err := s.client.Entries(unreadFilter)
	MinifluxAPIDurationSeconds.WithLabelValues("entries").Observe(time.Since(start).Seconds())
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		EntriesProcessingErrorsTotal.WithLabelValues("fetch_entries").Inc()
		return 0, 0, 0
	}

	processed := 0
	errors := 0

	for _, entry := range entries.Entries {
		log.Printf("Fetching entry %d: %s", entry.ID, entry.Title)
		startFetch := time.Now()
		_, err := s.client.FetchEntryOriginalContent(entry.ID)
		MinifluxAPIDurationSeconds.WithLabelValues("fetch_entry").Observe(time.Since(startFetch).Seconds())
		if err != nil {
			log.Printf("Error fetching entry %d: %v", entry.ID, err)
			EntriesProcessingErrorsTotal.WithLabelValues("fetch_entry").Inc()
			errors++
			continue
		}

		log.Printf("Saving entry %d: %s", entry.ID, entry.Title)
		startSave := time.Now()
		err = s.client.SaveEntry(entry.ID)
		MinifluxAPIDurationSeconds.WithLabelValues("save_entry").Observe(time.Since(startSave).Seconds())
		if err != nil {
			log.Printf("Error saving entry %d: %v", entry.ID, err)
			EntriesProcessingErrorsTotal.WithLabelValues("save_entry").Inc()
			errors++
			continue
		}

		startUpdate := time.Now()
		err = s.client.UpdateEntries([]int64{entry.ID}, c.EntryStatusRead)
		MinifluxAPIDurationSeconds.WithLabelValues("update_entries").Observe(time.Since(startUpdate).Seconds())
		if err != nil {
			log.Printf("Error marking entry %d as read: %v", entry.ID, err)
			EntriesProcessingErrorsTotal.WithLabelValues("mark_read").Inc()
			errors++
			continue
		}

		processed++
	}

	log.Printf("Processing complete: %d processed, %d errors", processed, errors)
	EntriesProcessedTotal.Add(float64(processed))

	return processed, errors, len(entries.Entries)
}
