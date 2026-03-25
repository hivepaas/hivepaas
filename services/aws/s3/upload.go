package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

const (
	defaultContentType = "application/octet-stream"
)

func (client *Client) Upload(
	ctx context.Context,
	bucketName string,
	objectKey string,
	data []byte,
) error {
	_, err := manager.NewUploader(client.client).Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(defaultContentType),
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (client *Client) UploadEx(
	ctx context.Context,
	bucketName string,
	objectKey string,
	partSize int64,
	concurrency int,
	data io.Reader,
) error {
	uploader := manager.NewUploader(client.client, func(u *manager.Uploader) {
		if partSize > 0 {
			u.PartSize = partSize
		}
		if concurrency > 0 {
			u.Concurrency = concurrency
		}
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        data,
		ContentType: aws.String(defaultContentType),
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
