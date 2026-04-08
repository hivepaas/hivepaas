package fileuc

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc/filedto"
	"github.com/localpaas/localpaas/services/aws/s3"
)

const (
	defaultPresignExpiration = time.Minute * 5
)

func (uc *UC) DownloadFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.DownloadFileReq,
) (_ *filedto.DownloadFileResp, err error) {
	req.Type = currentSettingType

	needParseToken := req.Token != "" || auth == nil || auth.User.Role != base.UserRoleAdmin
	if needParseToken {
		tokenClaims, err := uc.FileService.ParseFileDownloadToken(req.Token)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
		}
		if tokenClaims.FileID != req.ID {
			return nil, apperrors.New(apperrors.ErrUnauthorized).WithMsgLog("file ID not match")
		}
		if tokenClaims.RequireLogin && (auth == nil || auth.User.ID != tokenClaims.UserID) {
			return nil, apperrors.New(apperrors.ErrUnauthorized).WithMsgLog("user ID not match")
		}
	}

	setting, err := uc.SettingRepo.GetByID(ctx, uc.DB, nil, req.Type, req.ID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	file := setting.MustAsFile()
	if !setting.IsActive() || file.Deleted {
		return nil, apperrors.NewNotFound("File")
	}

	switch file.StorageType {
	case base.FileStorageLocal:
		return uc.downloadLocalFile(ctx, req, setting)
	case base.FileStorageCloud:
		return uc.downloadCloudFile(ctx, req, setting)
	default:
		return nil, apperrors.NewUnsupported("Storage type")
	}
}

func (uc *UC) downloadLocalFile(
	_ context.Context,
	req *filedto.DownloadFileReq,
	setting *entity.Setting,
) (_ *filedto.DownloadFileResp, err error) {
	file := setting.MustAsFile()
	respData := &filedto.DownloadFileDataResp{
		ContentType:   file.Mimetype,
		ContentLength: file.Size,
		ExtraHeaders: map[string]string{
			"Content-Disposition": gofn.If(req.ViewInline, "inline; ", "attachment; ") +
				`filename*=UTF-8''` + url.QueryEscape(file.Name),
		},
	}

	filePath := filepath.Join(config.Current.AppPath, file.Path, file.Name)
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	defer func() {
		if err != nil {
			_ = reader.Close()
		}
	}()
	respData.Content = reader
	return &filedto.DownloadFileResp{Data: respData}, nil
}

func (uc *UC) downloadCloudFile(
	ctx context.Context,
	req *filedto.DownloadFileReq,
	setting *entity.Setting,
) (_ *filedto.DownloadFileResp, err error) {
	file := setting.MustAsFile()

	refObjects, err := uc.SettingService.LoadReferenceObjects(ctx, uc.DB, nil,
		true, false, setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	storageSttg := refObjects.RefSettings[file.Storage.ID]
	if storageSttg == nil {
		return nil, apperrors.NewInactive("Storage setting")
	}

	respData := &filedto.DownloadFileDataResp{}
	usePresignURL := req.UsePresignURLOnFileSize > 0 && file.Size > req.UsePresignURLOnFileSize

	if !usePresignURL {
		respData.ExtraHeaders = map[string]string{
			"Content-Disposition": gofn.If(req.ViewInline, "inline; ", "attachment; ") +
				`filename*=UTF-8''` + url.QueryEscape(file.Name),
		}
	}

	switch base.CloudStorageKind(storageSttg.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := s3.NewClientFromSetting(ctx, storageSttg)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		objectKey := filepath.Join(file.Path, file.Name)
		if !usePresignURL {
			s3Object, err := s3Client.GetObject(ctx, file.Bucket, objectKey)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			defer func() {
				if err != nil {
					_ = s3Object.Body.Close()
				}
			}()

			if s3Object.ContentType != nil {
				respData.ContentType = *s3Object.ContentType
			}
			if s3Object.ContentLength != nil {
				respData.ContentLength = *s3Object.ContentLength
			}
			respData.Content = s3Object.Body
			return &filedto.DownloadFileResp{Data: respData}, nil
		}

		expiration := gofn.Coalesce(req.PresignExpiration.ToDuration(), defaultPresignExpiration)
		presignURL, err := s3Client.PresignGetObject(ctx, file.Bucket, objectKey, file.Name, file.Mimetype,
			req.ViewInline, expiration)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		respData.RedirectURL = presignURL
		return &filedto.DownloadFileResp{Data: respData}, nil

	default:
		return nil, apperrors.NewUnsupported("Storage type")
	}
}
