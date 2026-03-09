package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	c "miniflux.app/v2/client"
)

// MockMinifluxClient is a mock of MinifluxClient interface
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
		req, _ := http.NewRequest("GET", "/healthz", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.healthzHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})

	t.Run("Unhealthy - API Error", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		mockClient.On("Me").Return(nil, errors.New("api error"))

		s := &Server{client: mockClient}
		req, _ := http.NewRequest("GET", "/healthz", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.healthzHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "unhealthy", response["status"])
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		s := &Server{}
		req, _ := http.NewRequest("POST", "/healthz", nil)
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
		req, _ := http.NewRequest("POST", "/process", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(s.processEntriesHandler)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["processed"])
		assert.Equal(t, float64(0), response["errors"])
		assert.Equal(t, float64(1), response["total"])
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		s := &Server{}
		req, _ := http.NewRequest("GET", "/process", nil)
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

	t.Run("Update entries error", func(t *testing.T) {
		mockClient := new(MockMinifluxClient)
		entries := &c.EntryResultSet{
			Total:   1,
			Entries: c.Entries{{ID: 1, Title: "Update Error Case"}},
		}
		mockClient.On("Entries", mock.Anything).Return(entries, nil)
		mockClient.On("SaveEntry", int64(1)).Return(nil)
		mockClient.On("UpdateEntries", []int64{123}, c.EntryStatusRead).Return(errors.New("update error")).Maybe() // IDs might differ in actual code call
		// Correcting expectation for UpdateEntries:
		mockClient.On("UpdateEntries", []int64{1}, c.EntryStatusRead).Return(errors.New("update error"))

		s := &Server{client: mockClient}
		processed, errs, total := s.Process(&c.Filter{})

		assert.Equal(t, 0, processed)
		assert.Equal(t, 1, errs)
		assert.Equal(t, 1, total)
	})
}
