package fandom_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/fandom-cli/fandom"
)

// testClient returns a Client pointed at ts with rate limiting disabled.
func testClient(ts *httptest.Server) *fandom.Client {
	cfg := fandom.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return fandom.NewClient(cfg)
}

func TestUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()
	c := testClient(srv)
	_, _ = c.Search(context.Background(), "", "test", 1)
}

func TestSearch(t *testing.T) {
	resp := map[string]any{
		"items": []map[string]any{
			{"id": 1234, "title": "Luke Skywalker", "url": "/wiki/Luke_Skywalker", "quality": 99, "ns": 0},
			{"id": 5678, "title": "Darth Vader", "url": "/wiki/Darth_Vader", "quality": 95, "ns": 0},
		},
		"batches": 1, "currentBatch": 1,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "Search/List") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	items, err := c.Search(context.Background(), "starwars", "luke", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if items[0].Title != "Luke Skywalker" {
		t.Errorf("first title = %q, want Luke Skywalker", items[0].Title)
	}
	if items[0].ID != 1234 {
		t.Errorf("first id = %d, want 1234", items[0].ID)
	}
}

func TestTop(t *testing.T) {
	resp := map[string]any{
		"items": []map[string]any{
			{"id": 1, "title": "Main Page", "url": "/wiki/Main_Page", "quality": 0, "ns": 0},
			{"id": 2, "title": "Luke Skywalker", "url": "/wiki/Luke_Skywalker", "quality": 99, "ns": 0},
		},
		"basepath": "https://starwars.fandom.com",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "Articles/Top") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	items, err := c.Top(context.Background(), "starwars", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if items[1].ID != 2 {
		t.Errorf("second id = %d, want 2", items[1].ID)
	}
}

func TestList(t *testing.T) {
	resp := map[string]any{
		"items": []map[string]any{
			{"id": 10, "title": "Alderan", "url": "/wiki/Alderan", "quality": 80, "ns": 0, "abstract": "A peaceful planet."},
			{"id": 11, "title": "Bespin", "url": "/wiki/Bespin", "quality": 75, "ns": 0, "abstract": "A gas giant."},
			{"id": 12, "title": "Coruscant", "url": "/wiki/Coruscant", "quality": 90, "ns": 0, "abstract": "The capital."},
		},
		"basepath": "https://starwars.fandom.com",
		"offset":   25,
		"limit":    25,
		"total":    172000,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "Articles/List") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	items, err := c.List(context.Background(), "starwars", 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %d", len(items))
	}
	if items[2].Title != "Coruscant" {
		t.Errorf("third title = %q, want Coruscant", items[2].Title)
	}
}

func TestGetArticle(t *testing.T) {
	resp := map[string]any{
		"items": map[string]any{
			"9875": map[string]any{
				"id":       9875,
				"title":    "Lightsaber",
				"abstract": "A lightsaber is an energy sword.",
				"url":      "/wiki/Lightsaber",
			},
		},
		"basepath": "https://starwars.fandom.com",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "Articles/Details") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	art, err := c.GetArticle(context.Background(), "starwars", 9875)
	if err != nil {
		t.Fatal(err)
	}
	if art.Title != "Lightsaber" {
		t.Errorf("title = %q, want Lightsaber", art.Title)
	}
	if art.Abstract != "A lightsaber is an energy sword." {
		t.Errorf("abstract = %q", art.Abstract)
	}
}

func TestGetArticleNotFound(t *testing.T) {
	resp := map[string]any{
		"items":    map[string]any{},
		"basepath": "https://starwars.fandom.com",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	_, err := c.GetArticle(context.Background(), "starwars", 99999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want contains 'not found'", err.Error())
	}
}

func TestActivity(t *testing.T) {
	resp := map[string]any{
		"items": []map[string]any{
			{
				"articleId":  1234,
				"revisionId": 56789,
				"timestamp":  "2024-01-15T10:30:00Z",
				"user":       map[string]any{"name": "JediMaster99"},
				"article":    map[string]any{"id": 1234, "title": "Luke Skywalker", "url": "/wiki/Luke_Skywalker"},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "Activity/LatestActivity") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	items, err := c.Activity(context.Background(), "starwars", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if items[0].User != "JediMaster99" {
		t.Errorf("user = %q, want JediMaster99", items[0].User)
	}
	if items[0].ArticleTitle != "Luke Skywalker" {
		t.Errorf("article title = %q, want Luke Skywalker", items[0].ArticleTitle)
	}
}

func TestInfo(t *testing.T) {
	resp := map[string]any{
		"data": map[string]any{
			"basePath": "https://starwars.fandom.com",
			"siteName": "Wookieepedia",
			"lang":     "en",
			"topic":    "Star Wars",
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "Mercury/WikiVariables") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	info, err := c.Info(context.Background(), "starwars")
	if err != nil {
		t.Fatal(err)
	}
	if info.SiteName != "Wookieepedia" {
		t.Errorf("site_name = %q, want Wookieepedia", info.SiteName)
	}
	if info.Lang != "en" {
		t.Errorf("lang = %q, want en", info.Lang)
	}
	if info.Topic != "Star Wars" {
		t.Errorf("topic = %q, want Star Wars", info.Topic)
	}
	if info.Wiki != "starwars" {
		t.Errorf("wiki = %q, want starwars", info.Wiki)
	}
}

func TestRetryOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"siteName":"Test","basePath":"","lang":"en","topic":""}}`))
	}))
	defer srv.Close()

	cfg := fandom.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := fandom.NewClient(cfg)

	_, err := c.Info(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}
