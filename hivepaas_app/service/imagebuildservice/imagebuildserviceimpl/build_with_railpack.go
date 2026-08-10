package imagebuildserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) buildImageWithRailpack(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (err error) {
	// TODO: implement later
	return nil
}
