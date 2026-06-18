package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tamnd/fandom-cli/fandom"
)

func (a *App) cookiesCmd() *cobra.Command {
	var dockerMode bool
	var keepAlive bool

	cmd := &cobra.Command{
		Use:   "cookies",
		Short: "Grab Cloudflare cookies for a wiki and cache them",
		Long: `Opens a stealth browser session for the wiki, waits for the Cloudflare
challenge to resolve automatically, and caches the cookies (including
cf_clearance) so all other fandom commands work without the flag.

Run this once per wiki when you see HTTP 403 errors:

  fandom --wiki starwars cookies
  fandom --wiki minecraft cookies

Cookies are cached in ~/.cache/fandom-cli/ and expire after 20 hours.

BACKENDS (tried in order):

  1. CloakBrowser via Docker (--docker / auto if docker is available)
     CloakBrowser is a modified Chromium with 58 C++ patches that defeat
     Cloudflare Turnstile automatically.  No user interaction needed.

       docker pull cloakhq/cloakbrowser
       fandom --wiki starwars cookies --docker

  2. Chrome with remote debugging (if already running at localhost:9222)
     Start Chrome first with the debugging port:

       /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
         --remote-debugging-port=9222 \
         --remote-allow-origins='*' \
         --user-data-dir=$HOME/.cache/fandom-cli/chrome-profile \
         https://starwars.fandom.com/

     Then: fandom --wiki starwars cookies

After cookies are cached, all other commands use them automatically.
Use --browser to force every request through the stealth browser:

  fandom --wiki starwars --browser top
  fandom --wiki starwars --browser allpages | jq .`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wiki := a.wiki
			w := cmd.ErrOrStderr()

			// --docker: launch CloakServer, run grab, then stop unless --keep.
			if dockerMode {
				_, _ = fmt.Fprintf(w, "launching CloakBrowser via Docker...\n")
				handle, err := fandom.LaunchCloakServer()
				if err != nil {
					return fmt.Errorf("docker launch: %w", err)
				}
				if keepAlive {
					// Keep running and let the user ctrl-c.
					_, _ = fmt.Fprintf(w, "CloakServer running at localhost:9222 (ctrl-c to stop)\n")
					sig := make(chan os.Signal, 1)
					signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
					defer func() {
						<-sig
						handle.Stop()
					}()
				} else {
					defer handle.Stop()
				}
			}

			_, _ = fmt.Fprintf(w, "grabbing cookies for %s.fandom.com...\n", wiki)
			cookies, err := fandom.GrabCookies(wiki)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(w, "cached %d cookies for %s\n", len(cookies), wiki)
			if v, ok := cookies["cf_clearance"]; ok {
				fmt.Printf("cf_clearance=%s\n", v)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dockerMode, "docker", false, "launch CloakBrowser via Docker automatically")
	cmd.Flags().BoolVar(&keepAlive, "keep", false, "keep CloakServer running after cookies are grabbed (implies --docker)")
	return cmd
}
