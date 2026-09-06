package syscleanupserviceimpl

import (
	"testing"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// The floor is what stops a shortened retention from being a way to erase the
// record of one's own reveals.
func TestAuditLogCutoffs(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		retention   time.Duration
		wantGeneral time.Duration // age at which a general entry may go
		wantReveal  time.Duration // age at which a reveal entry may go
	}{
		{
			name:        "a retention shorter than the floor does not shorten reveals",
			retention:   timeutil.Day,
			wantGeneral: timeutil.Day,
			wantReveal:  90 * timeutil.Day,
		},
		{
			name:        "a retention longer than the floor wins for reveals too",
			retention:   365 * timeutil.Day,
			wantGeneral: 365 * timeutil.Day,
			wantReveal:  365 * timeutil.Day,
		},
		{
			name:        "the floor itself",
			retention:   90 * timeutil.Day,
			wantGeneral: 90 * timeutil.Day,
			wantReveal:  90 * timeutil.Day,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			general, reveal := auditLogCutoffs(tt.retention, now)
			if got := now.Sub(general); got != tt.wantGeneral {
				t.Errorf("general cutoff age = %v, want %v", got, tt.wantGeneral)
			}
			if got := now.Sub(reveal); got != tt.wantReveal {
				t.Errorf("reveal cutoff age = %v, want %v", got, tt.wantReveal)
			}
			if reveal.After(general) {
				t.Error("reveal entries must never be swept earlier than general ones")
			}
		})
	}
}
