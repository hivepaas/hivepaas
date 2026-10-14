package imagescanserviceimpl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/moby/moby/api/types/registry"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	agentimagescan "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	imageID     = "sha256:19a85ffa0ee14b7c64ca5104dc4e1c5021fa04bd6b4a692074691ce154f5d074"
	indexDigest = "sha256:451eee8bedcb2f029756dc3e9d73bab0e7943c1ac55cff3a4861c52a0fdd3e98"
)

// docker inspect's lines are an image's ID and the reference the container was
// made with; anything else is passed over.
func TestParseContainerImages(t *testing.T) {
	out := imageID + " registry.local/shop/api:abc123@sha256:aaaa\n" +
		"Error: No such object: c9\n" +
		imageID + " registry.local/shop/api:abc123\n"
	images := parseContainerImages(out, "n1")
	assert.Equal(t, []runningImage{
		{ID: imageID, Ref: "registry.local/shop/api:abc123", NodeID: "n1"},
		{ID: imageID, Ref: "registry.local/shop/api:abc123", NodeID: "n1"},
	}, images)
}

// docker image inspect's lines are an image's ID, the node's platform and the
// digests it was pulled or pushed by; the digest a node without the image reads
// it by is its repository's.
func TestParseImageDetails(t *testing.T) {
	out := imageID + ` linux/arm64 ["mirror.local/alpine@` + indexDigest + `","alpine@` + indexDigest + `"]` +
		"\nsha256:" + "0000000000000000000000000000000000000000000000000000000000000000" + " linux/amd64 []\n" +
		"Error: No such image: x\n"
	details := parseImageDetails(out)
	if !assert.Len(t, details, 2) {
		return
	}
	detail := details[imageID]
	assert.Equal(t, "linux/arm64", detail.platform)
	assert.Equal(t, "alpine@"+indexDigest, repoDigestFor("alpine:3.10", detail.repoDigests))
	assert.Equal(t, "alpine@"+indexDigest, repoDigestFor("docker.io/library/alpine:3.10", detail.repoDigests))
	assert.Equal(t, "mirror.local/alpine@"+indexDigest, repoDigestFor("other:1", detail.repoDigests),
		"the same content, by whatever name")
	assert.Empty(t, repoDigestFor("built-here:1", nil), "built on the node, never pushed")
}

// Trivy's error is shown without its time and level; a fatal one with the
// lines its message goes on over.
func TestTrivyErrors(t *testing.T) {
	assert.Equal(t, "Fatal error	run error: twirp error unauthenticated: invalid token",
		trivyMessage("2026-10-10T18:18:09Z\tFATAL\tFatal error\trun error: twirp error unauthenticated: invalid token"))
	assert.Equal(t, "plain", trivyMessage(" plain "))

	log := "2026-10-14T13:21:13Z\tINFO\t[vuln] Vulnerability scanning is enabled\n" +
		"2026-10-14T13:21:13Z\tFATAL\tFatal error\trun error: unable to find the specified image\n" +
		"\t* remote error: UNAUTHORIZED: authentication required\n"
	assert.Equal(t, "Fatal error\trun error: unable to find the specified image\n"+
		"\t* remote error: UNAUTHORIZED: authentication required", trivyMessage(lastErrorLine(log)))
	assert.Equal(t, "the last line", lastErrorLine("a\nthe last line\n"))
}

func volumeOf(vol *entity.ClusterVolume, labelNodes ...string) *scanVolume {
	return &scanVolume{
		VolumeMount: &volumeservice.VolumeMount{Setting: &entity.Setting{ID: "vol1"}, Volume: vol},
		labelNodes:  labelNodes,
	}
}

// A volume pinned to a node has every scan run there; one pinned by label, on
// the node of the label running the image when one does; a shared one, on the
// node running the image. One scan at a time per copy of the volume.
func TestWhereScansRun(t *testing.T) {
	pinned := volumeOf(&entity.ClusterVolume{NodeID: "n1"})
	assert.Equal(t, "n1", pinned.scanNode("n2"))
	assert.Equal(t, "imagescan:volume:lock:vol1:n1", pinned.lockKey("n1"))

	byLabel := volumeOf(&entity.ClusterVolume{NodeLabel: "scan=yes"}, "n1", "n3")
	assert.Equal(t, "n3", byLabel.scanNode("n3"))
	assert.Equal(t, "n1", byLabel.scanNode("n2"))
	assert.Equal(t, "imagescan:volume:lock:vol1:n3", byLabel.lockKey("n3"), "a copy on each node")

	shared := volumeOf(&entity.ClusterVolume{})
	assert.Equal(t, "n2", shared.scanNode("n2"))
	assert.Equal(t, "imagescan:volume:lock:vol1", shared.lockKey("n2"), "one copy for the cluster")
	assert.Equal(t, "/data/hivepaas/scanning/trivy", cacheDir())
}

// The app's credential reaches only its own registry's images: under the key
// Docker Hub's credentials are looked up by, or the registry's host.
func TestDockerConfig(t *testing.T) {
	auth := func(address string) *registry.AuthConfig {
		return &registry.AuthConfig{Username: "u", Password: "p", ServerAddress: address}
	}
	keyOf := func(data []byte) string {
		var cfg struct {
			Auths map[string]struct {
				Auth string `json:"auth"`
			} `json:"auths"`
		}
		if !assert.NoError(t, json.Unmarshal(data, &cfg)) {
			return ""
		}
		for key, entry := range cfg.Auths {
			assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("u:p")), entry.Auth)
			return key
		}
		return ""
	}

	data, err := dockerConfig(auth("https://registry.example.com/v2/"), "registry.example.com/shop/api@"+indexDigest)
	assert.NoError(t, err)
	assert.Equal(t, "registry.example.com", keyOf(data))

	data, _ = dockerConfig(auth("index.docker.io"), "acme/api@"+indexDigest)
	assert.Equal(t, hubAuthKey, keyOf(data))

	data, _ = dockerConfig(auth("registry.example.com"), "alpine@"+indexDigest)
	assert.Nil(t, data, "another registry's image is read as anybody may")
	data, _ = dockerConfig(nil, "alpine@"+indexDigest)
	assert.Nil(t, data)
}

