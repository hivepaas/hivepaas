package systembackupdto

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const repoID = "01JAB9XED0GTXBSQDFVYAJ8WS1"

func invalidFields(t *testing.T, req *UpdateSystemBackupReq) string {
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

func backupReq(mod func(*SystemBackupBaseReq)) *UpdateSystemBackupReq {
	req := NewUpdateSystemBackupReq()
	req.SystemBackupBaseReq = &SystemBackupBaseReq{
		Status:           base.SettingStatusActive,
		Schedule:         ScheduleReq{Interval: timeutil.Duration(24 * time.Hour)},
		IncludeDB:        true,
		IncludeSpec:      true,
		SpecSecrets:      specmodel.SecretsModeEncrypted,
		SpecPassphrase:   "correct horse battery staple",
		TargetRepository: basedto.ObjectIDReq{ID: repoID},
	}
	mod(req.SystemBackupBaseReq)
	return req
}

func TestASystemBackupOfTheDatabaseAndTheSpec(t *testing.T) {
	req := backupReq(func(*SystemBackupBaseReq) {})

	assert.Equal(t, "", invalidFields(t, req))
	backup := req.ToEntity()
	assert.True(t, backup.IncludeDB)
	assert.True(t, backup.IncludeSpec)
	assert.Equal(t, string(specmodel.SecretsModeEncrypted), backup.SpecSecrets)
	assert.Equal(t, repoID, backup.TargetRepository.ID)
	plain, err := backup.SpecPassphrase.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "correct horse battery staple", plain)
}

// A system backup takes something, into a repository.
func TestASystemBackupTakesSomethingIntoARepository(t *testing.T) {
	assert.Contains(t, invalidFields(t, backupReq(func(req *SystemBackupBaseReq) {
		req.IncludeDB, req.IncludeSpec = false, false
	})), "includeDB")
	assert.Contains(t, invalidFields(t, backupReq(func(req *SystemBackupBaseReq) {
		req.TargetRepository = basedto.ObjectIDReq{}
	})), "targetRepository")

	// A disabled backup is not asked for what it would take, nor where.
	assert.Equal(t, "", invalidFields(t, backupReq(func(req *SystemBackupBaseReq) {
		req.Status = base.SettingStatusDisabled
		req.IncludeDB, req.IncludeSpec = false, false
		req.TargetRepository = basedto.ObjectIDReq{}
	})))
}

// The spec's secrets are one of the export's modes; encrypted asks for a
// passphrase, the others need none.
func TestASystemBackupsSpecSecrets(t *testing.T) {
	assert.Contains(t, invalidFields(t, backupReq(func(req *SystemBackupBaseReq) {
		req.SpecPassphrase = ""
	})), "specPassphrase")
	assert.Contains(t, invalidFields(t, backupReq(func(req *SystemBackupBaseReq) {
		req.SpecSecrets = "reveal"
	})), "specSecrets")
	assert.Equal(t, "", invalidFields(t, backupReq(func(req *SystemBackupBaseReq) {
		req.SpecSecrets, req.SpecPassphrase = specmodel.SecretsModeOmit, ""
	})))
	assert.Equal(t, "", invalidFields(t, backupReq(func(req *SystemBackupBaseReq) {
		req.IncludeSpec, req.SpecSecrets, req.SpecPassphrase = false, "", ""
	})), "no spec, no secrets mode")
}

// The passphrase the GET response masks is kept as stored when it comes back
// masked.
func TestASystemBackupKeepsAMaskedPassphrase(t *testing.T) {
	stored := backupReq(func(*SystemBackupBaseReq) {}).ToEntity()
	req := backupReq(func(req *SystemBackupBaseReq) { req.SpecPassphrase = basedto.MaskedSecret })

	backup := req.ToEntity()
	req.KeepMaskedSecrets(backup, stored)

	plain, err := backup.SpecPassphrase.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "correct horse battery staple", plain)
}
