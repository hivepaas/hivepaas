package projectsettingsuc

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) CreateProjectTag(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.CreateProjectTagReq,
) (*projectsettingsdto.CreateProjectTagResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		tagData := &createProjectTagData{}
		err := uc.loadProjectTagDataForAddNew(ctx, db, req, tagData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.preparePersistingProjectTags(tagData.Project, []string{req.Tag}, tagData.NextIndex, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectsettingsdto.CreateProjectTagResp{}, nil
}

type createProjectTagData struct {
	Project   *entity.Project
	NextIndex int
}

func (uc *UC) loadProjectTagDataForAddNew(
	ctx context.Context,
	db database.IDB,
	req *projectsettingsdto.CreateProjectTagReq,
	data *createProjectTagData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectRelation("Tags", bunex.SelectOrder("index")),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	nextIndex := 0
	for _, projectTag := range project.Tags {
		if projectTag.DeletedAt.IsZero() && strings.EqualFold(projectTag.Tag, req.Tag) {
			return apperrors.NewAlreadyExist("Project tag")
		}
		nextIndex = max(nextIndex, projectTag.Index+1)
	}
	data.NextIndex = nextIndex

	return nil
}

func (uc *UC) preparePersistingProjectTags(
	project *entity.Project,
	tags []string,
	startIndex int,
	persistingData *persistingProjectData,
) {
	index := startIndex
	for _, tag := range tags {
		persistingData.UpsertingTags = append(persistingData.UpsertingTags,
			&entity.Tag{
				ObjectID: project.ID,
				Tag:      tag,
				Index:    index,
			})
		index++
	}
}
