package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	c "miniflux.app/v2/client"
)

// healthzHandler returns the health status of the application
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
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"message": "Cannot connect to Miniflux API",
			"error":   err.Error(),
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	select {
	case <-ctx.Done():
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"message": "Health check timeout",
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// processEntriesHandler processes unread entries
func (s *Server) processEntriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Processing unread entries...")

	unreadFilter := &c.Filter{
		Status: c.EntryStatusUnread,
	}
	processed, errors, entries := s.Process(unreadFilter)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"processed": processed,
		"errors":    errors,
		"total":     entries,
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) Process(unreadFilter *c.Filter) (int, int, int) {

	entries, err := s.client.Entries(unreadFilter)
	if err != nil {
		log.Printf("Error fetching entries: %v", err)
		// http.Error(w, fmt.Sprintf("Error fetching entries: %v", err), http.StatusInternalServerError)
		return 0, 0, 0
	}

	processed := 0
	errors := 0

	for _, entry := range entries.Entries {
		log.Printf("Saving entry %d: %s", entry.ID, entry.Title)
		if err := s.client.SaveEntry(entry.ID); err != nil {
			log.Printf("Error saving entry %d: %v", entry.ID, err)
			errors++
			continue
		}

		if err := s.client.UpdateEntries([]int64{entry.ID}, c.EntryStatusRead); err != nil {
			log.Printf("Error marking entry %d as read: %v", entry.ID, err)
			errors++
			continue
		}

		processed++
	}

	log.Printf("Processing complete: %d processed, %d errors", processed, errors)

	return processed, errors, len(entries.Entries)
}
