package apptemplatedto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

func TestAManagedRequestNamesExactlyOneArchive(t *testing.T) {
	repo, storage := &basedto.ObjectIDReq{ID: "repo-1"}, &basedto.ObjectIDReq{ID: "cs-1"}
	for name, tc := range map[string]struct {
		managed *ManagedCreateReq
		valid   bool
	}{
		"a repository":     {&ManagedCreateReq{Storage: ManagedStorageReq{BackupRepo: repo}}, true},
		"a cloud storage":  {&ManagedCreateReq{Storage: ManagedStorageReq{CloudStorage: storage}}, true},
		"none":             {&ManagedCreateReq{}, false},
		"both":             {&ManagedCreateReq{Storage: ManagedStorageReq{BackupRepo: repo, CloudStorage: storage}}, false},
		"a retention of 0": {&ManagedCreateReq{Storage: ManagedStorageReq{BackupRepo: repo}}, true},
		"a negative retention": {&ManagedCreateReq{Storage: ManagedStorageReq{BackupRepo: repo},
			RetainFull: -1}, false},
	} {
		const project = "01HZZZZZZZZZZZZZZZZZZZZZZZ"
		req := &CreateAppFromTemplateReq{ProjectID: project, ProjectEnvID: project + ":prod",
			Name: "db", Template: "postgres-managed", Managed: tc.managed}
		errs := req.Validate()
		assert.Equal(t, tc.valid, len(errs) == 0, "%s: %v", name, errs)
	}
}
