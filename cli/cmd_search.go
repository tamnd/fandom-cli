package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search articles in a Fandom wiki",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(20)
			a.progressf("searching %s.fandom.com for %q...", a.wiki, args[0])
			items, err := a.client.Search(cmd.Context(), a.wiki, args[0], n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(items, len(items))
		},
	}
	return cmd
}
