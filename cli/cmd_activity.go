package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) activityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activity",
		Short: "Show recent edit activity on a Fandom wiki",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching recent activity from %s.fandom.com...", a.wiki)
			items, err := a.client.Activity(cmd.Context(), a.wiki, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(items, len(items))
		},
	}
	return cmd
}
