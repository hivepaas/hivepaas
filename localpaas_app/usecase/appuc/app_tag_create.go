package appuc

import (
	"context"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) CreateAppTag(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.CreateAppTagReq,
) (*appdto.CreateAppTagResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		tagData := &createAppTagData{}
		err := uc.loadAppTagDataForAddNew(ctx, db, req, tagData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.preparePersistingAppTags(tagData.App, []string{req.Tag}, tagData.NextDisplayOrder, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.CreateAppTagResp{}, nil
}

type createAppTagData struct {
	App              *entity.App
	NextDisplayOrder int
}

func (uc *UC) loadAppTagDataForAddNew(
	ctx context.Context,
	db database.IDB,
	req *appdto.CreateAppTagReq,
	data *createAppTagData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("Tags", bunex.SelectOrder("display_order")),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	nextDisplayOrder := 0
	for _, tag := range app.Tags {
		if tag.DeletedAt.IsZero() && strings.EqualFold(tag.Tag, req.Tag) {
			return apperrors.NewAlreadyExist("App tag")
		}
		nextDisplayOrder = max(nextDisplayOrder, tag.DisplayOrder+1)
	}
	data.NextDisplayOrder = nextDisplayOrder

	return nil
}
