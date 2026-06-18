package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/fandom-cli/fandom"
)

func (a *App) wikisCmd() *cobra.Command {
	var hub, search string

	cmd := &cobra.Command{
		Use:   "wikis",
		Short: "Discover Fandom wikis by hub or search query",
		Long: `Discover wikis hosted on fandom.com.

Use --search to find wikis matching a keyword (e.g. "star wars", "minecraft").
Use --hub to browse wikis by content category (Gaming, Movies, TV, Books, Comics,
Anime, Music, Lifestyle, Entertainment, Education).

Known hubs: ` + knownHubsStr() + `

Each record emits slug, name (when available), hub, and url.
Pipe output to "fandom --wiki <slug> info" or "fandom --wiki <slug> allpages" to explore a discovered wiki.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(25)
			var stubs []fandom.WikiStub
			var err error

			switch {
			case search != "":
				a.progressf("searching fandom.com wikis for %q...", search)
				stubs, err = a.client.SearchWikis(cmd.Context(), search, n)
			case hub != "":
				a.progressf("listing fandom.com wikis for hub %q...", hub)
				stubs, err = a.client.ListWikisByHub(cmd.Context(), hub, n)
			default:
				a.progressf("listing fandom.com wikis for all hubs...")
				for _, h := range fandom.KnownHubs {
					batch, berr := a.client.ListWikisByHub(cmd.Context(), h, 10)
					if berr != nil {
						continue
					}
					stubs = append(stubs, batch...)
					if len(stubs) >= n {
						stubs = stubs[:n]
						break
					}
				}
			}
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(stubs, len(stubs))
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search wikis by keyword")
	cmd.Flags().StringVar(&hub, "hub", "", "filter by hub (Gaming, Movies, TV, Books, Comics, Anime, ...)")
	return cmd
}

func knownHubsStr() string {
	s := ""
	for i, h := range fandom.KnownHubs {
		if i > 0 {
			s += ", "
		}
		s += h
	}
	return s
}
