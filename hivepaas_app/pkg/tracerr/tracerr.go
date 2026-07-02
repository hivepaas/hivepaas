package tracerr

import (
	"fmt"

	goerrors "github.com/go-errors/errors"
)

// Wrap wraps an error with adding stack trace
func Wrap(err error, msg ...string) error {
	if err == nil {
		return nil
	}
	err = goerrors.Wrap(err, 1)
	if len(msg) == 0 {
		return err
	}
	return fmt.Errorf("%s: %w", msg[0], err)
}
