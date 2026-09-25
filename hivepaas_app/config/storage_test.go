package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectDataDirs(t *testing.T) {
	for name, tc := range map[string]struct {
		storage              Storage
		wantBase, wantPrefix string
		wantRoot             string
	}{
		// The default is made on first use, so the base is the storage root that
		// the install created and project_data is created under it.
		"the default, under the storage root": {
			Storage{HostDir: "/srv/hivepaas"}, "/srv/hivepaas", "project_data", "/srv/hivepaas/project_data",
		},
		// A directory of the operator's own is used as it is.
		"a directory of the operator's own": {
			Storage{HostDir: "/srv/hivepaas", ProjectDataHostDir: "/data/projects/"},
			"/data/projects", "", "/data/projects",
		},
		"nothing configured": {Storage{}, "", "", ""},
	} {
		base, prefix := tc.storage.ProjectDataDirs()
		assert.Equal(t, tc.wantBase, base, name)
		assert.Equal(t, tc.wantPrefix, prefix, name)
		assert.Equal(t, tc.wantRoot, tc.storage.ProjectDataRoot(), name)
	}
}

func TestStorageValidate(t *testing.T) {
	assert.NoError(t, Storage{}.Validate())
	assert.NoError(t, Storage{ProjectDataHostDir: "/data/projects"}.Validate())
	for _, bad := range []string{"data/projects", "./projects", "/", "/data/../"} {
		assert.ErrorIs(t, Storage{ProjectDataHostDir: bad}.Validate(), ErrStorageInvalid, bad)
	}
}
