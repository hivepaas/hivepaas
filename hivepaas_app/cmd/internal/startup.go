package internal

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/startupservice"
)

func FinalizeStartup(
	startupService startupservice.Service,
) {
	safego.Go("finalizeStartup", func() {
		time.Sleep(5 * time.Second) //nolint:mnd
		startupService.Shutdown()
	})
}
