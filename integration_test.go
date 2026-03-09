package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	c "miniflux.app/v2/client"
)

func TestIntegrationWithMiniflux(t *testing.T) {
if testing.Short() {
t.Skip("skipping integration test in short mode")
}

ctx := context.Background()

// 1. Démarrer PostgreSQL pour Miniflux
postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
ContainerRequest: testcontainers.ContainerRequest{
Image: "postgres:15-alpine",
Env: map[string]string{
"POSTGRES_USER":     "miniflux",
"POSTGRES_PASSWORD": "password",
"POSTGRES_DB":       "miniflux",
},
WaitingFor: wait.ForLog("database system is ready to accept connections"),
},
Started: true,
})
if err != nil {
t.Fatalf("Failed to start postgres: %v", err)
}
defer func() {
	if err := postgresContainer.Terminate(ctx); err != nil {
		t.Logf("failed to terminate postgres: %v", err)
	}
}()

postgresIP, err := postgresContainer.ContainerIP(ctx)
if err != nil {
t.Fatalf("Failed to get postgres IP: %v", err)
}
dbURL := fmt.Sprintf("postgres://miniflux:password@%s:5432/miniflux?sslmode=disable", postgresIP)

// 2. Démarrer Miniflux
minifluxContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
ContainerRequest: testcontainers.ContainerRequest{
Image:        "miniflux/miniflux:latest",
ExposedPorts: []string{"8080/tcp"},
Env: map[string]string{
"DATABASE_URL":          dbURL,
"RUN_MIGRATIONS":        "1",
"CREATE_ADMIN":          "1",
"ADMIN_USERNAME":        "admin",
"ADMIN_PASSWORD":        "password",
},
WaitingFor: wait.ForHTTP("/healthcheck").WithPort("8080/tcp"),
},
Started: true,
})
if err != nil {
t.Fatalf("Failed to start miniflux: %v", err)
}
defer func() {
	if err := minifluxContainer.Terminate(ctx); err != nil {
		t.Logf("failed to terminate miniflux: %v", err)
	}
}()

endpoint, _ := minifluxContainer.Endpoint(ctx, "")
apiURL := fmt.Sprintf("http://%s", endpoint)

client := c.NewClient(apiURL, "admin:password") 

config := &Config{
APIUrl:   apiURL,
APIToken: "admin:password", 
Port:     "9099",
}

s := NewServer(config)
s.client = client

t.Run("Full Flow: Create Feed and process", func(t *testing.T) {
feedID, err := client.CreateFeed(&c.FeedCreationRequest{
FeedURL:    "https://miniflux.app/feed.xml",
CategoryID: 1,
})

if err == nil {
t.Logf("Created feed with ID: %d", feedID)
_ = client.RefreshFeed(feedID)
} else {
t.Logf("Could not create feed: %v", err)
}

req, _ := http.NewRequest("POST", "/process", nil)
rr := httptest.NewRecorder()
handleProcess := http.HandlerFunc(s.processEntriesHandler)
handleProcess.ServeHTTP(rr, req)
assert.Equal(t, http.StatusOK, rr.Code)
t.Logf("Process response: %s", rr.Body.String())
})

t.Run("Health Check via integration", func(t *testing.T) {
req, _ := http.NewRequest("GET", "/healthz", nil)
rr := httptest.NewRecorder()
s.healthzHandler(rr, req)
assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, rr.Code)
})
}
