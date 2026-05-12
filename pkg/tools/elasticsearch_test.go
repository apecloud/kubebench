package tools

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestElasticsearchCheckConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "elastic" || password != "secret" {
			t.Fatalf("unexpected auth: %s %s %t", user, password, ok)
		}
		if r.URL.Path != "/_cluster/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"green"}`))
	}))
	defer server.Close()

	client := elasticsearchClientFromURL(t, server.URL)
	client.Username = "elastic"
	client.Password = "secret"
	if err := client.InitClient(); err != nil {
		t.Fatal(err)
	}
	if err := client.CheckConnection(); err != nil {
		t.Fatal(err)
	}
}

func TestElasticsearchCheckConnectionStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := elasticsearchClientFromURL(t, server.URL)
	if err := client.InitClient(); err != nil {
		t.Fatal(err)
	}
	if err := client.CheckConnection(); err == nil {
		t.Fatal("expected status error")
	}
}

func elasticsearchClientFromURL(t *testing.T, rawURL string) *ElasticsearchClient {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	return &ElasticsearchClient{
		Scheme: u.Scheme,
		Host:   u.Hostname(),
		Port:   port,
	}
}
