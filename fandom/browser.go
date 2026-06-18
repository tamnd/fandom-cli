package fandom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"context"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// cookieCache is the on-disk format for persisted fandom cookies.
type cookieCache struct {
	Cookies   map[string]string `json:"cookies"`
	UserAgent string            `json:"user_agent"`
	SavedAt   time.Time         `json:"saved_at"`
}

// cachePath returns the path to the cookie cache file for a wiki.
func cachePath(wiki string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "fandom-cli", "cookies-"+wiki+".json")
}

// LoadCookieCache reads the cached cookies and user-agent for wiki.
// Returns nil, "" when the file does not exist or is stale (>20h).
func LoadCookieCache(wiki string) (map[string]string, string) {
	data, err := os.ReadFile(cachePath(wiki))
	if err != nil {
		return nil, ""
	}
	var c cookieCache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, ""
	}
	if time.Since(c.SavedAt) > 20*time.Hour {
		return nil, ""
	}
	return c.Cookies, c.UserAgent
}

// saveCookieCache writes cookies to the cache file.
func saveCookieCache(wiki string, cookies map[string]string, ua string) {
	data, err := json.Marshal(cookieCache{
		Cookies:   cookies,
		UserAgent: ua,
		SavedAt:   time.Now(),
	})
	if err != nil {
		return
	}
	p := cachePath(wiki)
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, data, 0o600)
}

// CookieHeader converts a cookie map to a Cookie: header value.
func CookieHeader(cookies map[string]string) string {
	parts := make([]string, 0, len(cookies))
	for k, v := range cookies {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "; ")
}

// GrabCookies uses Chrome with CDP to obtain fandom.com cookies (including
// cf_clearance) for the given wiki slug.
//
// It requires Chrome to be running with remote debugging:
//
//	/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
//	  --remote-debugging-port=9222 \
//	  --remote-allow-origins='*' \
//	  --user-data-dir=$HOME/data/fandom/chrome-profile \
//	  https://<wiki>.fandom.com/
//
// Cookies are cached in ~/.cache/fandom-cli/cookies-{wiki}.json (20h TTL).
func GrabCookies(wiki string) (map[string]string, error) {
	cookies, err := grabFromCDP(wiki)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to grab cookies via Chrome CDP: %w\n\n"+
				"Start Chrome with remote debugging first:\n\n"+
				"  /Applications/Google\\ Chrome.app/Contents/MacOS/Google\\ Chrome \\\n"+
				"    --remote-debugging-port=9222 \\\n"+
				"    --remote-allow-origins='*' \\\n"+
				"    --user-data-dir=$HOME/data/fandom/chrome-profile \\\n"+
				"    https://%s.fandom.com/\n\n"+
				"Then retry: fandom --wiki %s cookies",
			err, wiki, wiki)
	}
	saveCookieCache(wiki, cookies, "")
	return cookies, nil
}

// grabFromCDP connects to Chrome at localhost:9222, opens the wiki page, and
// polls until cf_clearance appears (Chrome solves the JS challenge automatically
// when launched without --enable-automation).
func grabFromCDP(wiki string) (map[string]string, error) {
	wsURL, err := launcher.ResolveURL("localhost:9222")
	if err != nil {
		return nil, fmt.Errorf("chrome not found at localhost:9222: %w", err)
	}

	b := rod.New().ControlURL(wsURL)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("connect to Chrome: %w", err)
	}

	wikiURL := fmt.Sprintf("https://%s.fandom.com/", wiki)

	// Check existing tabs for an already-loaded fandom page.
	pages, err := b.Pages()
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}

	var fandomPage *rod.Page
	for _, pg := range pages {
		info, err := pg.Info()
		if err != nil {
			continue
		}
		if strings.Contains(info.URL, "fandom.com") {
			fandomPage = pg
			break
		}
	}

	// Open a new tab if no fandom tab is open.
	if fandomPage == nil {
		fmt.Printf("opening %s in Chrome...\n", wikiURL)
		fandomPage, err = b.Page(proto.TargetCreateTarget{URL: wikiURL})
		if err != nil {
			return nil, fmt.Errorf("open page: %w", err)
		}
	}

	// Poll up to 45 seconds for cf_clearance to appear.
	// Chrome solves the JS challenge on its own without user interaction when
	// the browser was launched without the --enable-automation flag.
	fmt.Printf("waiting for Cloudflare challenge on %s.fandom.com...\n", wiki)
	deadline := time.Now().Add(45 * time.Second)
	var ua string
	for time.Now().Before(deadline) {
		rawCookies, err := b.GetCookies()
		if err == nil {
			m := fandomCookies(rawCookies)
			if _, ok := m["cf_clearance"]; ok {
				// Capture the User-Agent from any open page so HTTP clients can
				// use the same UA that the cf_clearance was issued for.
				if pages, err := b.Pages(); err == nil {
					for _, pg := range pages {
						if res, err := pg.Eval(`() => navigator.userAgent`); err == nil {
							ua = res.Value.String()
							break
						}
					}
				}
				saveCookieCache(wiki, m, ua)
				fmt.Printf("cf_clearance obtained (%d fandom.com cookies)\n", len(m))
				return m, nil
			}
		}
		time.Sleep(2 * time.Second)
		// Re-navigate if still on the challenge screen.
		if info, err := fandomPage.Info(); err == nil {
			if strings.Contains(info.Title, "moment") || strings.Contains(info.Title, "Cloudflare") {
				if !strings.Contains(info.URL, "fandom.com") {
					_ = fandomPage.Navigate(wikiURL)
				}
			}
		}
	}

	// Grab whatever cookies we have even without cf_clearance.
	rawCookies, _ := b.GetCookies()
	m := fandomCookies(rawCookies)
	if len(m) > 0 {
		fmt.Printf("got %d fandom.com cookies (no cf_clearance — challenge may not have passed)\n", len(m))
		saveCookieCache(wiki, m, "")
		return m, nil
	}
	return nil, fmt.Errorf("no fandom.com cookies found after 45s")
}

