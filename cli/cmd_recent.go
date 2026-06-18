package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) recentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recent",
		Short: "Stream recent changes (edits, new pages, log actions)",
		Long: `Stream recent changes using the MediaWiki recentchanges API.

Each record includes the change type (edit/new/log/categorize), page title,
page ID, revision ID, editor username, timestamp, and edit comment.

This is richer than the "activity" command: it includes new page creations,
log events, and categorize actions in addition to edits.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(50)
			a.progressf("fetching recent changes from %s.fandom.com...", a.wiki)
			changes, err := a.client.RecentChanges(cmd.Context(), a.wiki, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(changes, len(changes))
		},
	}
	return cmd
}
