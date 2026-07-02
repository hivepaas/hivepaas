package projectsettingsuc

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) DeleteProjectTags(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.DeleteProjectTagsReq,
) (*projectsettingsdto.DeleteProjectTagsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		tagData := &deleteProjectTagData{}
		err := uc.loadProjectTagDataForDelete(ctx, db, req, tagData)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData := &persistingProjectData{}
		uc.prepareDeletingProjectTag(tagData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &projectsettingsdto.DeleteProjectTagsResp{}, nil
}

type deleteProjectTagData struct {
	Project             *entity.Project
	DeletingProjectTags []*entity.ProjectTag
	UpdatingOrderTags   []*entity.ProjectTag
}

func (uc *UC) loadProjectTagDataForDelete(
	ctx context.Context,
	db database.IDB,
	req *projectsettingsdto.DeleteProjectTagsReq,
	data *deleteProjectTagData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectRelation("Tags", bunex.SelectOrder("display_order")),
	)
	if err != nil {
		return apperrors.New(err)
	}
	data.Project = project

	lowerTags := gofn.MapSlice(req.Tags, strings.ToLower)
	for _, tag := range project.Tags {
		if tag.DeletedAt.IsZero() && gofn.Contain(lowerTags, strings.ToLower(tag.Tag)) {
			data.DeletingProjectTags = append(data.DeletingProjectTags, tag)
		} else {
			data.UpdatingOrderTags = append(data.UpdatingOrderTags, tag)
		}
	}
	if len(data.DeletingProjectTags) != len(req.Tags) {
		return apperrors.NewNotFound("Project tag").
			WithMsgLog("one or more tags not found in project")
	}

	return nil
}

func (uc *UC) prepareDeletingProjectTag(
	tagData *deleteProjectTagData,
	persistingData *persistingProjectData,
) {
	timeNow := timeutil.NowUTC()

	// Deletes the tags
	for _, projectTag := range tagData.DeletingProjectTags {
		projectTag.DeletedAt = timeNow
		persistingData.UpsertingTags = append(persistingData.UpsertingTags, projectTag)
	}

	// Updates order of the active tags
	for i, projectTag := range tagData.UpdatingOrderTags {
		if projectTag.DisplayOrder != i {
			projectTag.DisplayOrder = i
			persistingData.UpsertingTags = append(persistingData.UpsertingTags, projectTag)
		}
	}
}
