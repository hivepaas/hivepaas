package schedjobdto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const (
	backupApp = "01JAB9XED0GTXBSQDFVYAJ8WB1"
	repoA     = "01JAB9XED0GTXBSQDFVYAJ8WB2"
	volA      = "01JAB9XED0GTXBSQDFVYAJ8WB3"
)

func dataBackupReq(dataBackup *SchedJobDataBackupReq) *CreateSchedJobReq {
	dataBackup.TargetRepository = basedto.ObjectIDReq{ID: repoA}
	req := NewCreateSchedJobReq()
	req.SchedJobBaseReq = &SchedJobBaseReq{
		Name:       "nightly db",
		JobType:    base.SchedJobTypeDataBackup,
		App:        basedto.ObjectIDReq{ID: backupApp},
		DataBackup: dataBackup,
	}
	return req
}

func commandBackupReq() *SchedJobDataBackupReq {
	return &SchedJobDataBackupReq{
		Source:         base.SchedJobDataBackupSourceCommand,
		SourceCommand:  &commandtemplatedto.CommandTemplateBaseReq{Command: "pg_dump app"},
		SourceFileName: "db.sql",
	}
}

func volumeBackupReq(subpath string) *SchedJobDataBackupReq {
	return &SchedJobDataBackupReq{
		Source:              base.SchedJobDataBackupSourceVolume,
		SourceVolume:        basedto.ObjectIDReq{ID: volA},
		SourceVolumeSubpath: subpath,
	}
}

func TestADataBackupOfACommand(t *testing.T) {
	req := dataBackupReq(commandBackupReq())
	req.DataBackup.Tags = map[string]string{"env": "prod"}

	assert.Equal(t, "", invalidFields(t, req))
	job := req.ToEntity()
	if assert.NotNil(t, job.DataBackup) {
		assert.Equal(t, "pg_dump app", job.DataBackup.SourceCommand.Command)
		assert.Equal(t, "db.sql", job.DataBackup.SourceFileName)
		assert.Equal(t, repoA, job.DataBackup.TargetRepository.ID)
		assert.Equal(t, map[string]string{"env": "prod"}, job.DataBackup.Tags)
	}
	assert.Nil(t, job.Command, "the command is the backup's, not the job's")
}

func TestADataBackupOfACommandNeedsTheCommandAndAFileName(t *testing.T) {
	req := dataBackupReq(commandBackupReq())
	req.DataBackup.SourceCommand = nil
	assert.Contains(t, invalidFields(t, req), "dataBackup.sourceCommand")

	for _, name := range []string{"", "dumps/db.sql", "db sql", strings.Repeat("a", 101)} {
		req = dataBackupReq(commandBackupReq())
		req.DataBackup.SourceFileName = name
		assert.Contains(t, invalidFields(t, req), "dataBackup.sourceFileName", name)
	}

	req = dataBackupReq(commandBackupReq())
	req.DataBackup.SourceVolume = basedto.ObjectIDReq{ID: volA}
	assert.Contains(t, invalidFields(t, req), "dataBackup.sourceVolume")
}

func TestADataBackupOfAVolume(t *testing.T) {
	req := dataBackupReq(volumeBackupReq("uploads/2026"))

	assert.Equal(t, "", invalidFields(t, req))
	assert.Equal(t, volA, req.ToEntity().DataBackup.SourceVolume.ID)
	assert.Equal(t, "", invalidFields(t, dataBackupReq(volumeBackupReq(""))), "all of the volume")
}

func TestADataBackupSubpathStaysInsideTheVolume(t *testing.T) {
	for _, subpath := range []string{"/etc", "../a2", "uploads/../../a2", ".."} {
		assert.Contains(t, invalidFields(t, dataBackupReq(volumeBackupReq(subpath))),
			"dataBackup.sourceVolumeSubpath", subpath)
	}
	req := dataBackupReq(volumeBackupReq(""))
	req.DataBackup.SourceVolume = basedto.ObjectIDReq{}
	assert.Contains(t, invalidFields(t, req), "dataBackup.sourceVolume")
}

