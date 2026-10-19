package postgres

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	// markLSN is what walg-mark prints, as KEY=value: the point's LSN. Its
	// arguments are the name and how many seconds to wait for the WAL it is in
	// to be archived (design §4, "Restore points").
	markLSN = "LSN"
	// markWaits is how many archive timeouts walg-mark waits: the segment it
	// switches to is archived at once, and archive_timeout bounds how late.
	markWaits = 2
	// reservedPointPrefix is HivePaaS's own points', such as the restore test's.
	reservedPointPrefix = "hp-"
	// pointNameMax is what Postgres keeps of a point's name, in bytes.
	pointNameMax = 63
)

// pointNamePattern is what a point's name may be: Postgres takes up to 63 bytes
// of anything, and the name goes into SQL, so it is held to a label's letters.
var pointNamePattern = regexp.MustCompile(fmt.Sprintf(`^[A-Za-z0-9][A-Za-z0-9._-]{0,%d}$`, pointNameMax-1))

// ValidPointName says whether a restore point may have that name.
func ValidPointName(name string) bool {
	return pointNamePattern.MatchString(name) && !strings.HasPrefix(name, reservedPointPrefix)
}

func (e *Engine) MarkRestorePoint(managed *entity.AppManagedSettings, name string) (
	managedmodel.MarkCommand, error) {
	if !ValidPointName(name) {
		return managedmodel.MarkCommand{}, hperrors.NewArgumentInvalid("name")
	}
	archiveTimeout := defaultArchiveTimeout
	if managed != nil && managed.Database != nil && managed.Database.ArchiveTimeout > 0 {
		archiveTimeout = managed.Database.ArchiveTimeout.ToDuration()
	}
	timeout := markWaits * archiveTimeout
	return managedmodel.MarkCommand{
		Command: fmt.Sprintf("%s/walg-mark %s %d", scriptDir, name, int(timeout/time.Second)),
		Timeout: timeout,
	}, nil
}

func (e *Engine) MarkResult(stdout string) (string, error) {
	for _, line := range strings.Split(stdout, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), "=")
		if found && key == markLSN && lsnPattern.MatchString(value) {
			return value, nil
		}
	}
	return "", hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("walg-mark printed no LSN")
}
