package functiontest

import (
	"context"
	"strings"
	"sync"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functionbuild"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// LibrariesLogMax is how much of an install's log comes back: its end, where
// an error is.
const LibrariesLogMax = 64 * unit.KB

// NewLibrariesBuilder builds libraries images through the image build service,
// on this node, as local images.
func NewLibrariesBuilder(builds imagebuildservice.Service) LibrariesBuilder {
	return &imageBuildLibraries{builds: builds}
}

type imageBuildLibraries struct {
	builds imagebuildservice.Service
}

func (b *imageBuildLibraries) BuildLibraries(ctx context.Context, req *LibrariesBuildReq) (string, error) {
	log := &tailLog{max: int(LibrariesLogMax)}
	logStore := tasklog.NewForwardStore("function-test:libraries",
		func(_ context.Context, frames []*tasklog.LogFrame) error {
			for _, frame := range frames {
				log.add(frame.Data)
			}
			return nil
		})
	_, err := b.builds.ImageBuild(ctx, nil, &imagebuildservice.ImageBuildReq{
		TaskExecData: &queue.TaskExecData{
			Task:     &entity.Task{ID: gofn.Must(ulid.NewStringULID())},
			LogStore: logStore,
		},
		Dockerfile: entity.DeploymentDockerfile{
			Source:  base.DockerfileSourceManual,
			Path:    functionbuild.DockerfilePath,
			Content: req.Dockerfile,
		},
		LocalImage:         req.Image,
		ImageBuildSettings: req.BuildSettings,
		CheckoutDir:        req.ContextDir,
		TempDir:            req.TempDir,
		Inputs:             req.Inputs,
	})
	return log.String(), hperrors.Wrap(err)
}

// tailLog keeps the last max bytes of the lines it is given.
type tailLog struct {
	mu  sync.Mutex
	max int
	buf strings.Builder
}

func (l *tailLog) add(line string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf.WriteString(line + "\n")
	if l.buf.Len() > 2*l.max {
		kept := l.buf.String()[l.buf.Len()-l.max:]
		l.buf.Reset()
		l.buf.WriteString(kept)
	}
}

func (l *tailLog) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := l.buf.String()
	if len(out) > l.max {
		out = out[len(out)-l.max:]
	}
	return out
}
