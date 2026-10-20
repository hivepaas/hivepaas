package walgkit

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
	// markWaits is how many archive timeouts walg-mark waits for the log the
	// point is in to be archived: it asks for that at once, and the archive
	// timeout bounds how late it is.
	markWaits = 2
	// reservedPointPrefix is HivePaaS's own points', such as the restore test's.
	reservedPointPrefix = "hp-"
	// PointNameMax is the longest point name the engines take, in bytes:
	// Postgres's, which keeps the name.
	PointNameMax = 63
)

// pointNamePattern is what a point's name may be: it reaches walg-mark's
// arguments and, for Postgres, SQL, so it is held to a label's letters.
var pointNamePattern = regexp.MustCompile(fmt.Sprintf(`^[A-Za-z0-9][A-Za-z0-9._-]{0,%d}$`, PointNameMax-1))

// ValidPointName says whether a restore point may have that name.
func ValidPointName(name string) bool {
	return pointNamePattern.MatchString(name) && !strings.HasPrefix(name, reservedPointPrefix)
}

// MarkRestorePoint is walg-mark's command for a point of that name (design §4,
// "Restore points"): its arguments are the name and how many seconds to wait
// for the log the point is in to be archived.
func MarkRestorePoint(managed *entity.AppManagedSettings, name string) (managedmodel.MarkCommand, error) {
	if !ValidPointName(name) {
		return managedmodel.MarkCommand{}, hperrors.NewArgumentInvalid("name")
	}
	archiveTimeout := DefaultArchiveTimeout
	if managed != nil && managed.Database != nil && managed.Database.ArchiveTimeout > 0 {
		archiveTimeout = managed.Database.ArchiveTimeout.ToDuration()
	}
	timeout := markWaits * archiveTimeout
	return managedmodel.MarkCommand{
		Command: fmt.Sprintf("%s/walg-mark %s %d", ScriptDir, name, int(timeout/time.Second)),
		Timeout: timeout,
	}, nil
}

// MarkResult is the value of key that walg-mark printed, as KEY=value, when it
// is a position as pattern says one is.
func MarkResult(stdout, key string, pattern *regexp.Regexp) (string, error) {
	for _, line := range strings.Split(stdout, "\n") {
		k, value, found := strings.Cut(strings.TrimSpace(line), "=")
		if found && k == key && pattern.MatchString(value) {
			return value, nil
		}
	}
	return "", hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("walg-mark printed no %s", key)
}