// getViaChrome makes an HTTP GET request by evaluating fetch() in a Chrome tab
// connected via CDP at localhost:9222.  The request runs inside Chrome's own
// TLS session so cf_clearance is automatically sent and the fingerprint matches.
//
// If the open fandom tab is still on the CF challenge page, this waits up to
// 90 seconds for Chrome to solve it before making the API fetch.
func getViaChrome(ctx context.Context, rawURL string) ([]byte, error) {
	wsURL, err := launcher.ResolveURL("localhost:9222")
	if err != nil {
		return nil, fmt.Errorf("chrome not at :9222 (start Chrome with --remote-debugging-port=9222): %w", err)
	}
	b := rod.New().ControlURL(wsURL)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("chrome connect: %w", err)
	}

	pages, err := b.Pages()
	if err != nil || len(pages) == 0 {
		return nil, fmt.Errorf("no pages open in chrome")
	}

	// Pick a fandom.com page if available; otherwise use whatever is open.
	var page *rod.Page
	for _, pg := range pages {
		info, _ := pg.Info()
		if info != nil && strings.Contains(info.URL, "fandom.com") {
			page = pg
			break
		}
	}
	if page == nil {
		page = pages[0]
	}

	// Wait up to 90 s for the CF challenge to pass on the fandom tab.
	// When Chrome is pointed at a fandom.com page, the managed challenge should
	// resolve without user interaction once Chrome has a trusted session.
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		info, err := page.Info()
		if err != nil {
			break
		}
		onChallenge := strings.Contains(info.Title, "moment") ||
			strings.Contains(info.Title, "Cloudflare") ||
			strings.Contains(info.URL, "challenges.cloudflare.com")
		if !onChallenge {
			break
		}
		time.Sleep(3 * time.Second)
	}

	// Check one more time that we are not on the challenge page.
	if info, err := page.Info(); err == nil {
		if strings.Contains(info.Title, "moment") {
			return nil, fmt.Errorf("chrome: Cloudflare challenge still active after 90s — try visiting %s in Chrome manually", info.URL)
		}
	}

	js := fmt.Sprintf(
		`fetch(%q, {credentials: 'include'}).then(r => r.text())`,
		rawURL,
	)
	res, err := page.Eval(js)
	if err != nil {
		return nil, fmt.Errorf("chrome eval fetch(%q): %w", rawURL, err)
	}
	body := res.Value.String()
	if strings.Contains(body, "Just a moment") || strings.Contains(body, "cf-challenge") {
		return nil, fmt.Errorf("chrome: fetch returned CF challenge page")
	}
	return []byte(body), nil
}

// fandomCookies filters a raw cookie list to only fandom.com cookies.
func fandomCookies(cookies []*proto.NetworkCookie) map[string]string {
	m := make(map[string]string)
	for _, c := range cookies {
		if strings.Contains(c.Domain, "fandom.com") && c.Value != "" {
			m[c.Name] = c.Value
		}
	}
	return m
}
