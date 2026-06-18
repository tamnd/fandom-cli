package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) revisionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revisions <title>",
		Short: "List revision history for a page",
		Long: `List the revision history of a page using the MediaWiki Action API.

Each record includes the revision ID, editor username, timestamp, edit comment,
and page size after the edit. Use --limit to control how many revisions to fetch.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			n := a.effectiveLimit(50)
			a.progressf("fetching revisions for %q on %s.fandom.com...", title, a.wiki)
			revs, err := a.client.GetRevisions(cmd.Context(), a.wiki, title, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(revs, len(revs))
		},
	}
	return cmd
}
