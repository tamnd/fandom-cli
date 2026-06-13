package cli

import (
	"errors"

	"github.com/tamnd/fandom-cli/fandom"
)

func isNotFound(err error) bool {
	return errors.Is(err, fandom.ErrNotFound)
}
