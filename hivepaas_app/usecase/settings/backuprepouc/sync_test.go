package backuprepouc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// repoConfigChanged decides whether a sync writes at all, so a wrong "no" silently keeps the
// setting lying about the repository, and a wrong "yes" bumps the version on every sync.
func TestRepoConfigChanged(t *testing.T) {
	t.Parallel()

	retention := func(last int) *entity.BackupRetentionPolicy {
		return &entity.BackupRetentionPolicy{KeepLast: last, KeepDaily: 7}
	}

	tests := []struct {
		name   string
		before *entity.BackupRepo
		after  *entity.BackupRepo
		want   bool
	}{
		{
			name:   "identical",
			before: &entity.BackupRepo{PackSize: 32 * unit.MB, Compression: "zstd", Retention: retention(5)},
			after:  &entity.BackupRepo{PackSize: 32 * unit.MB, Compression: "zstd", Retention: retention(5)},
			want:   false,
		},
		{
			name:   "pack size differs",
			before: &entity.BackupRepo{PackSize: 32 * unit.MB},
			after:  &entity.BackupRepo{PackSize: 48 * unit.MB},
			want:   true,
		},
		{
			name:   "compression differs",
			before: &entity.BackupRepo{Compression: "zstd-fastest"},
			after:  &entity.BackupRepo{Compression: "zstd-better-compression"},
			want:   true,
		},
		{
			// Distinct pointers holding equal values must not read as a change: the repository
			// hands back a freshly built policy every time it is read.
			name:   "retention equal through different pointers",
			before: &entity.BackupRepo{Retention: retention(5)},
			after:  &entity.BackupRepo{Retention: retention(5)},
			want:   false,
		},
		{
			name:   "retention values differ",
			before: &entity.BackupRepo{Retention: retention(5)},
			after:  &entity.BackupRepo{Retention: retention(9)},
			want:   true,
		},
		{
			name:   "retention appeared",
			before: &entity.BackupRepo{},
			after:  &entity.BackupRepo{Retention: retention(5)},
			want:   true,
		},
		{
			name:   "retention removed",
			before: &entity.BackupRepo{Retention: retention(5)},
			after:  &entity.BackupRepo{},
			want:   true,
		},
		{
			// Only what the repository holds counts. Everything else on the setting belongs to the
			// app, and a sync must never treat it as drift.
			name:   "only app-owned fields differ",
			before: &entity.BackupRepo{Description: "before", StoragePrefix: "a"},
			after:  &entity.BackupRepo{Description: "after", StoragePrefix: "b"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, repoConfigChanged(tt.before, tt.after))
		})
	}
}
