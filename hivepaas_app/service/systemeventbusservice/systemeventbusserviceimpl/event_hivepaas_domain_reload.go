package systemeventbusserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

//nolint:unparam
func (s *service) onHivepaasDomainReload() error {
	config.SetAppDomainToNeedReload()
	return nil
}
