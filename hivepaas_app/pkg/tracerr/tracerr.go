package tracerr

import (
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/errstack"
)

// Wrap wraps an error with adding stack trace
func Wrap(err error, msg ...string) error {
	if err == nil {
		return nil
	}
	err = errstack.Wrap(err, 1)
	if len(msg) == 0 {
		return err
	}
	return fmt.Errorf("%s: %w", msg[0], err)
}
