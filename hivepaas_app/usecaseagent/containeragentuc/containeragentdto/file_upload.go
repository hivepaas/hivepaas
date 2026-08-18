package containeragentdto

import "io"

// UploadFileInput represents input for copying/uploading a TAR stream to a container
type UploadFileInput struct {
	ContainerID string
	DstPath     string
	TarReader   io.Reader
	Overwrite   bool
	CopyUIDGID  bool
}
