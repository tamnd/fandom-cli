package fandom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// KnownHubs is the curated list of Fandom content hubs.
var KnownHubs = []string{
	"Gaming", "Movies", "TV", "Books", "Comics", "Anime", "Music",
	"Lifestyle", "Entertainment", "Education",
}

// SearchWikis searches for wikis matching query.
// It first tries the Fandom f2 JSON feed, then falls back to scraping
// the fandom.com search page for wiki slugs.
func (c *Client) SearchWikis(ctx context.Context, query string, limit int) ([]WikiStub, error) {
	if limit <= 0 {
		limit = 25
	}
	feedURL := fmt.Sprintf("https://www.fandom.com/f2/api/public/feed?topics[]=%s&page=1",
		url.QueryEscape(query))

	var stubs []WikiStub

	body, err := c.get(ctx, feedURL)
	if err == nil {
		stubs = parseFeedJSON(body)
	}

	if len(stubs) == 0 {
		// HTML fallback: scrape search results page.
		searchURL := "https://www.fandom.com/search?query=" + url.QueryEscape(query)
		body2, err2 := c.get(ctx, searchURL)
		if err2 == nil {
			stubs = scrapeFandomSlugs(string(body2))
		}
	}

	if len(stubs) > limit {
		stubs = stubs[:limit]
	}
	return stubs, nil
}

// ListWikisByHub returns wikis for a given hub by scraping the Fandom topic page.
// It also tries the f2 feed as a secondary source.
func (c *Client) ListWikisByHub(ctx context.Context, hub string, limit int) ([]WikiStub, error) {
	if limit <= 0 {
		limit = 50
	}

	topicURL := hubTopicURL(hub)
	body, err := c.get(ctx, topicURL)
	if err != nil {
		return nil, fmt.Errorf("fetch hub %q: %w", hub, err)
	}

	stubs := scrapeFandomSlugs(string(body))

	// Supplement via f2 feed if needed.
	if len(stubs) < limit {
		feedURL := fmt.Sprintf("https://www.fandom.com/f2/api/public/feed?topics[]=%s&page=1",
			url.QueryEscape(strings.ToLower(hub)))
		if body2, err2 := c.get(ctx, feedURL); err2 == nil {
			for _, s := range parseFeedJSON(body2) {
				stubs = appendUnique(stubs, s)
			}
		}
	}

	if len(stubs) > limit {
		stubs = stubs[:limit]
	}
	return stubs, nil
}

// scrapeFandomSlugs extracts unique wiki stubs from arbitrary HTML/JSON content
// by scanning for *.fandom.com subdomains.
func scrapeFandomSlugs(content string) []WikiStub {
	seen := make(map[string]bool)
	var stubs []WikiStub

	for i := 0; i < len(content)-11; i++ {
		if content[i] != '.' || content[i:i+11] != ".fandom.com" {
			continue
		}
		start := i - 1
		for start >= 0 && isSlugByte(content[start]) {
			start--
		}
		start++
		if start >= i {
			i += 10
			continue
		}
		slug := content[start:i]
		if len(slug) < 2 || slugBlocklist[slug] || seen[slug] {
			i += 10
			continue
		}
		seen[slug] = true
		stubs = append(stubs, WikiStub{
			Slug: slug,
			URL:  "https://" + slug + ".fandom.com",
		})
		i += 10
	}
	return stubs
}

// parseFeedJSON parses wiki stubs from Fandom's f2 feed JSON response.
// Falls back to slug extraction on parse failure.
func parseFeedJSON(body []byte) []WikiStub {
	var resp struct {
		Wikis []struct {
			Name   string `json:"name"`
			Domain string `json:"domain"`
			Hub    string `json:"hub"`
		} `json:"wikis"`
		Items []struct {
			Wiki struct {
				Name   string `json:"name"`
				Domain string `json:"domain"`
			} `json:"wiki"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &resp); err == nil && (len(resp.Wikis) > 0 || len(resp.Items) > 0) {
		var stubs []WikiStub
		for _, w := range resp.Wikis {
			slug := slugFromDomain(w.Domain)
			if slug != "" {
				stubs = append(stubs, WikiStub{
					Slug: slug,
					Name: w.Name,
					Hub:  w.Hub,
					URL:  "https://" + slug + ".fandom.com",
				})
			}
		}
		for _, it := range resp.Items {
			slug := slugFromDomain(it.Wiki.Domain)
			if slug != "" {
				stubs = append(stubs, WikiStub{
					Slug: slug,
					Name: it.Wiki.Name,
					URL:  "https://" + slug + ".fandom.com",
				})
			}
		}
		if len(stubs) > 0 {
			return stubs
		}
	}
	// Fallback: extract slugs from raw JSON bytes.
	return scrapeFandomSlugs(string(body))
}

// slugFromDomain derives a wiki slug from its domain (e.g. "starwars.fandom.com" → "starwars").
func slugFromDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	if parts := strings.SplitN(domain, ".", 2); len(parts) >= 1 {
		s := parts[0]
		if len(s) >= 2 && !slugBlocklist[s] {
			return s
		}
	}
	return ""
}

// hubTopicURL maps a hub name to a fandom.com topic browse URL.
func hubTopicURL(hub string) string {
	m := map[string]string{
		"Gaming":        "https://www.fandom.com/topics/gaming",
		"Movies":        "https://www.fandom.com/topics/movies",
		"TV":            "https://www.fandom.com/topics/tv",
		"Books":         "https://www.fandom.com/topics/books",
		"Comics":        "https://www.fandom.com/topics/comics",
		"Anime":         "https://www.fandom.com/topics/anime-manga",
		"Music":         "https://www.fandom.com/topics/music",
		"Lifestyle":     "https://www.fandom.com/topics/lifestyle",
		"Entertainment": "https://www.fandom.com/topics/entertainment",
		"Education":     "https://www.fandom.com/topics/education",
	}
	if u, ok := m[hub]; ok {
		return u
	}
	return "https://www.fandom.com/topics/" + strings.ToLower(hub)
}

// slugBlocklist contains Fandom infrastructure subdomains (not actual wikis).
var slugBlocklist = map[string]bool{
	"www": true, "community": true, "services": true, "static": true,
	"script": true, "beacon": true, "cdn-gl": true, "auth": true,
	"about": true, "iframes": true, "widget": true, "consent": true,
	"storage": true, "img": true, "support": true, "help": true,
}

func isSlugByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '-'
}

func appendUnique(stubs []WikiStub, s WikiStub) []WikiStub {
	for _, existing := range stubs {
		if existing.Slug == s.Slug {
			return stubs
		}
	}
	return append(stubs, s)
}
