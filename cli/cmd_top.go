package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) topCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: "List popular articles on a Fandom wiki",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(25)
			a.progressf("fetching top articles from %s.fandom.com...", a.wiki)
			items, err := a.client.Top(cmd.Context(), a.wiki, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(items, len(items))
		},
	}
	return cmd
}
