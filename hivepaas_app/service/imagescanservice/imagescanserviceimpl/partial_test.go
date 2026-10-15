package imagescanserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
)

var (
	imageA = "sha256:" + strings.Repeat("a", 64)
	imageB = "sha256:" + strings.Repeat("b", 64)
)

// nodesOf answers for the nodes it knows the image their container runs,
// and fails for the others: their agents do not answer.
type nodesOf map[string]string

func (n nodesOf) ExecCommand(_ context.Context, req *nodeexecservice.CommandExecReq) (
	*nodeexecservice.CommandExecResp, error) {
	image, ok := n[req.NodeID]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrInfraActionFailed).WithExtraDetail("agent unreachable")
	}
	if req.Command[1] == "inspect" {
		_, _ = fmt.Fprintf(req.Stdout, "%s shop/web:1\n", image)
	}
	return &nodeexecservice.CommandExecResp{}, nil
}

// Some nodes answering, their images are given, and which nodes did not.
func TestImagesOnNodesSaysWhichNodesWereNotRead(t *testing.T) {
	svc := &service{nodeExecService: nodesOf{"A": imageA}}

	images, err := svc.imagesOnNodes(context.Background(), map[string][]string{"A": {"c1"}, "B": {"c2"}},
		tasklog.NewNullStore())

	var unread *unreadNodes
	if assert.True(t, errors.As(err, &unread), "got %v", err) {
		assert.Equal(t, []string{"B"}, unread.nodeIDs)
	}
	if assert.Len(t, images, 1) {
		assert.Equal(t, imageA, images[0].ID)
	}

	// None answering is the error alone, as before.
	images, err = (&service{nodeExecService: nodesOf{}}).imagesOnNodes(context.Background(),
		map[string][]string{"B": {"c2"}}, tasklog.NewNullStore())
	assert.Nil(t, images)
	assert.False(t, errors.As(err, &unread))
	assert.Error(t, err)
}

// appScans holds an app's last scan, finds every image scanned within the
// hour, and keeps what is written.
type appScans struct {
	repository.ImageScanRepo
	previous *entity.AppImageScan
	written  *entity.AppImageScan
}

func (r *appScans) GetAppScan(context.Context, database.IDB, string) (*entity.AppImageScan, error) {
	if r.previous == nil {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return r.previous, nil
}

func (r *appScans) ListScans(_ context.Context, _ database.IDB, imageIDs []string, _ bool,
	_ ...bunex.SelectQueryOption) ([]*entity.ImageScan, error) {
	return []*entity.ImageScan{{ImageID: imageIDs[0], Status: base.ImageScanStatusDone, ScannedAt: time.Now()}}, nil
}

func (r *appScans) UpsertAppScan(_ context.Context, _ database.IDB, appScan *entity.AppImageScan) error {
	r.written = appScan
	return nil
}

// A node that does not answer leaves its image as the last scan found it: kept
// on the app, its findings neither leave the screens nor come back as new the
// next day, and the app says why.
func TestAPartlyReadAppKeepsTheImagesOfTheNodesNotRead(t *testing.T) {
	repo := &appScans{previous: &entity.AppImageScan{AppID: "app", ImageIDs: []string{imageA, imageB},
		Reported: []string{"CVE-1|pkg"}}}
	svc := &service{imageScanRepo: repo}
	unread := &unreadNodes{nodeIDs: []string{"B"}, last: errors.New("agent unreachable")}

	out, err := svc.scanSubject(context.Background(), &subject{id: "app",
		images: []runningImage{{ID: imageA, Ref: "shop/web:1", NodeID: "A"}}, imagesErr: unread},
		&imageScanReq{log: tasklog.NewNullStore()})

	assert.ErrorIs(t, err, hperrors.ErrImageScanFailed)
	assert.Equal(t, []string{imageA, imageB}, repo.written.ImageIDs)
	assert.Equal(t, []string{"CVE-1|pkg"}, repo.written.Reported)
	assert.Contains(t, repo.written.Error, "node B could not be read")
	assert.Equal(t, repo.written.Error, out.Error)
	assert.Len(t, out.Images, 1, "the image read is scanned")
}

// Every node read, the app runs what was read, and nothing of before is kept.
func TestAWhollyReadAppRunsWhatWasRead(t *testing.T) {
	repo := &appScans{previous: &entity.AppImageScan{AppID: "app", ImageIDs: []string{imageB}}}
	svc := &service{imageScanRepo: repo}

	_, err := svc.scanSubject(context.Background(), &subject{id: "app",
		images: []runningImage{{ID: imageA, NodeID: "A"}}}, &imageScanReq{log: tasklog.NewNullStore()})

	assert.NoError(t, err)
	assert.Equal(t, []string{imageA}, repo.written.ImageIDs)
	assert.Empty(t, repo.written.Error)
}

// A scan is started only with the time to end: its task's deadline would cut
// Trivy mid-run once the wait for the volume had eaten the time.
func TestAScanIsStartedOnlyWithTheTimeToEnd(t *testing.T) {
	assert.NoError(t, timeLeftFor(context.Background()), "no deadline")

	ctx, cancel := context.WithTimeout(context.Background(), callTimeout+time.Minute)
	defer cancel()
	assert.NoError(t, timeLeftFor(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := timeLeftFor(ctx)
	assert.ErrorIs(t, err, hperrors.ErrImageScanNotReady)
	assert.Contains(t, hperrors.GetErrorDetail(err, ""), "scanned on the next run")
}
