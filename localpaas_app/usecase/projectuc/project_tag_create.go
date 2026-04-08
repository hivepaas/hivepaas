package projectuc

import (
	"context"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) CreateProjectTag(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.CreateProjectTagReq,
) (*projectdto.CreateProjectTagResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		tagData := &createProjectTagData{}
		err := uc.loadProjectTagDataForAddNew(ctx, db, req, tagData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.preparePersistingProjectTags(tagData.Project, []string{req.Tag},
			tagData.NextDisplayOrder, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.CreateProjectTagResp{}, nil
}

type createProjectTagData struct {
	Project          *entity.Project
	NextDisplayOrder int
}

func (uc *UC) loadProjectTagDataForAddNew(
	ctx context.Context,
	db database.IDB,
	req *projectdto.CreateProjectTagReq,
	data *createProjectTagData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectRelation("Tags", bunex.SelectOrder("display_order")),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	nextDisplayOrder := 0
	for _, projectTag := range project.Tags {
		if projectTag.DeletedAt.IsZero() && strings.EqualFold(projectTag.Tag, req.Tag) {
			return apperrors.NewAlreadyExist("Project tag")
		}
		nextDisplayOrder = max(nextDisplayOrder, projectTag.DisplayOrder+1)
	}
	data.NextDisplayOrder = nextDisplayOrder

	return nil
}
