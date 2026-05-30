package fileserviceimpl

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/fileservice"
	"github.com/localpaas/localpaas/services/aws/s3"
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
	storageSttg := file.Storage

	switch base.CloudStorageKind(storageSttg.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := s3.NewClientFromSetting(ctx, storageSttg)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		urlStr, err := s3Client.PresignGetObject(ctx, file.Bucket, filepath.Join(file.Path, file.Name),
			file.Name, file.Mimetype, req.ViewInline, req.Expiration)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return &fileservice.GetDownloadURLResp{URL: urlStr}, nil
	default:
		return nil, apperrors.NewUnsupported("File storage type")
	}
}
