package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) listCmd() *cobra.Command {
	var offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List articles on a Fandom wiki",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(25)
			a.progressf("listing articles from %s.fandom.com (offset %d)...", a.wiki, offset)
			items, err := a.client.List(cmd.Context(), a.wiki, n, offset)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(items, len(items))
		},
	}
	cmd.Flags().IntVar(&offset, "offset", 0, "pagination offset")
	return cmd
}
