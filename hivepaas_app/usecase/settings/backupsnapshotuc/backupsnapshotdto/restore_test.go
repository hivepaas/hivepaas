package backupsnapshotdto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const (
	snapshotRecordID = "01JAB9XED0GTXBSQDFVYAJ8WS9"
	targetApp        = "01M3HT2P22YRG4PE1GGSDABGDF"
	targetVolume     = "01M3HT2P22YRG4PE1GGSDABGD0"
)

func invalidFields(t *testing.T, req *RestoreBackupSnapshotReq) string {
	t.Helper()
	assert.NoError(t, req.ModifyRequest())
	errs := req.Validate()
	if len(errs) == 0 {
		return ""
	}
	paths := make([]string, 0, len(errs))
	for _, inner := range errs.Build(translation.LangEn).InnerErrors {
		paths = append(paths, inner.Path)
	}
	return strings.Join(paths, " ")
}

func commandRestoreReq() *RestoreBackupSnapshotReq {
	req := NewRestoreBackupSnapshotReq()
	req.ID = snapshotRecordID
	req.TargetApp = basedto.ObjectIDReq{ID: targetApp}
	req.Command = &commandtemplatedto.CommandTemplateBaseReq{Command: "psql -U app app", TTY: true}
	return req
}

func volumeRestoreReq(mode base.BackupRestoreMode, stop bool) *RestoreBackupSnapshotReq {
	req := NewRestoreBackupSnapshotReq()
	req.ID = snapshotRecordID
	req.TargetApp = basedto.ObjectIDReq{ID: targetApp}
	req.Volume = basedto.ObjectIDReq{ID: targetVolume}
	req.Subpath = "data/"
	req.SnapshotPath = "./uploads"
	req.StopApp = stop
	req.Mode = mode
	return req
}

// A command restore has a command, and nothing of a volume's; it reads the
// file on its stdin, without a TTY.
func TestRestoreRequestOfACommand(t *testing.T) {
	req := commandRestoreReq()
	assert.Equal(t, "", invalidFields(t, req))
	assert.False(t, req.Command.TTY)

	req = commandRestoreReq()
	req.Mode = base.BackupRestoreModeOverwrite
	assert.Contains(t, invalidFields(t, req), "mode")

	req = commandRestoreReq()
	req.SnapshotPath = "uploads"
	assert.Contains(t, invalidFields(t, req), "snapshotPath")

	req = commandRestoreReq()
	req.Command = &commandtemplatedto.CommandTemplateBaseReq{}
	assert.Contains(t, invalidFields(t, req), "command")
}

// A volume restore names a volume and how to write; paths stay inside.
func TestRestoreRequestOfAVolume(t *testing.T) {
	req := volumeRestoreReq(base.BackupRestoreModeReplace, true)
	assert.Equal(t, "", invalidFields(t, req))
	assert.Equal(t, "data", req.Subpath)
	assert.Equal(t, "uploads", req.SnapshotPath)

	assert.Equal(t, "", invalidFields(t, volumeRestoreReq(base.BackupRestoreModeOverwrite, false)),
		"overwrite may leave the app running")
	assert.Contains(t, invalidFields(t, volumeRestoreReq(base.BackupRestoreModeReplace, false)), "stopApp",
		"replace stops the app")
	assert.Contains(t, invalidFields(t, volumeRestoreReq("merge", true)), "mode")
	assert.Contains(t, invalidFields(t, volumeRestoreReq("", true)), "mode")

	for _, bad := range []string{"/etc", "../app2", "data/../../app2", ".."} {
		req = volumeRestoreReq(base.BackupRestoreModeOverwrite, true)
		req.Subpath = bad
		assert.Contains(t, invalidFields(t, req), "subpath", bad)
		req = volumeRestoreReq(base.BackupRestoreModeOverwrite, true)
		req.SnapshotPath = bad
		assert.Contains(t, invalidFields(t, req), "snapshotPath", bad)
	}
}

// A restore is one of the two, into an app.
func TestRestoreRequestIsACommandsOrAVolumes(t *testing.T) {
	req := volumeRestoreReq(base.BackupRestoreModeOverwrite, true)
	req.Command = &commandtemplatedto.CommandTemplateBaseReq{Command: "psql"}
	assert.Contains(t, invalidFields(t, req), "command")

	req = NewRestoreBackupSnapshotReq()
	req.ID = snapshotRecordID
	req.TargetApp = basedto.ObjectIDReq{ID: targetApp}
	assert.Contains(t, invalidFields(t, req), "command")

	req = commandRestoreReq()
	req.TargetApp = basedto.ObjectIDReq{}
	assert.Contains(t, invalidFields(t, req), "targetApp")
}

// What a job says of restoring its snapshots comes with the snapshot: the file,
// the command that loads it, and the volume it was read from.
func TestTransformBackupSnapshotGivesItsJobsRestore(t *testing.T) {
	refs := refs()
	refs.Jobs["j1"].MustSetData(&entityJobWithRestore)

	resp := TransformBackupSnapshot(snapshotRecord(t, "s1"), []string{"hivepaas.job:j1"}, refs)

	if assert.NotNil(t, resp.Job) {
		assert.Equal(t, "db.sql", resp.Job.FileName)
		if assert.NotNil(t, resp.Job.RestoreCommand) {
			assert.Equal(t, "psql app", resp.Job.RestoreCommand.Command)
			assert.Equal(t, "set -e\npsql app", resp.Job.RestoreCommand.Script, "an inline script comes too")
		}
	}
}
