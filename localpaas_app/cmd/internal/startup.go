package internal

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/service/startupservice"
)

func FinalizeStartup(
	startupService startupservice.Service,
) {
	go func() {
		time.Sleep(5 * time.Second) //nolint:mnd
		startupService.Shutdown()
	}()
}
