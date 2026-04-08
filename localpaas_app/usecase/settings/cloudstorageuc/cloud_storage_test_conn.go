package cloudstorageuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/localpaas/localpaas/services/aws/s3"
)

func (uc *UC) TestCloudStorageConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *cloudstoragedto.TestCloudStorageConnReq,
) (*cloudstoragedto.TestCloudStorageConnResp, error) {
	switch req.Kind {
	case base.CloudStorageKindS3:
		return uc.testCloudStorageS3Conn(ctx, req)
	default:
		return nil, apperrors.NewUnsupported("Storage kind")
	}
}

func (uc *UC) testCloudStorageS3Conn(
	ctx context.Context,
	req *cloudstoragedto.TestCloudStorageConnReq,
) (*cloudstoragedto.TestCloudStorageConnResp, error) {
	storage := req.ToEntity()
	s3Client, err := s3.NewClient(ctx, &s3.Config{
		AccessKeyID:     storage.S3.AccessKeyID,
		SecretAccessKey: storage.S3.SecretKey.MustGetPlain(),
		Region:          gofn.Coalesce(req.S3.Region, storage.S3.CloudProviderAWS.Region),
		Endpoint:        req.S3.Endpoint,
		Bucket:          req.S3.Bucket,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	_, err = s3Client.HeadBucket(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &cloudstoragedto.TestCloudStorageConnResp{}, nil
}
