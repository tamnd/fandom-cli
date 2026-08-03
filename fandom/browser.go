package fandom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// CDP port used by CloakServer and by Chrome launched with --remote-debugging-port.
const cdpPort = "9222"

// ─── cookie cache ─────────────────────────────────────────────────────────────

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

// ─── CloakServer (CloakBrowser CDP server) ────────────────────────────────────

// CloakServerHandle holds the Docker container ID so the caller can stop it.
type CloakServerHandle struct {
	containerID string
}

// Stop terminates the CloakServer Docker container.
func (h *CloakServerHandle) Stop() {
	if h == nil || h.containerID == "" {
		return
	}
	_ = exec.Command("docker", "stop", h.containerID).Run()
}

// LaunchCloakServer starts a CloakBrowser CDP server in a detached Docker
// container and waits until it is accepting connections on localhost:9222.
//
// CloakBrowser (github.com/CloakHQ/cloakbrowser) is a modified Chromium build
// with 58 source-level C++ patches that defeat Cloudflare Turnstile, reCAPTCHA,
// and 30+ other bot-detection systems without any runtime injection.  The
// managed challenge that fandom.com uses auto-resolves in the first navigation.
//
// It requires Docker to be installed and running.  Returns an error if Docker is
// not available or the container fails to start.
func LaunchCloakServer() (*CloakServerHandle, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Pull image if not cached (silent, errors are non-fatal).
	pullCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pull := exec.CommandContext(pullCtx, "docker", "pull", "cloakhq/cloakbrowser")
	pull.Stdout = os.Stderr // progress to stderr
	pull.Stderr = os.Stderr
	_ = pull.Run()

	out, err := exec.Command(
		"docker", "run", "-d", "--rm",
		"-p", "127.0.0.1:"+cdpPort+":9222",
		"--shm-size=2g",
		"cloakhq/cloakbrowser",
		"cloakserve",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("docker run cloakbrowser: %w", err)
	}
	id := strings.TrimSpace(string(out))

	// Wait up to 30 s for the CDP server to accept connections.
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "localhost:"+cdpPort, time.Second)
		if err == nil {
			_ = conn.Close()
			return &CloakServerHandle{containerID: id}, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	_ = exec.Command("docker", "stop", id).Run()
	return nil, fmt.Errorf("cloakserve did not accept connections within 30s")
}

