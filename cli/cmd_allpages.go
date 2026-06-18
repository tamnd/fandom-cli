package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) allPagesCmd() *cobra.Command {
	var continueToken string

	cmd := &cobra.Command{
		Use:   "allpages",
		Short: "Stream all page stubs from a wiki (BFS seed)",
		Long: `Stream every page stub in namespace 0 from a wiki using the MediaWiki allpages API.

Each record emits id, ns, title, and url. Without --limit the command paginates
through all pages automatically, which is the BFS seed for a full-wiki crawl.

Pipe to jq or another tool to filter by namespace, title prefix, or ID range:

  fandom allpages --wiki minecraft | jq 'select(.title | startswith("Creeper"))' | ...`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			limit := a.limit
			total := 0

			token := continueToken
			for {
				batch := 500
				if limit > 0 {
					remaining := limit - total
					if remaining <= 0 {
						break
					}
					if remaining < batch {
						batch = remaining
					}
				}

				a.progressf("fetching allpages from %s.fandom.com (offset token=%q)...", a.wiki, token)
				stubs, next, err := a.client.AllPages(cmd.Context(), a.wiki, token, batch)
				if err != nil {
					return mapFetchErr(err)
				}
				if len(stubs) == 0 {
					break
				}

				for _, s := range stubs {
					if err := a.render(s); err != nil {
						return err
					}
					total++
					if limit > 0 && total >= limit {
						break
					}
				}

				if next == "" || (limit > 0 && total >= limit) {
					break
				}
				token = next
			}

			if total == 0 {
				return codeError(exitNoData, nil)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&continueToken, "continue", "", "continue token from a previous run")
	return cmd
}
