package cloudstorageuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/cloudstorageservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

func (uc *UC) TestCloudStorageConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.TestCloudStorageConnReq,
) (*cloudstoragedto.TestCloudStorageConnResp, error) {
	switch req.Kind {
	case base.CloudStorageKindS3:
		return uc.testCloudStorageS3Conn(ctx, auth, req)
	default:
		return nil, hperrors.NewUnsupported("Storage kind")
	}
}

func (uc *UC) testCloudStorageS3Conn(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.TestCloudStorageConnReq,
) (*cloudstoragedto.TestCloudStorageConnResp, error) {
	storage := req.ToEntity()
	keyAuthSetting, err := uc.SettingRepo.GetByID(ctx, uc.DB, nil, base.SettingTypeKeyAuth,
		storage.S3.KeyAuth.ID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The endpoint takes any signed-in caller: the key it tries is one they may
	// read, or it is not tried.
	if err = uc.CheckAccessOnSetting(ctx, uc.DB, auth, keyAuthSetting, base.ActionTypeRead); err != nil {
		return nil, hperrors.Wrap(err)
	}
	keyAuth, err := keyAuthSetting.AsKeyAuth()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	secretKey, err := keyAuth.SecretKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	s3Client, err := s3.NewClient(ctx, &s3.Config{
		AccessKeyID:     keyAuth.KeyID,
		SecretAccessKey: secretKey,
		Region:          storage.S3.Region,
		Endpoint:        storage.S3.Endpoint,
		Bucket:          storage.S3.Bucket,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	_, err = s3Client.HeadBucket(ctx)
	if err != nil {
		return nil, hperrors.Wrap(cloudstorageservice.ConnError(err))
	}

	return &cloudstoragedto.TestCloudStorageConnResp{}, nil
}
