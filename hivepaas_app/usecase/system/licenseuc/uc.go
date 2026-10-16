// Package licenseuc serves System > License: the license in force with how much
// of each limit is used, and installing a license key.
//
// It records its own audit entry rather than going through usecase/settings.
// Only licenseservice writes the license setting, and the settings DTOs would put
// updateVer, inheritable and default on the wire, none of which a license has.
package licenseuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

type UC struct {
	// db is the interface rather than *database.DB, so a test can run the
	// transaction without a database.
	db             database.IDB
	licenseService licenseservice.Service
	auditService   auditservice.Service
}

func New(
	db *database.DB,
	licenseService licenseservice.Service,
	auditService auditservice.Service,
) *UC {
	return &UC{
		db:             db,
		licenseService: licenseService,
		auditService:   auditService,
	}
}
