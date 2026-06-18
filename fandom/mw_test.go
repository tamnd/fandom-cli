package fandom_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAllPages(t *testing.T) {
	resp := map[string]any{
		"query": map[string]any{
			"allpages": []map[string]any{
				{"pageid": 1, "ns": 0, "title": "Jedi"},
				{"pageid": 2, "ns": 0, "title": "Sith"},
				{"pageid": 3, "ns": 0, "title": "Force"},
			},
		},
		"continue": map[string]any{
			"apcontinue": "G",
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "list=allpages") {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	stubs, next, err := c.AllPages(context.Background(), "starwars", "", 500)
	if err != nil {
		t.Fatal(err)
	}
	if len(stubs) != 3 {
		t.Fatalf("want 3 stubs, got %d", len(stubs))
	}
	if stubs[0].Title != "Jedi" {
		t.Errorf("title = %q, want Jedi", stubs[0].Title)
	}
	if stubs[0].ID != 1 {
		t.Errorf("id = %d, want 1", stubs[0].ID)
	}
	if next != "G" {
		t.Errorf("next token = %q, want G", next)
	}
}

func TestGetPage(t *testing.T) {
	resp := map[string]any{
		"query": map[string]any{
			"pages": map[string]any{
				"9875": map[string]any{
					"pageid":    9875,
					"ns":        0,
					"title":     "Lightsaber",
					"length":    4200,
					"lastrevid": 1001,
					"revisions": []map[string]any{
						{
							"revid":     1001,
							"user":      "JediEditor",
							"timestamp": "2024-03-01T12:00:00Z",
							"size":      4200,
							"slots": map[string]any{
								"main": map[string]any{
									"*":            "A '''lightsaber''' is an energy sword.\n\n== History ==\nInvented long ago.",
									"contentmodel": "wikitext",
								},
							},
						},
					},
					"categories": []map[string]any{
						{"ns": 14, "title": "Category:Weapons"},
					},
					"pageprops": map[string]any{},
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	art, err := c.GetPage(context.Background(), "starwars", "Lightsaber")
	if err != nil {
		t.Fatal(err)
	}
	if art.ID != 9875 {
		t.Errorf("id = %d, want 9875", art.ID)
	}
	if art.Title != "Lightsaber" {
		t.Errorf("title = %q, want Lightsaber", art.Title)
	}
	if art.LastEditor != "JediEditor" {
		t.Errorf("last_editor = %q, want JediEditor", art.LastEditor)
	}
	if art.RevisionID != 1001 {
		t.Errorf("revision_id = %d, want 1001", art.RevisionID)
	}
	if len(art.Categories) == 0 {
		t.Error("expected at least one category")
	}
	if art.Wikitext == "" {
		t.Error("expected non-empty wikitext")
	}
	if art.PlainText == "" {
		t.Error("expected non-empty plain_text")
	}
	if art.WordCount <= 0 {
		t.Errorf("word_count = %d, want > 0", art.WordCount)
	}
}

func TestGetPageNotFound(t *testing.T) {
	resp := map[string]any{
		"query": map[string]any{
			"pages": map[string]any{
				"-1": map[string]any{
					"ns":      0,
					"title":   "No Such Page",
					"missing": "",
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	_, err := c.GetPage(context.Background(), "starwars", "No Such Page")
	if err == nil {
		t.Fatal("expected error for missing page")
	}
}

func TestGetRevisions(t *testing.T) {
	resp := map[string]any{
		"query": map[string]any{
			"pages": map[string]any{
				"9875": map[string]any{
					"pageid": 9875,
					"ns":     0,
					"title":  "Lightsaber",
					"revisions": []map[string]any{
						{"revid": 1001, "user": "EditorA", "timestamp": "2024-03-01T12:00:00Z", "size": 4200, "comment": "initial"},
						{"revid": 1000, "user": "EditorB", "timestamp": "2024-02-28T10:00:00Z", "size": 3900, "comment": "fix typo"},
					},
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	revs, err := c.GetRevisions(context.Background(), "starwars", "Lightsaber", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 2 {
		t.Fatalf("want 2 revisions, got %d", len(revs))
	}
	if revs[0].Editor != "EditorA" {
		t.Errorf("editor = %q, want EditorA", revs[0].Editor)
	}
	if revs[0].RevID != 1001 {
		t.Errorf("rev_id = %d, want 1001", revs[0].RevID)
	}
}

func TestGetSiteInfo(t *testing.T) {
	resp := map[string]any{
		"query": map[string]any{
			"general": map[string]any{
				"sitename": "Wookieepedia",
				"base":     "https://starwars.fandom.com/wiki/Main_Page",
			},
			"statistics": map[string]any{
				"articles": 180000,
				"pages":    350000,
				"edits":    4200000,
				"images":   120000,
				"users":    50000,
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "meta=siteinfo") {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	info, err := c.GetSiteInfo(context.Background(), "starwars")
	if err != nil {
		t.Fatal(err)
	}
	if info.SiteName != "Wookieepedia" {
		t.Errorf("site_name = %q, want Wookieepedia", info.SiteName)
	}
	if info.Articles != 180000 {
		t.Errorf("articles = %d, want 180000", info.Articles)
	}
	if info.Wiki != "starwars" {
		t.Errorf("wiki = %q, want starwars", info.Wiki)
	}
}

func TestRecentChanges(t *testing.T) {
	resp := map[string]any{
		"query": map[string]any{
			"recentchanges": []map[string]any{
				{
					"type": "edit", "ns": 0, "title": "Luke Skywalker",
					"pageid": 1234, "revid": 5001, "user": "JediEditor",
					"timestamp": "2024-03-01T10:00:00Z", "comment": "updated infobox",
				},
				{
					"type": "new", "ns": 0, "title": "Baby Yoda",
					"pageid": 9999, "revid": 5002, "user": "MandalorianFan",
					"timestamp": "2024-03-01T09:00:00Z", "comment": "created page",
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "list=recentchanges") {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer srv.Close()
	c := testClient(srv)

	changes, err := c.RecentChanges(context.Background(), "starwars", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 {
		t.Fatalf("want 2 changes, got %d", len(changes))
	}
	if changes[0].Type != "edit" {
		t.Errorf("type = %q, want edit", changes[0].Type)
	}
	if changes[0].User != "JediEditor" {
		t.Errorf("user = %q, want JediEditor", changes[0].User)
	}
	if changes[1].Type != "new" {
		t.Errorf("type = %q, want new", changes[1].Type)
	}
}
