package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		return nil, hperrors.Wrap(err)
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

func (client *Client) HeadBucket(ctx context.Context) (*s3.HeadBucketOutput, error) {
	result, err := client.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(client.Config.Bucket),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return result, nil
}
