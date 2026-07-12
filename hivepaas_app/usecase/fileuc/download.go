package fileuc

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

const (
	defaultPresignExpiration = time.Minute * 5
)

func (uc *UC) DownloadFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.DownloadFileReq,
) (_ *filedto.DownloadFileResp, err error) {
	needParseToken := req.Token != "" || auth == nil || auth.User.Role != base.UserRoleAdmin
	if needParseToken {
		tokenClaims, err := uc.fileService.ParseDownloadToken(req.Token)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.ErrTokenInvalid).WithCause(err)
		}
		if tokenClaims.FileID != req.ID {
			return nil, apperrors.Wrap(apperrors.ErrUnauthorized).WithMsgLog("file ID not match")
		}
		if tokenClaims.RequireLogin && (auth == nil || auth.User.ID != tokenClaims.UserID) {
			return nil, apperrors.Wrap(apperrors.ErrUnauthorized).WithMsgLog("user ID not match")
		}
	}

	file, err := uc.fileRepo.GetByID(ctx, uc.db, req.ID,
		bunex.SelectRelation("Storage"),
		bunex.SelectWhere("file.status = ?", base.FileStatusActive),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if !file.IsActive() || file.Deleted {
		return nil, apperrors.NewNotFound("File")
	}

	switch file.StorageType {
	case base.FileStorageLocal:
		return uc.downloadLocalFile(ctx, req, file)
	case base.FileStorageCloud:
		return uc.downloadCloudFile(ctx, req, file)
	default:
		return nil, apperrors.NewUnsupported("Storage type")
	}
}

func (uc *UC) downloadLocalFile(
	_ context.Context,
	req *filedto.DownloadFileReq,
	file *entity.File,
) (_ *filedto.DownloadFileResp, err error) {
	respData := &filedto.DownloadFileDataResp{
		ContentType:   file.Mimetype,
		ContentLength: file.Size,
		ExtraHeaders: map[string]string{
			"Content-Disposition": gofn.If(req.ViewInline, "inline; ", "attachment; ") +
				`filename*=UTF-8''` + url.QueryEscape(file.Name),
		},
	}

	filePath := filepath.Join(config.Current.AppPath, file.Path)
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
	file *entity.File,
) (_ *filedto.DownloadFileResp, err error) {
	if file.Storage == nil {
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

	switch base.CloudStorageKind(file.Storage.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := s3.NewClientFromSetting(ctx, file.Storage)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		objectKey := file.Path
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
