package syscleanupserviceimpl

import (
	"context"
)

func (s *service) sysCleanupFiles(
	_ context.Context,
	data *sysCleanupData,
) (err error) {
	defer func() {
		if err != nil {
			data.TaskOutput.FileCleanup.Error = err.Error()
		}
	}()

	// TODO: add implementation

	return nil
}
