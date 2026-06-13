package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func (a *App) articleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "article <id>",
		Short: "Show details for a single article by numeric ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return codeError(exitUsage, fmt.Errorf("id must be a number: %w", err))
			}
			a.progressf("fetching article %d from %s.fandom.com...", id, a.wiki)
			art, err := a.client.GetArticle(cmd.Context(), a.wiki, id)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(art)
		},
	}
	return cmd
}
