package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
	_, err := transfermanager.New(client.client).UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(defaultContentType),
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (client *Client) UploadEx(
	ctx context.Context,
	bucketName string,
	objectKey string,
	partSizeBytes int64,
	concurrency int,
	data io.Reader,
) error {
	uploader := transfermanager.New(client.client, func(u *transfermanager.Options) {
		if partSizeBytes > 0 {
			u.PartSizeBytes = partSizeBytes
		}
		if concurrency > 0 {
			u.Concurrency = concurrency
		}
	})
	_, err := uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        data,
		ContentType: aws.String(defaultContentType),
	})
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
