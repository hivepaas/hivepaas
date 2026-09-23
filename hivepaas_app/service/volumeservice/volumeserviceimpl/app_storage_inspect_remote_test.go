package volumeserviceimpl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// fakeNodeExec stands in for the node holding a volume: it records what it was
// asked and answers with the paths it was told to call non-empty.
type fakeNodeExec struct {
	req      *nodeexecservice.CommandExecReq
	nonEmpty []string
	err      error
}

func (f *fakeNodeExec) ExecCommand(
	_ context.Context,
	req *nodeexecservice.CommandExecReq,
) (*nodeexecservice.CommandExecResp, error) {
	f.req = req
	if f.err != nil {
		return nil, f.err
	}
	for _, path := range f.nonEmpty {
		_, _ = req.Stdout.Write([]byte(path + "\n"))
	}
	return &nodeexecservice.CommandExecResp{}, nil
}

// pinnedVolume is a managed volume at the directory every test here names, on
// the node they all ask.
func pinnedVolume(t *testing.T) *entity.Setting {
	t.Helper()
	setting := clusterVolumeSetting(t, "vol-1", "/srv/data")
	const nodeID = "node-b"
	vol, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	vol.NodeID = nodeID
	data, err := json.Marshal(vol)
	assert.NoError(t, err)
	setting.Data = string(data)
	return setting
}

func TestStorageOnAnotherNodeIsReadThroughItsAgent(t *testing.T) {
	setting := pinnedVolume(t)
	states := []*volumeservice.AppStorageState{
		{AppKey: "web", Path: "prod/web"},
		{AppKey: "db", Path: "prod/db"},
	}
	exec := &fakeNodeExec{nonEmpty: []string{"/host/srv/data/prod/db"}}
	svc := &service{nodeExecService: exec}

	svc.inspectVolumeOnItsNode(t.Context(), setting, states)

	// Asked on the node the volume is pinned to, about both apps.
	assert.Equal(t, "node-b", exec.req.NodeID)
	assert.Contains(t, exec.req.Command, "/host/srv/data/prod/web")
	assert.Contains(t, exec.req.Command, "/host/srv/data/prod/db")
	assert.Contains(t, exec.req.Command, "/srv/data/prod/db")

	// Both were asked about, so both have an answer; only one holds anything.
	assert.True(t, states[0].Checked)
	assert.False(t, states[0].HasData())
	assert.True(t, states[1].HasData())
}

// The runtime that rewrites a bind source serves the same root at its own
// prefix, so both spellings are offered and either answer counts.
func TestStorageOnAnotherNodeAcceptsEitherHostSpelling(t *testing.T) {
	setting := pinnedVolume(t)
	states := []*volumeservice.AppStorageState{{AppKey: "db", Path: "prod/db"}}
	exec := &fakeNodeExec{nonEmpty: []string{"/host/host_mnt/srv/data/prod/db"}}

	(&service{nodeExecService: exec}).inspectVolumeOnItsNode(t.Context(), setting, states)

	assert.True(t, states[0].HasData())
}

// Nobody to ask is not the same as nothing there: an unpinned volume this node
// cannot open leaves every state unchecked rather than reporting it clean.
func TestStorageWithNoNodeToAskStaysUnchecked(t *testing.T) {
	setting := clusterVolumeSetting(t, "vol-1", "/srv/data") // no node id or label
	states := []*volumeservice.AppStorageState{{AppKey: "db", Path: "prod/db"}}
	exec := &fakeNodeExec{}

	(&service{nodeExecService: exec}).inspectVolumeOnItsNode(t.Context(), setting, states)

	assert.Nil(t, exec.req)
	assert.False(t, states[0].Checked)
}

func TestStorageOnANodeThatCannotBeAskedStaysUnchecked(t *testing.T) {
	setting := pinnedVolume(t)
	states := []*volumeservice.AppStorageState{{AppKey: "db", Path: "prod/db"}}
	exec := &fakeNodeExec{err: hperrors.ErrUnavailable}

	(&service{nodeExecService: exec}).inspectVolumeOnItsNode(t.Context(), setting, states)

	assert.False(t, states[0].Checked)
}

func TestHostCandidatesCoverEveryWayAnAgentSeesTheHost(t *testing.T) {
	assert.Equal(t,
		[]string{
			"/host/srv/data/prod/db",
			"/host/host_mnt/srv/data/prod/db",
			// An agent running on the host itself, which is how a development
			// installation runs one.
			"/srv/data/prod/db",
		},
		hostCandidatesFor("/srv/data/prod/db"))
}

// An agent on the host answers about the path as it is.
func TestStorageOnANodeWhoseAgentRunsOnTheHost(t *testing.T) {
	setting := pinnedVolume(t)
	states := []*volumeservice.AppStorageState{{AppKey: "db", Path: "prod/db"}}
	exec := &fakeNodeExec{nonEmpty: []string{"/srv/data/prod/db"}}

	(&service{nodeExecService: exec}).inspectVolumeOnItsNode(t.Context(), setting, states)

	assert.True(t, states[0].HasData())
}
