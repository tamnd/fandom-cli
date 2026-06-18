package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) siteInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "siteinfo",
		Short: "Show wiki statistics (articles, pages, edits, images, users)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.progressf("fetching site info from %s.fandom.com...", a.wiki)
			info, err := a.client.GetSiteInfo(cmd.Context(), a.wiki)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(info)
		},
	}
	return cmd
}
