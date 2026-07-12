package fileuc

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

func (uc *UC) CreateFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.CreateFileReq,
) (*filedto.CreateFileResp, error) {
	createData := &fileCreateData{
		baseFileData: &baseFileData{},
	}
	if err := uc.loadCreateFileData(ctx, uc.db, req, createData); err != nil {
		return nil, apperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	fileName := gofn.LastOr(strings.Split(req.FilePath, "/"), "")
	mimetype := fileutil.TypeByExtension(filepath.Ext(fileName))
	var fileSize int64
	storageSetting, err := uc.settingRepo.GetByID(ctx, uc.db, req.Scope, base.SettingTypeCloudStorage,
		req.StorageID, true)
	if err != nil {
		return nil, apperrors.Wrap(err).WithMsgLog("failed to get storage setting")
	}
	s3Client, err := s3.NewClientFromSetting(ctx, storageSetting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	s3Object, err := s3Client.HeadObject(ctx, req.Bucket, req.FilePath)
	if err != nil {
		return nil, apperrors.Wrap(err).WithMsgLog("failed to head object from S3")
	}
	if s3Object.ContentLength != nil {
		fileSize = *s3Object.ContentLength
	}

	file := &entity.File{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       req.Scope.ScopeType(),
		ObjectID:    req.Scope.MainObjectID(),
		Type:        req.FileType,
		Kind:        string(req.FileKind),
		Status:      base.FileStatusActive,
		Name:        fileName,
		Path:        req.FilePath,
		Size:        fileSize,
		Mimetype:    gofn.Coalesce(mimetype, "application/octet-stream"),
		StorageType: base.FileStorageCloud,
		StorageID:   req.StorageID,
		Bucket:      req.Bucket,
		UpdateVer:   1,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	err = uc.fileRepo.Insert(ctx, uc.db, file)
	if err != nil {
		return nil, apperrors.Wrap(err).WithMsgLog("failed to insert file")
	}

	return &filedto.CreateFileResp{
		Data: &basedto.ObjectIDResp{ID: file.ID},
	}, nil
}

type fileCreateData struct {
	*baseFileData
}

func (uc *UC) loadCreateFileData(
	ctx context.Context,
	db database.IDB,
	req *filedto.CreateFileReq,
	data *fileCreateData,
) (err error) {
	err = uc.loadScopeData(ctx, db, req.Scope, data.baseFileData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
