package s3

import (
	"context"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (client *Client) GetObject(
	ctx context.Context,
	bucketName string,
	objectKey string,
) (*s3.GetObjectOutput, error) {
	result, err := client.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, apperrors.New(err)
	}
	return result, nil
}

func (client *Client) PresignGetObject(
	ctx context.Context,
	bucketName string,
	objectKey string,
	fileName string,
	mimetype string,
	viewInline bool,
	expiration time.Duration,
) (string, error) {
	if mimetype == "" {
		mimetype = mime.TypeByExtension(filepath.Ext(fileName))
	}
	objectInput := &s3.GetObjectInput{
		Bucket:              aws.String(bucketName),
		Key:                 aws.String(objectKey),
		ResponseContentType: aws.String(mimetype),
		ResponseContentDisposition: aws.String(gofn.If(viewInline, "inline; ", "attachment; ") +
			`filename*=UTF-8''` + url.QueryEscape(fileName)),
	}

	request, err := client.presignClient.PresignGetObject(ctx, objectInput, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", apperrors.New(err)
	}
	return request.URL, nil
}
