package fileserviceimpl

import (
	"context"
	"net/url"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

func (s *service) GetDownloadURL(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *fileservice.GetDownloadURLReq,
) (*fileservice.GetDownloadURLResp, error) {
	file := req.File
	if file.StorageType == base.FileStorageLocal || !req.CloudPresign {
		token, err := s.GenerateDownloadToken(auth.User.ID, req.File.ID, req.RequireLogin, req.Expiration)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		urlStr, err := url.JoinPath(config.Current.BaseAPIURL(), "files", req.File.ID, "download")
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		urlStr += "?token=" + token
		if req.ViewInline {
			urlStr += "&viewInline=true"
		}
		return &fileservice.GetDownloadURLResp{URL: urlStr}, nil
	}

	// File is stored in an external cloud and presign allowed
	storageSetting := file.Storage

	switch base.CloudStorageKind(storageSetting.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := s3.NewClientFromSetting(ctx, storageSetting)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		urlStr, err := s3Client.PresignGetObject(ctx, file.Bucket, file.Path,
			file.Name, file.Mimetype, req.ViewInline, req.Expiration)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return &fileservice.GetDownloadURLResp{URL: urlStr}, nil
	default:
		return nil, apperrors.NewUnsupported("File storage type")
	}
}
