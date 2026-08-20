package oas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGetHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthStatus{Status: Ok})
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	resp, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

type fullSurfaceOps interface {
	ListTenants(http.ResponseWriter, *http.Request)
	ListClusters(http.ResponseWriter, *http.Request, string)
	ListCatalog(http.ResponseWriter, *http.Request, string, ListCatalogParams)
	ListInstances(http.ResponseWriter, *http.Request, string, ListInstancesParams)
	DeployCatalogItem(http.ResponseWriter, *http.Request, string)
	ListCapabilities(http.ResponseWriter, *http.Request, string, string)
	RenderInstallManifest(http.ResponseWriter, *http.Request, string, string)
}

var _ fullSurfaceOps = (ServerInterface)(nil)

func TestServerInterfaceCoversFullSurface(t *testing.T) {}