func TestADataBackupNeedsARepositoryAndASource(t *testing.T) {
	req := dataBackupReq(commandBackupReq())
	req.DataBackup.TargetRepository = basedto.ObjectIDReq{}
	assert.Contains(t, invalidFields(t, req), "dataBackup.targetRepository")

	req = dataBackupReq(commandBackupReq())
	req.DataBackup.Source = "database"
	assert.Contains(t, invalidFields(t, req), "dataBackup.source")

	req = dataBackupReq(commandBackupReq())
	req.DataBackup = nil
	assert.Contains(t, invalidFields(t, req), "dataBackup")
}

// Tags are kopia's: a key without ':' - HivePaaS's own keys are reserved - and
// a short value.
func TestADataBackupsTags(t *testing.T) {
	bad := []map[string]string{
		{"a:b": "c"},
		{"": "c"},
		{"hivepaas.job": "x"},
		{"env": ""},
		{"env": strings.Repeat("v", 101)},
		{"env": "two words"},
	}
	for _, tags := range bad {
		req := dataBackupReq(commandBackupReq())
		req.DataBackup.Tags = tags
		assert.Contains(t, invalidFields(t, req), "dataBackup.tags", tags)
	}

	many := map[string]string{}
	for i := range maxDataBackupTags + 1 {
		many[string(rune('a'+i))] = "v"
	}
	req := dataBackupReq(commandBackupReq())
	req.DataBackup.Tags = many
	assert.Contains(t, invalidFields(t, req), "dataBackup.tags")
}

// A data backup's command is its source's; it has no command, command output
// or app other than its own, and no other type has a data backup.
func TestADataBackupRunsNoCommandOfItsOwn(t *testing.T) {
	req := dataBackupReq(commandBackupReq())
	req.Command = &commandtemplatedto.CommandTemplateBaseReq{Command: "echo hi"}
	assert.Contains(t, invalidFields(t, req), "command")

	req = dataBackupReq(commandBackupReq())
	req.CommandOutput = &CommandOutputReq{}
	assert.Contains(t, invalidFields(t, req), "commandOutput")

	req = dataBackupReq(commandBackupReq())
	req.App = basedto.ObjectIDReq{}
	assert.Contains(t, invalidFields(t, req), "app")

	req = dataBackupReq(commandBackupReq())
	req.JobType = base.SchedJobTypeSystemCleanup
	assert.Contains(t, invalidFields(t, req), "dataBackup")
}

func TestTransformSchedJobNamesADataBackupsVolumeAndRepository(t *testing.T) {
	job := &entity.SchedJob{
		JobType: base.SchedJobTypeDataBackup,
		App:     entity.ObjectID{ID: backupApp},
		DataBackup: &entity.SchedJobDataBackup{
			Source:           base.SchedJobDataBackupSourceVolume,
			SourceVolume:     entity.ObjectID{ID: volA},
			TargetRepository: entity.ObjectID{ID: repoA},
			Tags:             map[string]string{"env": "prod"},
		},
	}
	setting := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob, Name: "files"}
	setting.MustSetData(job)
	refObjects := entity.NewRefObjects()
	refObjects.RefApps[backupApp] = &entity.App{ID: backupApp, Name: "backend"}
	refObjects.RefSettings[repoA] = &entity.Setting{ID: repoA, Type: base.SettingTypeBackupRepo, Name: "s3 repo",
		Status: base.SettingStatusActive}

	resp, err := TransformSchedJob(setting, refObjects, false)

	assert.NoError(t, err)
	if assert.NotNil(t, resp.DataBackup) {
		assert.Equal(t, "s3 repo", resp.DataBackup.TargetRepository.Name)
		// The volume is gone: it is named missing.
		assert.Equal(t, volA, resp.DataBackup.SourceVolume.ID)
		assert.Equal(t, map[string]string{"env": "prod"}, resp.DataBackup.Tags)
		assert.Nil(t, resp.DataBackup.SourceCommand)
	}
}
