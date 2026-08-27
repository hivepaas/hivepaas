package appsettingsuc

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) CreateAppTag(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.CreateAppTagReq,
) (*appsettingsdto.CreateAppTagResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		tagData := &createAppTagData{}
		err := uc.loadAppTagDataForAddNew(ctx, db, req, tagData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.preparePersistingAppTags(tagData.App, []string{req.Tag}, tagData.NextIndex, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.CreateAppTagResp{}, nil
}

type createAppTagData struct {
	App       *entity.App
	NextIndex int
}

func (uc *UC) loadAppTagDataForAddNew(
	ctx context.Context,
	db database.IDB,
	req *appsettingsdto.CreateAppTagReq,
	data *createAppTagData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Tags", bunex.SelectOrder("index")),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app

	nextIndex := 0
	for _, tag := range app.Tags {
		if tag.DeletedAt.IsZero() && strings.EqualFold(tag.Tag, req.Tag) {
			return hperrors.NewAlreadyExist("App tag")
		}
		nextIndex = max(nextIndex, tag.Index+1)
	}
	data.NextIndex = nextIndex

	return nil
}
