package imagebuildagentuc

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/srcpack"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

const (
	// sourceDirMode keeps a build's source, which may hold secrets, to the agent.
	sourceDirMode = 0o700
	// sourceCheckoutDir is the source's directory inside the build's own.
	sourceCheckoutDir = "checkout"
)

// ImageBuildFromSource builds an image from a source sent by the app, which
// checked it out on another node. The source is unpacked in a directory of this
// build's own under the day's temporary directory, and removed when the build
// ends, however it ends.
func (uc *UC) ImageBuildFromSource(
	ctx context.Context,
	req *imagebuildagentdto.ImageBuildReq,
	source io.Reader,
) (resp *imagebuildagentdto.ImageBuildResp, err error) {
	// The agent's stream calls have no recovery of their own: a panic here would
	// take the agent down with every other build running on it.
	defer func() {
		if r := recover(); r != nil {
			resp, err = nil, hperrors.NewPanic(r)
		}
	}()

	tempDir, err := fileutil.CreateTempDir(uc.tempBaseDir, "*", sourceDirMode)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer os.RemoveAll(tempDir)

	checkoutDir := filepath.Join(tempDir, sourceCheckoutDir)
	if err = os.Mkdir(checkoutDir, sourceDirMode); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if _, err = srcpack.Unpack(ctx, source, checkoutDir); err != nil {
		return nil, hperrors.Wrap(err)
	}

	req.CheckoutDir = checkoutDir
	req.TempDir = tempDir
	return uc.ImageBuild(ctx, req)
}
