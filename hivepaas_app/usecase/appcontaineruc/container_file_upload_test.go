package appcontaineruc

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
)

// A single file never replaces a directory, whatever the client asked: a path
// without its slash names the file, and /app given for the directory would
// put the file where the app's code was.
func TestASingleFileNeverReplacesADirectory(t *testing.T) {
	yes, no := true, false

	assert.False(t, allowDirReplaced(&appcontainerdto.UploadFileToContainerReq{Overwrite: &yes}))
	assert.False(t, allowDirReplaced(&appcontainerdto.UploadFileToContainerReq{}))
	assert.True(t, allowDirReplaced(&appcontainerdto.UploadFileToContainerReq{Extract: true}), "an archive, by default")
	assert.False(t, allowDirReplaced(&appcontainerdto.UploadFileToContainerReq{Extract: true, Overwrite: &no}))
}

// Docker's refusal to put a file where a directory is says what to do instead,
// as a conflict of the client's - not a 500 with docker's words.
func TestADirectoryInTheWayIsToldAsSuch(t *testing.T) {
	refused := hperrors.NewInfra(errors.New(`Error response from daemon: cannot overwrite directory "/app" ` +
		`with non-directory "/app"`))

	err := copyToError(refused, &appcontainerdto.UploadFileToContainerReq{Path: "/app"})

	assert.ErrorIs(t, err, hperrors.ErrContainerPathIsDir)
	info, _ := hperrors.ParseError(err, "")
	if assert.NotNil(t, info) {
		assert.Equal(t, http.StatusConflict, info.Status)
	}

	other := hperrors.NewInfra(errors.New("Error response from daemon: No such container: abc"))
	assert.NotErrorIs(t, copyToError(other, &appcontainerdto.UploadFileToContainerReq{Path: "/app"}),
		hperrors.ErrContainerPathIsDir)
}
