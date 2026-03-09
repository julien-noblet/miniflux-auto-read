package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EntriesProcessedTotal increments when an entry is successfully processed
	EntriesProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "miniflux_entries_processed_total",
		Help: "The total number of processed Miniflux entries",
	})

	// ProcessDurationSeconds tracks the duration of the processing logic
	ProcessDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "miniflux_entries_processing_duration_seconds",
		Help:    "Duration of entries processing in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// HTTPRequestDurationSeconds tracks the duration of HTTP requests
	HTTPRequestDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"method", "path", "status"})

	// MinifluxAPIDurationSeconds tracks the duration of calls to the Miniflux API
	MinifluxAPIDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "miniflux_api_duration_seconds",
		Help:    "Duration of Miniflux API calls in seconds",
		Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"operation"})

	// EntriesProcessingErrorsTotal increments when an error occurs during processing
	EntriesProcessingErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "miniflux_entries_processing_errors_total",
		Help: "Total number of errors encountered during Miniflux entry processing",
	}, []string{"type"})

	// HTTPRequestsTotal counts HTTP requests by method, path and status
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests processed",
	}, []string{"method", "path", "status"})
)
