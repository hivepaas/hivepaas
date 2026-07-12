package s3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (client *Client) DeleteObject(ctx context.Context, bucketName string, objectKey string) error {
	_, err := client.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		// https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html
		// TODO: needs more investigation here as NoSuchKey never be raised although
		// this code is similar to the official sample in the above page.
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return apperrors.Wrap(apperrors.ErrNotFound).
				WithMsgLog("object %s does not exist in bucket %s", objectKey, bucketName)
		}
		return apperrors.Wrap(err)
	}
	return nil
}
