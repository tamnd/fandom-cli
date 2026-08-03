// Package fandom is the library behind the fandom command: the HTTP client,
// request shaping, and the typed data models for Fandom wikis.
//
// The Fandom v1 API is open and requires no authentication. All endpoints live
// at https://{wiki}.fandom.com/api/v1/. The wiki slug is passed per-call so
// a single client can serve any of Fandom's 350,000+ wikis.
package fandom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// DefaultUserAgent mimics Chrome 146 so Cloudflare Bot Management does not
// flag the TLS fingerprint as a bot. The actual JA3/JA4 fingerprint is set by
// the bogdanfinn/tls-client Chrome_146 profile — this UA matches that version.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"

// ErrNotFound is returned when a requested article ID is not in the API response.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters for Client.
type Config struct {
	// BaseURL is used in tests to redirect all requests to a local server.
	// In production leave it empty or set to "https://%s.fandom.com".
	BaseURL   string
	Wiki      string
	UserAgent string
	Cookie    string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
	// UseBrowser routes all HTTP requests through Chrome CDP (fetch() eval)
	// instead of the tls-client.  Set automatically on first 403 fallback.
	UseBrowser bool
}

// DefaultConfig returns sensible production defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://%s.fandom.com",
		Wiki:      "starwars",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   15 * time.Second,
	}
}

// Client talks to the Fandom v1 API.
type Client struct {
	httpClient tls_client.HttpClient
	cfg        Config
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
// It uses a Chrome 131 TLS profile so Cloudflare Bot Management does not flag
// the connection as a non-browser based on JA3/JA4 fingerprint.
func NewClient(cfg Config) *Client {
	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}
	tlsc, _ := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithTimeoutSeconds(int(cfg.Timeout.Seconds())+5),
		tls_client.WithClientProfile(profiles.Chrome_146),
		tls_client.WithCookieJar(tls_client.NewCookieJar()),
	)
	return &Client{
		httpClient: tlsc,
		cfg:        cfg,
	}
}

// wikiBase returns the base URL for the given wiki slug.
// When BaseURL is set to a localhost test server, it is used as-is.
// When BaseURL is set to a custom remote URL (doesn't contain "fandom.com"),
// it is used directly, allowing the CLI to target any MediaWiki installation.
// Otherwise the default Fandom URL scheme is used.
func (c *Client) wikiBase(wiki string) string {
	base := c.cfg.BaseURL
	if base == "" {
		return fmt.Sprintf("https://%s.fandom.com", wiki)
	}
	if strings.HasPrefix(base, "http://127.0.0.1") ||
		strings.HasPrefix(base, "http://localhost") {
		return base
	}
	// Custom remote base URL (e.g. https://en.wikipedia.org/w or a self-hosted wiki).
	if !strings.Contains(base, "fandom.com") {
		return strings.TrimRight(base, "/")
	}
	return fmt.Sprintf("https://%s.fandom.com", wiki)
}

// effectiveWiki returns wiki if non-empty, else cfg.Wiki.
func (c *Client) effectiveWiki(wiki string) string {
	if wiki != "" {
		return wiki
	}
	return c.cfg.Wiki
}

// ─── public methods ───────────────────────────────────────────────────────────

// Search searches articles in a wiki using the Fandom v1 Search/List endpoint.
func (c *Client) Search(ctx context.Context, wiki, query string, limit int) ([]Article, error) {
	wiki = c.effectiveWiki(wiki)
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("namespaces", "0")

	rawURL := c.wikiBase(wiki) + "/api/v1/Search/List?" + params.Encode()
	var resp wireSearchResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, err
	}
	out := make([]Article, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, Article{
			ID:      it.ID,
			Title:   it.Title,
			URL:     it.URL,
			Quality: it.Quality,
			NS:      it.NS,
		})
	}
	return out, nil
}

// Top returns popular articles from a wiki using the Fandom v1 Articles/Top endpoint.
func (c *Client) Top(ctx context.Context, wiki string, limit int) ([]Article, error) {
	wiki = c.effectiveWiki(wiki)
	if limit <= 0 {
		limit = 25
	}
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("namespaces", "0")

	rawURL := c.wikiBase(wiki) + "/api/v1/Articles/Top?" + params.Encode()
	var resp wireTopResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, err
	}
	base := resp.BasePath
	out := make([]Article, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, Article{
			ID:       it.ID,
			Title:    it.Title,
			URL:      fullURL(base, it.URL),
			Abstract: it.Abstract,
			Quality:  it.Quality,
			NS:       it.NS,
		})
	}
	return out, nil
}