// IsCDPAvailable returns true when something is already listening on :9222.
func IsCDPAvailable() bool {
	conn, err := net.DialTimeout("tcp", "localhost:"+cdpPort, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// IsCloakServer returns true when the CDP endpoint at :9222 reports a CloakBrowser
// User-Agent in the /json/version response.
func IsCloakServer() bool {
	resp, err := http.Get("http://localhost:" + cdpPort + "/json/version")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return strings.Contains(strings.ToLower(string(body)), "cloak") ||
		strings.Contains(string(body), "CloakBrowser")
}

// ─── GrabCookies ──────────────────────────────────────────────────────────────

// GrabCookies obtains fandom.com cookies (including cf_clearance) for wiki.
//
// It uses this priority order:
//  1. CDP already running at localhost:9222 (CloakServer or Chrome with
//     --remote-debugging-port=9222) — navigate there and wait for CF to pass.
//  2. Docker is available — launch CloakServer automatically, wait for CF,
//     then stop the container.
//  3. Neither — print instructions and return an error.
//
// Cookies are cached in ~/.cache/fandom-cli/cookies-{wiki}.json (20h TTL).
// CloakBrowser resolves CF Turnstile managed challenges automatically; a plain
// Chrome debug session may take longer (or fail on high-risk IPs).
func GrabCookies(wiki string) (map[string]string, error) {
	if IsCDPAvailable() {
		fmt.Fprintf(os.Stderr, "connecting to CDP at localhost:%s...\n", cdpPort)
		cookies, err := grabFromCDP(wiki)
		if err != nil {
			return nil, err
		}
		return cookies, nil
	}

	// Try to start CloakServer via Docker.
	if _, err := exec.LookPath("docker"); err == nil {
		fmt.Fprintln(os.Stderr, "no CDP at :9222 — launching CloakBrowser via Docker...")
		handle, err := LaunchCloakServer()
		if err != nil {
			fmt.Fprintf(os.Stderr, "docker launch failed: %v\n", err)
		} else {
			defer handle.Stop()
			fmt.Fprintln(os.Stderr, "CloakServer ready")
			cookies, err := grabFromCDP(wiki)
			if err != nil {
				return nil, err
			}
			return cookies, nil
		}
	}

	// Give up with helpful instructions.
	return nil, fmt.Errorf(
		"no CDP endpoint available\n\n"+
			"Option A — CloakBrowser (recommended, bypasses CF automatically):\n"+
			"  docker run -p 127.0.0.1:9222:9222 --rm --shm-size=2g cloakhq/cloakbrowser cloakserve\n"+
			"  # in another terminal:\n"+
			"  fandom --wiki %s cookies\n\n"+
			"Option B — Chrome with remote debugging:\n"+
			"  open -a 'Google Chrome' --args \\\n"+
			"    --remote-debugging-port=9222 \\\n"+
			"    --remote-allow-origins='*' \\\n"+
			"    --user-data-dir=$HOME/.cache/fandom-cli/chrome-profile \\\n"+
			"    https://%s.fandom.com/\n"+
			"  # wait for the page to load, then:\n"+
			"  fandom --wiki %s cookies",
		wiki, wiki, wiki,
	)
}

// grabFromCDP connects to the CDP endpoint at :9222, navigates to the wiki,
// and polls until cf_clearance appears in the browser's cookie jar.
// CloakBrowser resolves the CF managed challenge automatically within ~5 s;
// a standard Chrome debug session may take up to 30 s on a trusted IP.
func grabFromCDP(wiki string) (map[string]string, error) {
	wsURL, err := launcher.ResolveURL("localhost:" + cdpPort)
	if err != nil {
		return nil, fmt.Errorf("resolve CDP at localhost:%s: %w", cdpPort, err)
	}

	b := rod.New().ControlURL(wsURL)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("connect to CDP: %w", err)
	}

	wikiURL := fmt.Sprintf("https://%s.fandom.com/", wiki)
	fmt.Fprintf(os.Stderr, "navigating to %s...\n", wikiURL)

	page, err := b.Page(proto.TargetCreateTarget{URL: wikiURL})
	if err != nil {
		return nil, fmt.Errorf("open page: %w", err)
	}

	// Poll until cf_clearance appears or timeout.
	// CloakBrowser solves CF Turnstile in ~5s; plain Chrome may need longer.
	timeout := 30 * time.Second
	if !IsCloakServer() {
		timeout = 90 * time.Second
	}
	fmt.Fprintf(os.Stderr, "waiting up to %s for Cloudflare challenge...\n", timeout.Round(time.Second))

	deadline := time.Now().Add(timeout)
	var ua string
	for time.Now().Before(deadline) {
		rawCookies, err := b.GetCookies()
		if err == nil {
			m := fandomCookies(rawCookies)
			if _, ok := m["cf_clearance"]; ok {
				// Capture the UA so the tls-client can send the same one.
				if pgs, err := b.Pages(); err == nil {
					for _, pg := range pgs {
						if res, err := pg.Eval(`() => navigator.userAgent`); err == nil {
							ua = res.Value.String()
							break
						}
					}
				}
				saveCookieCache(wiki, m, ua)
				fmt.Fprintf(os.Stderr, "cf_clearance obtained (%d cookies)\n", len(m))
				return m, nil
			}
		}
		time.Sleep(2 * time.Second)

		// If still on the challenge screen, try reloading.
		if info, err := page.Info(); err == nil {
			if strings.Contains(info.Title, "moment") || strings.Contains(info.Title, "Cloudflare") {
				// Only reload if URL is still on the target domain (not a redirect).
				if strings.Contains(info.URL, wiki+".fandom.com") {
					_ = page.Navigate(wikiURL)
				}
			}
		}
	}

	// Return whatever we have even without cf_clearance.
	rawCookies, _ := b.GetCookies()
	m := fandomCookies(rawCookies)
	if len(m) > 0 {
		fmt.Fprintf(os.Stderr, "got %d cookies (no cf_clearance — challenge may not have passed)\n", len(m))
		saveCookieCache(wiki, m, "")
		return m, nil
	}
	return nil, fmt.Errorf("no fandom.com cookies after %s — challenge did not pass", timeout.Round(time.Second))
}

// ─── getViaChrome ─────────────────────────────────────────────────────────────

// getViaChrome makes an HTTP GET by evaluating fetch() inside the CDP browser
// at :9222.  The request runs inside the browser's own TLS session so
// cf_clearance is attached automatically and the fingerprint matches.
//
// With CloakBrowser the CF challenge clears within seconds of navigation; with
// a plain Chrome debug session this waits up to 90 s.
func getViaChrome(ctx context.Context, rawURL string) ([]byte, error) {
	if !IsCDPAvailable() {
		return nil, fmt.Errorf(
			"CDP not available at localhost:%s\n"+
				"Run:  docker run -p 127.0.0.1:9222:9222 --rm --shm-size=2g cloakhq/cloakbrowser cloakserve",
			cdpPort,
		)
	}

	wsURL, err := launcher.ResolveURL("localhost:" + cdpPort)
	if err != nil {
		return nil, fmt.Errorf("resolve CDP: %w", err)
	}
	b := rod.New().ControlURL(wsURL)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("CDP connect: %w", err)
	}

	pages, err := b.Pages()
	if err != nil || len(pages) == 0 {
		return nil, fmt.Errorf("no pages open in CDP browser")
	}

	// Prefer a fandom.com page; fall back to whatever is open.
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

	// Wait for the CF challenge to clear.
	// CloakBrowser resolves it automatically; plain Chrome may take longer.
	cfTimeout := 90 * time.Second
	if IsCloakServer() {
		cfTimeout = 15 * time.Second
	}

	deadline := time.Now().Add(cfTimeout)
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
		if !strings.Contains(info.Title, "moment") && !strings.Contains(info.Title, "Cloudflare") {
			break
		}
		time.Sleep(2 * time.Second)
	}

	// Confirm we are past the challenge.
	if info, err := page.Info(); err == nil && strings.Contains(info.Title, "moment") {
		return nil, fmt.Errorf(
			"cloudflare challenge not cleared after %s\n"+
				"with CloakBrowser this should be automatic — try: fandom --wiki <slug> cookies --docker",
			cfTimeout.Round(time.Second),
		)
	}

	js := fmt.Sprintf(`fetch(%q, {credentials: 'include'}).then(r => r.text())`, rawURL)
	res, err := page.Eval(js)
	if err != nil {
		return nil, fmt.Errorf("CDP fetch(%q): %w", rawURL, err)
	}
	body := res.Value.String()
	if strings.Contains(body, "Just a moment") || strings.Contains(body, "cf-challenge") {
		return nil, fmt.Errorf("CDP fetch returned CF challenge page")
	}
	return []byte(body), nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

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
