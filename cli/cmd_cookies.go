package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tamnd/fandom-cli/fandom"
)

func (a *App) cookiesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cookies",
		Short: "Capture Cloudflare cookies from Chrome for a wiki",
		Long: `Opens a Chrome/Chromium window for the wiki, waits for the Cloudflare
challenge to pass, and caches the resulting cookies (including cf_clearance) so
all other fandom commands can use them automatically.

Run this once per wiki slug when you see HTTP 403 errors:

  fandom --wiki starwars cookies
  fandom --wiki minecraft cookies

Cookies are cached in ~/.cache/fandom-cli/ and expire after 20 hours.

If Chrome is already running with --remote-debugging-port=9222 the CLI will
connect to that session instead of launching a new browser window.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wiki := a.wiki
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "grabbing cookies for %s.fandom.com...\n", wiki)
			cookies, err := fandom.GrabCookies(wiki)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "cached %d cookies for %s\n", len(cookies), wiki)
			if v, ok := cookies["cf_clearance"]; ok {
				// Print just the header the user can pass via --cookie
				fmt.Printf("cf_clearance=%s\n", v)
			}
			return nil
		},
	}
	return cmd
}