// List returns a paginated list of articles using the Fandom v1 Articles/List endpoint.
func (c *Client) List(ctx context.Context, wiki string, limit, offset int) ([]Article, error) {
	wiki = c.effectiveWiki(wiki)
	if limit <= 0 {
		limit = 25
	}
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("namespaces", "0")

	rawURL := c.wikiBase(wiki) + "/api/v1/Articles/List?" + params.Encode()
	var resp wireArticleListResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, err
	}
	base := resp.BasePath
	out := make([]Article, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, Article{
			ID:       it.ID,
			Title:    it.Title,
			URL:      fullURL(base, it.URL),
			Abstract: it.Abstract,
			Quality:  it.Quality,
			NS:       it.NS,
		})
	}
	return out, nil
}

// GetArticle fetches details for a single article by numeric ID.
// Returns ErrNotFound if the ID is not in the response.
func (c *Client) GetArticle(ctx context.Context, wiki string, id int) (Article, error) {
	wiki = c.effectiveWiki(wiki)
	params := url.Values{}
	params.Set("ids", strconv.Itoa(id))
	params.Set("abstract", "500")

	rawURL := c.wikiBase(wiki) + "/api/v1/Articles/Details?" + params.Encode()
	var resp wireDetailsResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return Article{}, err
	}
	key := strconv.Itoa(id)
	it, ok := resp.Items[key]
	if !ok {
		return Article{}, fmt.Errorf("article %d: %w", id, ErrNotFound)
	}
	return Article{
		ID:       it.ID,
		Title:    it.Title,
		URL:      fullURL(resp.BasePath, it.URL),
		Abstract: it.Abstract,
	}, nil
}

// Activity returns recent edit activity from a wiki.
func (c *Client) Activity(ctx context.Context, wiki string, limit int) ([]ActivityItem, error) {
	wiki = c.effectiveWiki(wiki)
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("allowDuplicates", "false")

	rawURL := c.wikiBase(wiki) + "/api/v1/Activity/LatestActivity?" + params.Encode()
	var resp wireActivityResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, err
	}
	out := make([]ActivityItem, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, ActivityItem{
			ArticleID:    it.ArticleID,
			ArticleTitle: it.Article.Title,
			RevisionID:   it.RevisionID,
			Timestamp:    it.Timestamp,
			User:         it.User.Name,
			URL:          it.Article.URL,
		})
	}
	return out, nil
}

// Info returns metadata for a wiki using the Mercury/WikiVariables endpoint.
func (c *Client) Info(ctx context.Context, wiki string) (WikiInfo, error) {
	wiki = c.effectiveWiki(wiki)
	rawURL := c.wikiBase(wiki) + "/api/v1/Mercury/WikiVariables"
	var resp wireWikiVarsResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return WikiInfo{}, err
	}
	return WikiInfo{
		SiteName: resp.Data.SiteName,
		BasePath: resp.Data.BasePath,
		Lang:     resp.Data.Lang,
		Topic:    resp.Data.Topic,
		Wiki:     wiki,
	}, nil
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────

func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	// Skip direct HTTP if we're routing through Chrome (CF-protected wikis).
	// Chrome CDP keeps the same TLS session that obtained cf_clearance, so the
	// request passes CF checks that a separate HTTP client cannot replicate.
	if c.cfg.UseBrowser {
		return getViaChrome(ctx, rawURL)
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		// On 403, try routing through Chrome CDP (same session as cf_clearance).
		if strings.Contains(err.Error(), "403") {
			if b, cerr := getViaChrome(ctx, rawURL); cerr == nil {
				c.cfg.UseBrowser = true // switch permanently for this session
				return b, nil
			}
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := fhttp.NewRequestWithContext(ctx, fhttp.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header = fhttp.Header{
		"User-Agent":         []string{c.cfg.UserAgent},
		"Accept":             []string{"application/json, text/html, */*; q=0.9"},
		"Accept-Language":    []string{"en-US,en;q=0.9"},
		"Accept-Encoding":    []string{"gzip, deflate, br"},
		fhttp.HeaderOrderKey: {"user-agent", "accept", "accept-language", "accept-encoding"},
	}
	if c.cfg.Cookie != "" {
		req.Header.Set("Cookie", c.cfg.Cookie)
		req.Header[fhttp.HeaderOrderKey] = append(req.Header[fhttp.HeaderOrderKey], "cookie")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 429 || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// fullURL joins basePath and a possibly-relative URL.
func fullURL(base, u string) string {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(u, "/")
}
