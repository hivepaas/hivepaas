package base

type CloudStorageKind string

const (
	CloudStorageKindS3 CloudStorageKind = "aws-s3"
)

var (
	AllCloudStorageKinds = []CloudStorageKind{CloudStorageKindS3}
)
