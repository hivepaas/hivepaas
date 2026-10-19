package imagescanserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	amd64SHA = "965a7147852435bc4b54f272799b63df826b54ef65cf7c4fc688c82b96c0c4eb"
	arm64SHA = "1111111111111111111111111111111111111111111111111111111111111111"
)

func walgArtifacts() *entity.AppArtifactsSettings {
	return &entity.AppArtifactsSettings{Artifacts: []*entity.AppArtifact{{
		Name: "wal-g", Version: "v3.0.8", MountPath: "/opt/hivepaas/bin/wal-g",
		Files: map[base.CPUArch]*entity.AppArtifactFile{
			base.CPUArchAMD64: {URL: "https://example.com/wal-g-amd64", SHA256: amd64SHA},
			base.CPUArchARM64: {URL: "https://example.com/wal-g-arm64", SHA256: arm64SHA},
		},
	}}}
}

// An app's artifacts are scanned as the files its tasks bind: the build for
// each node's architecture, once, by its sha256, named by the artifact; none on
// a node whose architecture is not known.
func TestArtifactTargets(t *testing.T) {
	images := []runningImage{
		{ID: imageID, Ref: "postgres:18.6-trixie", NodeID: "n1"},
		{ID: indexDigest, Ref: "postgres:18.6-trixie", NodeID: "n2"},
		{ID: imageID, Ref: "postgres:18.6-trixie", NodeID: "n3"},
		{ID: imageID, Ref: "postgres:18.6-trixie", NodeID: "n4"},
	}
	archOf := map[string]base.CPUArch{"n1": base.CPUArchAMD64, "n2": base.CPUArchARM64, "n3": base.CPUArchAMD64}

	targets := artifactTargets(walgArtifacts(), images, archOf)

	if !assert.Len(t, targets, 2) {
		return
	}
	assert.Equal(t, "sha256:"+amd64SHA, targets[0].ID)
	assert.Equal(t, "wal-g v3.0.8", targets[0].Ref)
	assert.Equal(t, "n1", targets[0].NodeID)
	assert.Equal(t, "linux/amd64", targets[0].Platform)
	assert.Equal(t, base.CPUArchAMD64, targets[0].Artifact.arch)
	assert.Equal(t, "sha256:"+arm64SHA, targets[1].ID)
	assert.Equal(t, "n2", targets[1].NodeID)
	assert.Empty(t, artifactTargets(nil, images, archOf), "an app without artifacts")
}

// An artifact's file is scanned from the artifacts cache, as a root
// filesystem: no registry, no credentials.
func TestAnArtifactsRunReq(t *testing.T) {
	cfg := &entity.ImageScanSettings{DBRepository: "mirror.local:5000/aquasec/trivy-db:2"}
	volume := volumeOf(&entity.ClusterVolume{NodeID: "n1"})
	target := artifactTargets(walgArtifacts(), []runningImage{{ID: imageID, NodeID: "n1"}},
		map[string]base.CPUArch{"n1": base.CPUArchAMD64})[0]

	req := runReqOf(target, cfg, volume, []byte(`{"auths":{}}`), "")

	assert.Equal(t, imagescanrun.ModeRootfs, req.Mode)
	assert.Equal(t, "sha256:"+amd64SHA, req.ImageID)
	assert.Empty(t, req.RemoteRef)
	assert.Empty(t, req.Platform)
	assert.Nil(t, req.DockerConfig)
	assert.Equal(t, "mirror.local:5000/aquasec/trivy-db:2", req.DBRepository)

	image := runReqOf(runningImage{ID: imageID, NodeID: "n1", RemoteRef: "alpine@" + indexDigest,
		Platform: "linux/amd64"}, cfg, volume, []byte(`{"auths":{}}`), imagescanrun.FormatCycloneDX)
	assert.Empty(t, image.Mode)
	assert.Equal(t, "alpine@"+indexDigest, image.RemoteRef)
	assert.Equal(t, imagescanrun.FormatCycloneDX, image.Format)
	assert.NotNil(t, image.DockerConfig)
}

// An artifact's scan is of the build its task binds, which a report of a file
// does not name.
func TestReadScanOfAnArtifact(t *testing.T) {
	target := artifactTargets(walgArtifacts(), []runningImage{{ID: imageID, NodeID: "n2"}},
		map[string]base.CPUArch{"n2": base.CPUArchARM64})[0]
	out := []byte(`{"exitCode":0,"version":null,"report":{"Trivy":{"Version":"0.75.0"},` +
		`"ArtifactName":"/scan/artifact","ArtifactType":"filesystem","Results":[]}}`)

	report, err := readScan(&imagescanrun.RunResp{Source: base.ImageScanSourceArtifact, Output: out}, target,
		&entity.ImageScanSettings{}, time.Now())

	if assert.NoError(t, err) {
		assert.Equal(t, "sha256:"+arm64SHA, report.Scan.ImageID)
		assert.Equal(t, "wal-g v3.0.8", report.Scan.ImageRef)
		assert.Equal(t, "linux/arm64", report.Scan.Platform)
		assert.Equal(t, base.ImageScanSourceArtifact, report.Scan.Source)
	}
}