// A container that printed its report is the image's scan, by the ID the node
// gives it, read where it was; one that did not failed with Trivy's error.
func TestReadScan(t *testing.T) {
	img := runningImage{ID: indexDigest, Ref: "alpine:3.10", NodeID: "n2"}
	cfg := &entity.ImageScanSettings{}
	at := time.Now()

	out := []byte(`{"exitCode":0,"version":{"VulnerabilityDB":{"UpdatedAt":"2026-10-14T07:11:52Z"}},` +
		`"report":{"Trivy":{"Version":"0.75.0"},"Metadata":{"ImageID":"` + imageID + `"},"Results":[]}}`)
	report, err := readScan(&imagescanrun.RunResp{Source: base.ImageScanSourceRegistry, Output: out}, img, cfg, at)
	if assert.NoError(t, err) {
		assert.Equal(t, indexDigest, report.Scan.ImageID)
		assert.Equal(t, base.ImageScanSourceRegistry, report.Scan.Source)
		assert.Equal(t, time.Date(2026, 10, 14, 7, 11, 52, 0, time.UTC), report.Scan.DBUpdatedAt)
	}

	_, err = readScan(&imagescanrun.RunResp{ExitCode: 1, Output: []byte(`{"exitCode":1,"version":null,"report":null}`),
		Log: "2026-10-14T13:21:13Z\tFATAL\tFatal error\tcache may be in use by another process\n"}, img, cfg, at)
	assert.ErrorIs(t, err, hperrors.ErrImageScanFailed)
	assert.Contains(t, hperrors.GetErrorDetail(err, ""), "cache may be in use by another process")

	_, err = readScan(&imagescanrun.RunResp{ExitCode: 137, OOMKilled: true}, img, cfg, at)
	assert.Contains(t, hperrors.GetErrorDetail(err, ""), "ran out of its memory limit")
}

// settingRepo answers the volumes it holds by ID.
type settingRepo struct {
	repository.SettingRepo
	volumes map[string]*entity.ClusterVolume
}

func (r *settingRepo) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	vol, ok := r.volumes[id]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	setting := &entity.Setting{ID: id, Name: id, Type: base.SettingTypeClusterVolume}
	setting.MustSetData(vol)
	return setting, nil
}

// Switched on, scanning needs a volume that is there and is not memory.
func TestValidateTheVolume(t *testing.T) {
	svc := &service{settingRepo: &settingRepo{volumes: map[string]*entity.ClusterVolume{
		"disk":   {NodeID: "n1", Managed: true, Driver: "local"},
		"memory": {Managed: true, Driver: "local", DriverOpts: map[string]string{"type": "tmpfs", "device": "tmpfs"}},
	}}}
	on := func(volumeID string) *entity.ImageScanSettings {
		return &entity.ImageScanSettings{Enabled: true, Volume: entity.ObjectID{ID: volumeID}}
	}
	ctx := context.Background()
	assert.NoError(t, svc.Validate(ctx, nil, on("disk")))
	assert.NoError(t, svc.Validate(ctx, nil, &entity.ImageScanSettings{}), "off, nothing is needed")
	for _, volumeID := range []string{"", "gone", "memory"} {
		err := svc.Validate(ctx, nil, on(volumeID))
		assert.ErrorIs(t, err, hperrors.ErrImageScanSettingsInvalid, volumeID)
	}
}

// node is a node's Docker that names itself.
type node struct {
	docker.Manager
	id string
}

func (n *node) NodeCurrentID(context.Context) (string, error) { return n.id, nil }

// scans counts the scans it runs.
type scans struct{ ran int }

func (s *scans) Run(context.Context, *imagescanrun.RunReq) (*imagescanrun.RunResp, error) {
	s.ran++
	return &imagescanrun.RunResp{Source: base.ImageScanSourceNode}, nil
}

func (s *scans) Close() error { return nil }

// agents answers every node's agent with its scans.
type agents struct {
	agentservice.Service
	asked []string
}

func (a *agents) GetAgentAddrForNode(_ context.Context, nodeID string) (string, error) {
	a.asked = append(a.asked, nodeID)
	return nodeID + ":10001", nil
}

// A scan on the app's own node runs in its process; on another, through that
// node's agent.
func TestScanOnItsNode(t *testing.T) {
	local, remote, agentsOf := &scans{}, &scans{}, &agents{}
	svc := &service{dockerManager: &node{id: "n1"}, localScanner: local, agentService: agentsOf,
		newAgentClient: func(string) (agentimagescan.ImageScanServiceClient, error) { return remote, nil }}

	_, err := svc.scanOn(context.Background(), "n1", &imagescanrun.RunReq{})
	assert.NoError(t, err)
	_, err = svc.scanOn(context.Background(), "n2", &imagescanrun.RunReq{})
	assert.NoError(t, err)
	assert.Equal(t, 1, local.ran)
	assert.Equal(t, 1, remote.ran)
	assert.Equal(t, []string{"n2"}, agentsOf.asked)
}
