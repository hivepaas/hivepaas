package fileservice

import (
	"context"
	"net/url"
	"path/filepath"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/appentity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
	"github.com/localpaas/localpaas/services/aws/s3"
)

type GetDownloadURLReq struct {
	File         *entity.Setting
	RequireLogin bool
	Expiration   time.Duration
	CloudPresign bool
	ViewInline   bool
}

type GetDownloadURLResp struct {
	URL string
}

func (s *fileService) GetDownloadURL(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *GetDownloadURLReq,
) (*GetDownloadURLResp, error) {
	if req.File.Type != base.SettingTypeFile {
		return nil, apperrors.NewTypeInvalid()
	}

	file := req.File.MustAsFile()
	if file.StorageType == base.FileStorageLocal || !req.CloudPresign {
		token, err := s.generateFileDownloadToken(auth.User.ID, req)
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
		return &GetDownloadURLResp{URL: urlStr}, nil
	}

	// File is stored in an external cloud and presign allowed
	refObjects, err := s.settingService.LoadReferenceObjectsByIDs(ctx, db, nil, true,
		false, file.GetRefObjectIDs())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	storageSttg := refObjects.RefSettings[file.Storage.ID]
	storage := storageSttg.MustAsCloudStorage()
	providerSttg := refObjects.RefSettings[storage.Provider.ID]
	provider := providerSttg.MustAsCloudProvider()
	storage.RefProvider = provider

	switch {
	case storage.S3 != nil:
		s3Client, err := s3.NewClientFromStorage(ctx, storage)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		urlStr, err := s3Client.PresignGet(ctx, file.Bucket, filepath.Join(file.Path, file.Name), req.ViewInline,
			file.Name, file.Mimetype, req.Expiration)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return &GetDownloadURLResp{URL: urlStr}, nil
	default:
		return nil, apperrors.NewUnsupported("File storage type")
	}
}

func (s *fileService) generateFileDownloadToken(
	userID string,
	req *GetDownloadURLReq,
) (string, error) {
	fileToken, err := jwtsession.GenerateToken(&appentity.FileDownloadTokenClaims{
		UserID:       userID,
		FileID:       req.File.ID,
		RequireLogin: req.RequireLogin,
	}, req.Expiration)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return fileToken, nil
}
