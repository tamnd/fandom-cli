package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show metadata for a Fandom wiki",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.progressf("fetching wiki info for %s.fandom.com...", a.wiki)
			info, err := a.client.Info(cmd.Context(), a.wiki)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(info)
		},
	}
	return cmd
}
