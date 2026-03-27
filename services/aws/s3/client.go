package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type Client struct {
	config        *Config
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
		return nil, apperrors.Wrap(err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(opts *s3.Options) {
		if cfg.Endpoint != "" {
			opts.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &Client{
		config:        cfg,
		client:        s3Client,
		presignClient: s3.NewPresignClient(s3Client),
	}, nil
}

func NewClientFromStorage(ctx context.Context, storage *entity.CloudStorage) (*Client, error) {
	return NewClient(ctx, &Config{
		AccessKeyID:     storage.RefProvider.AWS.AccessKeyID,
		SecretAccessKey: storage.RefProvider.AWS.SecretKey.MustGetPlain(),
		Endpoint:        storage.S3.Endpoint,
		Region:          gofn.Coalesce(storage.S3.Region, storage.RefProvider.AWS.Region),
		Bucket:          storage.S3.Bucket,
	})
}

func (client *Client) HeadBucket(ctx context.Context) (*s3.HeadBucketOutput, error) {
	result, err := client.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(client.config.Bucket),
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return result, nil
}
