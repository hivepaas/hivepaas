package backuprepouc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// retentionToApply is what an update tells the repository about its retention: a missed change
// leaves the repository keeping by the old rules while the setting shows the new ones, until a
// sync reads the old ones back over them.
func TestRetentionToApply(t *testing.T) {
	t.Parallel()

	retention := func(last, daily int) *entity.BackupRetentionPolicy {
		return &entity.BackupRetentionPolicy{KeepLast: last, KeepDaily: daily}
	}

	tests := []struct {
		name   string
		before *entity.BackupRetentionPolicy
		after  *entity.BackupRetentionPolicy
		want   *entity.BackupRetentionPolicy
	}{
		{name: "unchanged", before: retention(5, 7), after: retention(5, 7), want: nil},
		{name: "a rule raised", before: retention(5, 7), after: retention(6, 7), want: retention(6, 7)},
		{name: "a rule lowered to zero", before: retention(5, 7), after: retention(5, 0), want: retention(5, 0)},
		{name: "set where there was none", before: nil, after: retention(2, 0), want: retention(2, 0)},
		{name: "none asked for", before: retention(5, 7), after: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, retentionToApply(tt.before, tt.after))
		})
	}
}
