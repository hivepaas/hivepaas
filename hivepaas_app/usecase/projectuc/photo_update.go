package projectuc

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) UpdateProjectPhoto(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectPhotoReq,
) (*projectdto.UpdateProjectPhotoResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		profileData := &updateProjectPhotoData{}
		err := uc.loadProjectPhotoDataForUpdate(ctx, db, req, profileData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectPhotoData{}
		uc.preparePersistingProjectPhoto(req.ProjectPhotoReq, profileData.Project, persistingData)

		return uc.persistProjectPhotoData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.UpdateProjectPhotoResp{}, nil
}

type updateProjectPhotoData struct {
	Project *entity.Project
}

type persistingProjectPhotoData struct {
	UpdatingProject          *entity.Project
	UpsertingBinObjects      []*entity.BinObject
	HardDeletingBinObjectIDs []string
}

func (uc *UC) loadProjectPhotoDataForUpdate(
	ctx context.Context,
	db database.IDB,
	req *projectdto.UpdateProjectPhotoReq,
	data *updateProjectPhotoData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectRelation("PhotoData"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	return nil
}

func (uc *UC) preparePersistingProjectPhoto(
	req *projectdto.ProjectPhotoReq,
	project *entity.Project,
	persistingData *persistingProjectPhotoData,
) {
	if !req.IsChanged() {
		return
	}
	timeNow := timeutil.NowUTC()
	photoData := project.PhotoData

	if req.Delete {
		if photoData != nil && photoData.ID != "" {
			// Project photo may take a remarkable space, so we hard-delete it
			persistingData.HardDeletingBinObjectIDs = append(persistingData.HardDeletingBinObjectIDs, photoData.ID)
		}
		project.Photo = ""
		project.PhotoID = ""
		return
	}

	if photoData == nil {
		photoData = &entity.BinObject{
			ID:        gofn.Must(ulid.NewStringULID()),
			CreatedAt: timeNow,
		}
	}
	photoData.UpdatedAt = timeNow
	photoData.Type = base.BinObjectTypeProjectPhoto
	photoData.Status = base.BinObjectStatusActive
	photoData.Name = req.FileName
	photoData.ContentType = fileutil.TypeByExtension(filepath.Ext(req.FileName))
	photoData.Data = req.DataBytes
	persistingData.UpsertingBinObjects = append(persistingData.UpsertingBinObjects, photoData)

	project.PhotoID = photoData.ID
	project.Photo = fmt.Sprintf("%v/images/%v-%v", config.Current.HTTPServer.BasePath,
		project.PhotoID, rand.Int31n(1000)) //nolint
	project.UpdatedAt = timeNow
	persistingData.UpdatingProject = project
}

func (uc *UC) persistProjectPhotoData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingProjectPhotoData,
) error {
	err := uc.projectRepo.Update(ctx, db, persistingData.UpdatingProject,
		bunex.UpdateColumns("updated_at", "photo", "photo_id"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.binObjectRepo.UpsertMulti(ctx, db, persistingData.UpsertingBinObjects,
		entity.BinObjectUpsertingConflictCols, entity.BinObjectUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.binObjectRepo.DeleteByIDs(ctx, db, persistingData.HardDeletingBinObjectIDs,
		bunex.DeleteWithForceDelete())
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
