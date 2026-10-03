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

func qaClient(t *testing.T, h http.HandlerFunc) *ClientWithResponses {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := NewClientWithResponses(srv.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}
	return c
}

func TestClientListClustersRoundTrip(t *testing.T) {
	c := qaClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tenants/acme/clusters" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"clusters":[{"id":"c1","name":"prod","orgId":"acme","state":"connected","createdAt":"2026-01-01T00:00:00Z","labels":null,"distribution":null}]}`))
	})
	resp, err := c.ListClustersWithResponse(context.Background(), "acme")
	if err != nil {
		t.Fatalf("ListClusters: %v", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil || resp.JSON200.Clusters == nil || len(*resp.JSON200.Clusters) != 1 {
		t.Fatalf("bad response: %d %+v", resp.StatusCode(), resp.JSON200)
	}
	if got := (*resp.JSON200.Clusters)[0].Name; got != "prod" {
		t.Fatalf("cluster name = %q", got)
	}
}

func TestClientPathParamsEscaped(t *testing.T) {
	var gotPath string
	c := qaClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"clusters":[]}`))
	})
	if _, err := c.ListClustersWithResponse(context.Background(), "a/b"); err != nil {
		t.Fatalf("ListClusters: %v", err)
	}
	if gotPath != "/api/v1/tenants/a%2Fb/clusters" {
		t.Fatalf("escaped path = %s", gotPath)
	}
}

func TestClientErrorModelParsed(t *testing.T) {
	c := qaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"title": "not found", "detail": "no such cluster"})
	})
	resp, err := c.GetClusterWithResponse(context.Background(), "acme", "nope")
	if err != nil {
		t.Fatalf("GetCluster: %v", err)
	}
	if resp.StatusCode() != http.StatusNotFound || resp.ApplicationproblemJSONDefault == nil {
		t.Fatalf("status = %d, error model = %+v", resp.StatusCode(), resp.ApplicationproblemJSONDefault)
	}
}

func TestClientMalformedJSON(t *testing.T) {
	c := qaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not json`))
	})
	if _, err := c.ListClustersWithResponse(context.Background(), "acme"); err == nil {
		t.Fatal("expected error on malformed JSON, got nil")
	}
}

// TestGetClusterKubeconfigRequestPathAndParams verifies the generated request
// builder uses the correct path and serializes the new query parameters.
func TestGetClusterKubeconfigRequestPathAndParams(t *testing.T) {
	mode := GetClusterKubeconfigParamsModeGateway
	grant := DeviceCode
	server := "https://k8s.example.com"
	req, err := NewGetClusterKubeconfigRequest(
		"https://api.inari.example",
		"acme",
		"cluster-1",
		&GetClusterKubeconfigParams{
			Mode:      &mode,
			GrantType: &grant,
			Server:    &server,
		},
	)
	if err != nil {
		t.Fatalf("NewGetClusterKubeconfigRequest: %v", err)
	}
	if req.URL.Path != "/api/v1/tenants/acme/clusters/cluster-1/kubeconfig" {
		t.Fatalf("unexpected path: %s", req.URL.Path)
	}
	q := req.URL.Query()
	if q.Get("mode") != "gateway" {
		t.Fatalf("mode param mismatch: %q", q.Get("mode"))
	}
	if q.Get("grantType") != "device-code" {
		t.Fatalf("grantType param mismatch: %q", q.Get("grantType"))
	}
	if q.Get("server") != "https://k8s.example.com" {
		t.Fatalf("server param mismatch: %q", q.Get("server"))
	}
}

// TestGetClusterKubeconfigRequestNoParams verifies that nil parameters omit
// the query string entirely, letting the server apply its defaults.
func TestGetClusterKubeconfigRequestNoParams(t *testing.T) {
	req, err := NewGetClusterKubeconfigRequest(
		"https://api.inari.example",
		"acme",
		"cluster-1",
		nil,
	)
	if err != nil {
		t.Fatalf("NewGetClusterKubeconfigRequest: %v", err)
	}
	if req.URL.RawQuery != "" {
		t.Fatalf("expected empty query string, got %q", req.URL.RawQuery)
	}
}

// TestClusterAccessInfoNewFieldsRoundTrip verifies that the new optional
// ClusterAccessInfo fields marshal/unmarshal through the generated model.
func TestClusterAccessInfoNewFieldsRoundTrip(t *testing.T) {
	trueValue := true
	info := ClusterAccessInfo{
		IssuerUrl:               "https://keycloak.example/realms/inari",
		KubectlClientId:         "cluster-cluster-1",
		Audience:                "inari-server",
		Organization:            "acme",
		KubectlAccessEnabled:    &trueValue,
		ProxyUrl:                strPtr("https://kubeproxy.example"),
		TunnelAvailable:         &trueValue,
		TunnelUnavailableReason: strPtr("agent not upgraded"),
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ClusterAccessInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.KubectlAccessEnabled == nil || *got.KubectlAccessEnabled != true {
		t.Fatalf("kubectlAccessEnabled mismatch: got %v", got.KubectlAccessEnabled)
	}
	if got.ProxyUrl == nil || *got.ProxyUrl != "https://kubeproxy.example" {
		t.Fatalf("proxyUrl mismatch: got %v", got.ProxyUrl)
	}
	if got.TunnelAvailable == nil || *got.TunnelAvailable != true {
		t.Fatalf("tunnelAvailable mismatch: got %v", got.TunnelAvailable)
	}
	if got.TunnelUnavailableReason == nil || *got.TunnelUnavailableReason != "agent not upgraded" {
		t.Fatalf("tunnelUnavailableReason mismatch: got %v", got.TunnelUnavailableReason)
	}
}

func strPtr(s string) *string { return &s }
