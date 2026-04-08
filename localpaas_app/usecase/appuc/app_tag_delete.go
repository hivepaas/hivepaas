package appuc

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) DeleteAppTags(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.DeleteAppTagsReq,
) (*appdto.DeleteAppTagsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		tagData := &deleteAppTagData{}
		err := uc.loadAppTagDataForDelete(ctx, db, req, tagData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareDeletingAppTag(tagData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.DeleteAppTagsResp{}, nil
}

type deleteAppTagData struct {
	App               *entity.App
	DeletingAppTags   []*entity.AppTag
	UpdatingOrderTags []*entity.AppTag
}

func (uc *UC) loadAppTagDataForDelete(
	ctx context.Context,
	db database.IDB,
	req *appdto.DeleteAppTagsReq,
	data *deleteAppTagData,
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

	lowerTags := gofn.MapSlice(req.Tags, strings.ToLower)
	for _, tag := range app.Tags {
		if tag.DeletedAt.IsZero() && gofn.Contain(lowerTags, strings.ToLower(tag.Tag)) {
			data.DeletingAppTags = append(data.DeletingAppTags, tag)
		} else {
			data.UpdatingOrderTags = append(data.UpdatingOrderTags, tag)
		}
	}
	if len(data.DeletingAppTags) != len(req.Tags) {
		return apperrors.NewNotFound("App tag").
			WithMsgLog("one or more tags not found in app")
	}

	return nil
}

func (uc *UC) prepareDeletingAppTag(
	tagData *deleteAppTagData,
	persistingData *persistingAppData,
) {
	timeNow := timeutil.NowUTC()

	// Deletes the tags
	for _, tag := range tagData.DeletingAppTags {
		tag.DeletedAt = timeNow
		persistingData.UpsertingTags = append(persistingData.UpsertingTags, tag)
	}

	// Updates order of the active tags
	for i, tag := range tagData.UpdatingOrderTags {
		if tag.DisplayOrder != i {
			tag.DisplayOrder = i
			persistingData.UpsertingTags = append(persistingData.UpsertingTags, tag)
		}
	}
}
