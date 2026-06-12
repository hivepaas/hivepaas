package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc/filedto"
)

func (uc *UC) Upload(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.UploadReq,
) (*filedto.UploadResp, error) {
	uploadData := &fileUploadData{
		baseFileData: &baseFileData{},
	}
	if err := uc.loadUploadData(ctx, uc.db, req, uploadData); err != nil {
		return nil, apperrors.Wrap(err)
	}

	uploadReq := &fileservice.UploadReq{
		FileType:    req.FileType,
		StorageType: req.StorageType,
		StorageID:   req.StorageID,
		Scope:       req.Scope,
		SaveToDB:    true,
	}

	defer func() {
		for _, file := range uploadReq.Items {
			if file.FileData == nil {
				continue
			}
			_ = file.FileData.Close()
		}
	}()

	for _, file := range req.Files {
		f, err := file.Open()
		if err != nil {
			return nil, apperrors.New(err).WithMsgLog("failed to open file: %s", file.Filename)
		}
		uploadReq.Items = append(uploadReq.Items, &fileservice.UploadItemReq{
			FilePath: file.Filename,
			FileSize: file.Size,
			FileData: f,
		})
	}

	uploadResp, err := uc.fileService.Upload(ctx, uc.db, uploadReq)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to upload files")
	}

	resp, err := filedto.TransformFiles(uploadResp.Files)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.UploadResp{
		Data: resp,
	}, nil
}

type fileUploadData struct {
	*baseFileData
}

func (uc *UC) loadUploadData(
	ctx context.Context,
	db database.IDB,
	req *filedto.UploadReq,
	data *fileUploadData,
) (err error) {
	err = uc.loadScopeData(ctx, db, req.Scope, data.baseFileData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
