package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	c "miniflux.app/v2/client"
)

// MockMinifluxClient is a mock of MinifluxClient interface.
type MockMinifluxClient struct {
	mock.Mock
}

func (m *MockMinifluxClient) Me() (*c.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*c.User), args.Error(1)
}

func (m *MockMinifluxClient) Entries(filter *c.Filter) (*c.EntryResultSet, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*c.EntryResultSet), args.Error(1)
}

func (m *MockMinifluxClient) UpdateEntries(entryIDs []int64, status string) error {
	args := m.Called(entryIDs, status)
	return args.Error(0)
}

func (m *MockMinifluxClient) SaveEntry(entryID int64) error {
	args := m.Called(entryID)
	return args.Error(0)
}

func TestHealthzHandler(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		mockClient.On("Me").Return(&c.User{ID: 1}, nil)

		s := &Server{client: mockClient}
		req := httptest.NewRequestWithContext(t.Context(), "GET", "/healthz", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.healthzHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})

	t.Run("Unhealthy - API Error", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		mockClient.On("Me").Return(nil, errors.New("api error"))

		s := &Server{client: mockClient}
		req := httptest.NewRequestWithContext(t.Context(), "GET", "/healthz", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.healthzHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unhealthy", response["status"])
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		s := &Server{}
		req := httptest.NewRequestWithContext(t.Context(), "POST", "/healthz", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.healthzHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestProcessEntriesHandler(t *testing.T) {
	t.Run("Successful processing", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)

		entries := &c.EntryResultSet{
			Total: 1,
			Entries: c.Entries{
				{ID: 123, Title: "Test Entry"},
			},
		}

		mockClient.On("Entries", mock.Anything).Return(entries, nil)
		mockClient.On("SaveEntry", int64(123)).Return(nil)
		mockClient.On("UpdateEntries", []int64{123}, c.EntryStatusRead).Return(nil)

		s := &Server{client: mockClient}
		req := httptest.NewRequestWithContext(t.Context(), "POST", "/process", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.processEntriesHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.InDelta(t, float64(1), response["processed"], 0.01)
		assert.InDelta(t, float64(0), response["errors"], 0.01)
		assert.InDelta(t, float64(1), response["total"], 0.01)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		s := &Server{}
		req := httptest.NewRequestWithContext(t.Context(), "GET", "/process", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.processEntriesHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestProcess(t *testing.T) {
	t.Run("Fetching entries error", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		mockClient.On("Entries", mock.Anything).Return(nil, errors.New("fetch error"))

		s := &Server{client: mockClient}
		processed, errs, total := s.Process(&c.Filter{})

		assert.Equal(t, 0, processed)
		assert.Equal(t, 0, errs)
		assert.Equal(t, 0, total)
	})

	t.Run("Save entry error", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		entries := &c.EntryResultSet{
			Total:   1,
			Entries: c.Entries{{ID: 1, Title: "Error Case"}},
		}
		mockClient.On("Entries", mock.Anything).Return(entries, nil)
		mockClient.On("SaveEntry", int64(1)).Return(errors.New("save error"))

		s := &Server{client: mockClient}
		processed, errs, total := s.Process(&c.Filter{})

		assert.Equal(t, 0, processed)
		assert.Equal(t, 1, errs)
		assert.Equal(t, 1, total)
	})

	testProcessErrors(t)

	t.Run("Skip if already running", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		s := &Server{client: mockClient}

		// Simulate a run already in progress.
		s.processing.Store(true)

		processed, errs, total := s.Process(&c.Filter{})

		assert.Equal(t, 0, processed)
		assert.Equal(t, 0, errs)
		assert.Equal(t, 0, total)
		// The client must not have been called.
		mockClient.AssertNotCalled(t, "Entries")
	})
}

func testProcessErrors(t *testing.T) {
	t.Run("Update entries error", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		entries := &c.EntryResultSet{
			Total:   1,
			Entries: c.Entries{{ID: 1, Title: "Update Error Case"}},
		}
		mockClient.On("Entries", mock.Anything).Return(entries, nil)
		mockClient.On("SaveEntry", int64(1)).Return(nil)
		// Correcting expectation for UpdateEntries:
		mockClient.On("UpdateEntries", []int64{1}, c.EntryStatusRead).Return(errors.New("update error"))

		s := &Server{client: mockClient}
		processed, errs, total := s.Process(&c.Filter{})

		assert.Equal(t, 0, processed)
		assert.Equal(t, 1, errs)
		assert.Equal(t, 1, total)
	})
}