// artifactsOf has an app's artifacts and has them on nodes, recording which.
type artifactsOf struct {
	artifactservice.Service
	artifacts *entity.AppArtifactsSettings
	ensured   []*artifactservice.EnsureOnNodeReq
	sha       string
}

func (f *artifactsOf) ArtifactsOf(context.Context, database.IDB, string) (*entity.AppArtifactsSettings, error) {
	return f.artifacts, nil
}

func (f *artifactsOf) EnsureOnNode(_ context.Context, req *artifactservice.EnsureOnNodeReq) (
	*artifactservice.EnsureOnNodeResp, error) {
	f.ensured = append(f.ensured, req)
	sha := f.sha
	if sha == "" {
		sha = req.Artifact.Files[req.Arch].SHA256
	}
	return &artifactservice.EnsureOnNodeResp{SHA256: sha, Arch: req.Arch}, nil
}

// archNodes is a swarm whose nodes have the architectures it is given.
type archNodes struct {
	docker.Manager
	arch map[string]string
}

func (n *archNodes) NodeInspect(_ context.Context, nodeID string, _ ...docker.NodeInspectOption) (
	*client.NodeInspectResult, error) {
	arch, ok := n.arch[nodeID]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return &client.NodeInspectResult{Node: swarm.Node{ID: nodeID,
		Description: swarm.NodeDescription{Platform: swarm.Platform{Architecture: arch}}}}, nil
}

// The architecture of each node is the swarm's, and the targets follow it.
func TestAppArtifacts(t *testing.T) {
	svc := &service{artifactService: &artifactsOf{artifacts: walgArtifacts()},
		dockerManager: &archNodes{arch: map[string]string{"n1": "aarch64"}}}

	targets, err := svc.appArtifacts(context.Background(), &entity.App{ID: "a1"},
		[]runningImage{{ID: imageID, NodeID: "n1"}})

	assert.NoError(t, err)
	if assert.Len(t, targets, 1) {
		assert.Equal(t, "sha256:"+arm64SHA, targets[0].ID)
	}
}

// The node a scan runs on is given the build the task binds, whatever its own
// architecture; a file that is not that build is not scanned.
func TestAnArtifactIsHadWhereItIsScanned(t *testing.T) {
	artifacts := &artifactsOf{artifacts: walgArtifacts()}
	svc := &service{artifactService: artifacts}
	target := artifactTargets(walgArtifacts(), []runningImage{{ID: imageID, NodeID: "n2"}},
		map[string]base.CPUArch{"n2": base.CPUArchARM64})[0]

	assert.NoError(t, svc.ensureArtifact(context.Background(), "n1", target))
	if assert.Len(t, artifacts.ensured, 1) {
		assert.Equal(t, "n1", artifacts.ensured[0].NodeID)
		assert.Equal(t, base.CPUArchARM64, artifacts.ensured[0].Arch)
	}

	artifacts.sha = amd64SHA
	assert.ErrorIs(t, svc.ensureArtifact(context.Background(), "n1", target), hperrors.ErrImageScanFailed)
}

// runningNodes is a swarm whose service runs one task, on n1, an amd64 node.
type runningNodes struct {
	archNodes
}

func (n *runningNodes) ServiceTaskList(context.Context, string, []swarm.TaskState, ...docker.TaskListOption) (
	*client.TaskListResult, error) {
	return &client.TaskListResult{Items: []swarm.Task{{NodeID: "n1", Status: swarm.TaskStatus{
		State: swarm.TaskStateRunning, ContainerStatus: &swarm.ContainerStatus{ContainerID: "c1"}}}}}, nil
}

// noSettings has no setting of any type: the app has no registry credential.
type noSettings struct{ repository.SettingRepo }

func (noSettings) List(context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return nil, nil, nil
}

// An app's scan lists the files of its artifacts beside its images, and scans
// each as it does an image: here, each scanned within the hour already.
func TestAnAppsScanListsItsArtifacts(t *testing.T) {
	repo := &appScans{}
	svc := &service{imageScanRepo: repo, nodeExecService: nodesOf{"n1": imageA},
		artifactService: &artifactsOf{artifacts: walgArtifacts()}, settingRepo: noSettings{},
		dockerManager: &runningNodes{archNodes{arch: map[string]string{"n1": "x86_64"}}}}

	out, err := svc.scanApp(context.Background(), &entity.App{ID: "app", Name: "pg", ServiceID: "s1"}, nil,
		&imageScanReq{log: tasklog.NewNullStore()})

	assert.NoError(t, err)
	assert.Equal(t, []string{imageA, "sha256:" + amd64SHA}, repo.written.ImageIDs)
	if assert.Len(t, out.Images, 2) {
		assert.Equal(t, "wal-g v3.0.8", out.Images[1].ImageRef)
		assert.True(t, out.Images[1].Skipped)
	}
}
