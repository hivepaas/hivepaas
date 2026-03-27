package s3

import (
	"context"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func (client *Client) GetObject(
	ctx context.Context,
	bucketName string,
	objectKey string,
) ([]byte, error) {
	result, err := client.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return data, nil
}

func (client *Client) PresignGet(
	ctx context.Context,
	bucketName string,
	objectKey string,
	viewInline bool,
	fileName string,
	presignExp time.Duration,
) (string, error) {
	objectInput := &s3.GetObjectInput{
		Bucket:                     aws.String(bucketName),
		Key:                        aws.String(objectKey),
		ResponseContentDisposition: aws.String("attachment; filename=" + fileName),
	}
	if viewInline {
		fileExt := filepath.Ext(objectKey)
		if fileExt != "" {
			objectInput.ResponseContentType = aws.String(mime.TypeByExtension(fileExt))
			objectInput.ResponseContentDisposition = aws.String("inline; filename=" + fileName)
		}
	}
	request, err := client.presignClient.PresignGetObject(ctx, objectInput, func(opts *s3.PresignOptions) {
		opts.Expires = presignExp
	})
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return request.URL, nil
}
