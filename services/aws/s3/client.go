package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type Client struct {
	Config        *Config
	client        *s3.Client
	presignClient *s3.PresignClient
}

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	Region          string
	Bucket          string
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithDefaultRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(opts *s3.Options) {
		if cfg.Region != "" {
			opts.Region = cfg.Region
		}
		if cfg.Endpoint != "" {
			opts.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &Client{
		Config:        cfg,
		client:        s3Client,
		presignClient: s3.NewPresignClient(s3Client),
	}, nil
}

func NewClientFromSetting(ctx context.Context, storageSetting *entity.Setting) (*Client, error) {
	if storageSetting.Type != base.SettingTypeCloudStorage || storageSetting.Kind != string(base.CloudStorageKindS3) {
		return nil, apperrors.New(apperrors.ErrSettingTypeUnsupported).WithParam("Name", storageSetting.Type)
	}
	storage, err := storageSetting.AsCloudStorage()
	if err != nil {
		return nil, apperrors.New(err)
	}
	secretKey, err := storage.S3.SecretKey.GetPlain()
	if err != nil {
		return nil, apperrors.New(err)
	}
	return NewClient(ctx, &Config{
		AccessKeyID:     storage.S3.AccessKeyID,
		SecretAccessKey: secretKey,
		Endpoint:        storage.S3.Endpoint,
		Region:          gofn.Coalesce(storage.S3.Region, storage.S3.CloudProviderAWS.Region),
		Bucket:          storage.S3.Bucket,
	})
}

func (client *Client) HeadBucket(ctx context.Context) (*s3.HeadBucketOutput, error) {
	result, err := client.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(client.Config.Bucket),
	})
	if err != nil {
		return nil, apperrors.New(err)
	}
	return result, nil
}
