package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) pageCmd() *cobra.Command {
	var noWikitext bool

	cmd := &cobra.Command{
		Use:   "page <title>",
		Short: "Fetch full article by title (wikitext, categories, links, infobox)",
		Long: `Fetch a full article by title using the MediaWiki Action API.

Returns every field the API exposes: wikitext source, rendered Markdown,
category list, internal and external link graphs, infobox key-value pairs,
images, templates, thumbnail, last editor, revision ID, and word count.

Use --no-wikitext to suppress the raw wikitext field for compact output.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			a.progressf("fetching page %q from %s.fandom.com...", title, a.wiki)
			art, err := a.client.GetPage(cmd.Context(), a.wiki, title)
			if err != nil {
				return mapFetchErr(err)
			}
			if noWikitext {
				art.Wikitext = ""
			}
			return a.render(art)
		},
	}

	cmd.Flags().BoolVar(&noWikitext, "no-wikitext", false, "omit raw wikitext from output")
	return cmd
}
