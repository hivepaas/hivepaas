// Package managedservice is the engines that look after managed templates' apps.
package managedservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

type Service interface {
	// Engine is the engine of that name, and false when this HivePaaS has none.
	Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool)

	// ResolveS3 is an S3 archive with its credentials, decrypted for the engine:
	// the cloud storage of the backup repository or the one storage names, with
	// storage's own key in place of the storage's when it has one, and its
	// prefix. Settings are read from scope. A volume repository has no
	// credentials, and is refused as ERR_STORAGE_TYPE_UNSUPPORTED.
	ResolveS3(ctx context.Context, db database.IDB, scope *entity.ObjectScope, storage entity.ManagedStorage) (
		*managedmodel.S3Archive, error)
}
