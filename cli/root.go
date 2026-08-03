// Package cli builds the fandom command tree on top of the fandom library.
package cli

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/tamnd/fandom-cli/fandom"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// exit codes.
const (
	exitError  = 1
	exitUsage  = 2
	exitNoData = 3
)

// ExitError carries a process exit code up to main.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit %d", e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

func codeError(code int, err error) error { return &ExitError{Code: code, Err: err} }

// App holds shared state threaded through every command.
type App struct {
	client *fandom.Client
	cfg    fandom.Config

	// shared wiki flag
	wiki string

	output   string
	fields   []string
	noHeader bool
	template string
	limit    int
	quiet    bool
}

// Root builds the root command and its subtree.
func Root() *cobra.Command {
	app := &App{cfg: fandom.DefaultConfig()}

	root := &cobra.Command{
		Use:   "fandom",
		Short: "Browse Fandom wikis",
		Long: `fandom reads fan wikis hosted on fandom.com.
No API key is required. Use --wiki to select any of the 350,000+ Fandom wikis
by its subdomain slug (e.g. --wiki minecraft for minecraft.fandom.com).

Two API planes are used depending on the command:

  Fandom v1 API   — search, top, list, article, activity, info
  MediaWiki API   — page, allpages, revisions, siteinfo, recent, wikis

The "page" command fetches every field the MediaWiki API exposes: wikitext,
rendered Markdown, infobox key-value pairs, category list, internal and external
link graphs, images, templates, thumbnail, last editor, and revision ID.

The "allpages" command streams all page stubs in namespace 0, which is the BFS
seed for reconstructing an entire wiki without missing any page.

fandom is an independent tool and is not affiliated with Fandom, Inc.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return app.setup()
		},
	}

	pf := root.PersistentFlags()
	pf.StringVar(&app.wiki, "wiki", "starwars", "wiki subdomain (e.g. starwars, minecraft, harrypotter)")
	pf.StringVarP(&app.output, "output", "o", "auto", "output: table|json|jsonl|csv|tsv|url|raw (auto=table on TTY, jsonl piped)")
	pf.StringSliceVar(&app.fields, "fields", nil, "comma-separated columns to include")
	pf.BoolVar(&app.noHeader, "no-header", false, "omit the header row in table/csv/tsv")
	pf.StringVar(&app.template, "template", "", "Go text/template applied per record")
	pf.IntVarP(&app.limit, "limit", "n", 0, "limit number of records (0 = command default)")
	pf.BoolVarP(&app.quiet, "quiet", "q", false, "suppress progress on stderr")

	pf.DurationVar(&app.cfg.Rate, "delay", app.cfg.Rate, "minimum spacing between requests")
	pf.DurationVar(&app.cfg.Timeout, "timeout", app.cfg.Timeout, "per-request timeout")
	pf.IntVar(&app.cfg.Retries, "retries", app.cfg.Retries, "retry attempts on 429/5xx")
	pf.StringVar(&app.cfg.UserAgent, "user-agent", app.cfg.UserAgent, "User-Agent sent with each request")
	pf.StringVar(&app.cfg.Cookie, "cookie", "", "Cookie header value (e.g. cf_clearance=...) for Cloudflare-protected wikis")
	pf.BoolVar(&app.cfg.UseBrowser, "browser", false, "route all requests through Chrome CDP at localhost:9222 (bypasses Cloudflare)")
	pf.StringVar(&app.cfg.BaseURL, "base-url", "", "override wiki base URL (e.g. https://en.wikipedia.org/w for Wikipedia)")

	root.AddCommand(
		app.searchCmd(),
		app.topCmd(),
		app.listCmd(),
		app.articleCmd(),
		app.activityCmd(),
		app.infoCmd(),
		app.pageCmd(),
		app.allPagesCmd(),
		app.revisionsCmd(),
		app.siteInfoCmd(),
		app.recentCmd(),
		app.wikisCmd(),
		app.cookiesCmd(),
		newVersionCmd(),
	)
	return root
}

func (a *App) setup() error {
	if a.output == "" || a.output == "auto" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			a.output = string(FormatTable)
		} else {
			a.output = string(FormatJSONL)
		}
	}
	if !Format(a.output).Valid() {
		return codeError(exitUsage, fmt.Errorf("unknown output format %q", a.output))
	}
	a.cfg.Wiki = a.wiki
	// Auto-load cached cookies and UA from a prior `fandom cookies` run.
	if a.cfg.Cookie == "" {
		if cached, ua := fandom.LoadCookieCache(a.wiki); len(cached) > 0 {
			a.cfg.Cookie = fandom.CookieHeader(cached)
			if ua != "" && a.cfg.UserAgent == fandom.DefaultUserAgent {
				a.cfg.UserAgent = ua
			}
		}
	}
	a.client = fandom.NewClient(a.cfg)
	return nil
}

func (a *App) render(records any) error {
	r := NewRenderer(os.Stdout, Format(a.output), a.fields, a.noHeader, a.template)
	return r.Render(records)
}

func (a *App) renderOrEmpty(records any, n int) error {
	if err := a.render(records); err != nil {
		return err
	}
	if n == 0 {
		return codeError(exitNoData, nil)
	}
	return nil
}

func (a *App) progressf(format string, args ...any) {
	if a.quiet {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func mapFetchErr(err error) error {
	if err == nil {
		return nil
	}
	if isNotFound(err) {
		return codeError(exitNoData, err)
	}
	return codeError(exitError, err)
}

func (a *App) effectiveLimit(def int) int {
	if a.limit > 0 {
		return a.limit
	}
	return def
}
