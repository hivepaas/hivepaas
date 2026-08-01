package appuc

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/assets"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppPhoto(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppPhotoReq,
) (*appdto.UpdateAppPhotoResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		appData := &updateAppPhotoData{}
		err := uc.loadAppPhotoDataForUpdate(ctx, db, req, appData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppPhotoData{}
		err = uc.preparePersistingAppPhoto(ctx, db, req.AppPhotoReq, appData.App, appData, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return uc.persistAppPhotoData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppPhotoResp{}, nil
}

type updateAppPhotoData struct {
	App        *entity.App
	PresetIcon string
}

type persistingAppPhotoData struct {
	UpdatingApp              *entity.App
	UpsertingBinObjects      []*entity.BinObject
	HardDeletingBinObjectIDs []string
}

func (uc *UC) loadAppPhotoDataForUpdate(
	ctx context.Context,
	db database.IDB,
	req *appdto.UpdateAppPhotoReq,
	data *updateAppPhotoData,
) error {
	app, err := uc.appRepo.GetByID(ctx, db, req.ProjectID, req.AppID,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("PhotoData"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	if req.IsPresetIcon {
		data.PresetIcon, err = uc.parseAndVerifyPresetIcon(req.FileName)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *UC) parseAndVerifyPresetIcon(fileName string) (string, error) {
	presetIcon := filepath.Base(fileName)
	if filepath.Ext(presetIcon) == "" {
		presetIcon += ".svg"
	}
	stat, err := fs.Stat(assets.GetIconsFS(), presetIcon)
	if err != nil || stat.IsDir() {
		return "", apperrors.NewNotFound("Preset icon")
	}
	return presetIcon, nil
}

func (uc *UC) preparePersistingAppPhoto(
	ctx context.Context,
	db database.IDB,
	req *appdto.AppPhotoReq,
	app *entity.App,
	data *updateAppPhotoData,
	persistingData *persistingAppPhotoData,
) error {
	if !req.IsChanged() {
		return nil
	}
	timeNow := timeutil.NowUTC()
	photoData := app.PhotoData

	if photoData != nil && photoData.ID != "" {
		// App photo may take a remarkable space, so we hard-delete it if unused
		apps, _, err := uc.appRepo.List(ctx, db, "", nil,
			bunex.SelectWhere("app.photo = ?", photoData.ID),
			bunex.SelectWhere("app.id != ?", app.ID),
			bunex.SelectLimit(1),
			bunex.SelectColumns("id"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if len(apps) == 0 {
			persistingData.HardDeletingBinObjectIDs = append(persistingData.HardDeletingBinObjectIDs, photoData.ID)
		}
	}

	switch {
	case req.Delete:
		app.Photo = ""
	case req.IsPresetIcon:
		app.Photo = filepath.Join(config.Current.HttpPathStaticIcons(), data.PresetIcon)
	default:
		photoData = &entity.BinObject{
			ID:          gofn.Must(ulid.NewStringULID()),
			Type:        base.BinObjectTypeObjectIcon,
			Status:      base.BinObjectStatusActive,
			Name:        req.FileName,
			ContentType: fileutil.TypeByExtension(filepath.Ext(req.FileName)),
			Data:        req.DataBytes,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		persistingData.UpsertingBinObjects = append(persistingData.UpsertingBinObjects, photoData)
		app.Photo = photoData.ID
	}

	app.UpdatedAt = timeNow
	persistingData.UpdatingApp = app
	return nil
}

func (uc *UC) persistAppPhotoData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAppPhotoData,
) error {
	err := uc.appRepo.Update(ctx, db, persistingData.UpdatingApp,
		bunex.UpdateColumns("updated_at", "photo"),
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
