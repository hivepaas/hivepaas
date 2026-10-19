package postgres

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

var retentionNow = time.Date(2026, 10, 18, 2, 30, 0, 0, time.UTC)

// nightlies are a base backup each night at 02:00, the last one tonight's,
// oldest first: base_0...01 is nights-1 nights ago.
func nightlies(nights int) []managedmodel.BackupRef {
	out := make([]managedmodel.BackupRef, 0, nights)
	for i := range nights {
		at := time.Date(2026, 10, 18, 2, 0, 0, 0, time.UTC).AddDate(0, 0, i-nights+1)
		out = append(out, managedmodel.BackupRef{Name: backupName(i + 1), At: at, Bytes: 100})
	}
	return out
}

func backupName(n int) string {
	return fmt.Sprintf("base_%024X", n)
}

func point(name string, daysAgo int, keepDays int) *entity.ManagedRestorePoint {
	p := &entity.ManagedRestorePoint{Name: name, Position: "0/1", At: retentionNow.AddDate(0, 0, -daysAgo)}
	if keepDays != 0 {
		p.KeepUntil = retentionNow.AddDate(0, 0, keepDays)
	}
	return p
}

// The boundary is the older of the retainFull-th newest full backup and the
// full backup before the oldest point still kept (design §4).
func TestTheRetentionBoundary(t *testing.T) {
	backups := nightlies(10) // 10 nights; with 7 kept, base_..04 is the 7th newest
	for name, tc := range map[string]struct {
		backups []managedmodel.BackupRef
		points  []*entity.ManagedRestorePoint
		want    string
	}{
		"no points":                     {backups, nil, backupName(4)},
		"a kept point inside retention": {backups, []*entity.ManagedRestorePoint{point("p", 2, 30)}, backupName(4)},
		"a kept point older than it":    {backups, []*entity.ManagedRestorePoint{point("p", 8, 30)}, backupName(2)},
		"the oldest kept point decides": {backups, []*entity.ManagedRestorePoint{point("a", 8, 30),
			point("b", 9, 30)}, backupName(1)},
		"a lapsed point":   {backups, []*entity.ManagedRestorePoint{point("p", 8, -1)}, backupName(4)},
		"a point not kept": {backups, []*entity.ManagedRestorePoint{point("p", 8, 0)}, backupName(4)},
		"a point older than every backup": {backups, []*entity.ManagedRestorePoint{point("p", 30, 30)},
			backupName(4)},
		"fewer fulls than retainFull": {nightlies(5), nil, ""},
		"deltas are not counted": {append(nightlies(7), managedmodel.BackupRef{
			Name: backupName(99) + "_D_" + backupName(7)[5:], At: retentionNow}), nil, backupName(1)},
	} {
		assert.Equal(t, tc.want, walgkit.RetentionBoundary(profile, tc.backups, 7, tc.points, retentionNow), name)
	}
}

// The nightly run's second step deletes what is before the boundary, decided
// from the backups its first step listed; without a list, nothing.
func TestTheNightlyRetentionStep(t *testing.T) {
	managed := &entity.AppManagedSettings{Database: &entity.ManagedDatabase{RetainFull: 7,
		RestorePoints: []*entity.ManagedRestorePoint{point("p", 8, 30)}}}
	listed := ""
	for _, b := range nightlies(10) {
		listed += b.Name + "@" + b.At.Format(time.RFC3339) + "@100 "
	}

	retention, err := New().Retention(&managedmodel.RetentionReq{Managed: managed,
		Outputs: map[string]string{walgkit.OutBackups: listed}, Now: retentionNow})

	assert.NoError(t, err)
	assert.Equal(t, &managedmodel.Retention{Args: "retain " + backupName(2), Boundary: backupName(2)}, retention)
	assert.Equal(t, "backup", New().NightlyBackupArgs())

	retention, err = New().Retention(&managedmodel.RetentionReq{Managed: managed, Now: retentionNow})
	assert.NoError(t, err)
	assert.Equal(t, &managedmodel.Retention{Args: "retain"}, retention, "nothing listed, nothing deleted")
}

// A point can be restored while the base backup before it is kept: until the
// nightly run that leaves it out - or, while a point at or before it is kept,
// until the run after that point's date.
func TestUntilWhenAPointCanBeRestored(t *testing.T) {
	daily := &entity.SchedJobSchedule{CronExpr: "0 2 * * *", InitialTime: time.Date(2026, 1, 1, 0, 0, 0, 0,
		time.UTC)}
	backups := nightlies(7)
	recent := point("recent", 1, 0)     // its backup is last night's, the 2nd newest
	oldest := point("oldest", 6, 0)     // its backup is the 7th newest: gone at the next run
	ancient := point("ancient", 30, 30) // before every backup kept
	until := func(points []*managedmodel.RestorePointStatus, name string) *time.Time {
		for _, p := range points {
			if p.Point.Name == name {
				return p.RestorableUntil
			}
		}
		return nil
	}

	points := walgkit.PointsStatus(profile, []*entity.ManagedRestorePoint{recent, oldest, ancient}, backups, 7, daily,
		retentionNow)

	assert.Equal(t, time.Date(2026, 10, 24, 2, 0, 0, 0, time.UTC), *until(points, "recent"), "six more nights")
	assert.Equal(t, time.Date(2026, 10, 19, 2, 0, 0, 0, time.UTC), *until(points, "oldest"), "the next night")
	assert.Nil(t, until(points, "ancient"), "no backup kept reaches it")
	assert.Equal(t, []string{"ancient", "oldest", "recent"}, []string{points[0].Point.Name,
		points[1].Point.Name, points[2].Point.Name}, "oldest first")

	kept := point("kept", 6, 20) // beside the oldest, kept 20 days
	points = walgkit.PointsStatus(profile, []*entity.ManagedRestorePoint{recent, oldest, kept}, backups, 7, daily,
		retentionNow)

	afterKept := time.Date(2026, 11, 8, 2, 0, 0, 0, time.UTC)
	assert.Equal(t, afterKept, *until(points, "kept"), "the first run after its date")
	assert.Equal(t, afterKept, *until(points, "oldest"), "held by the kept point beside it")
	assert.Equal(t, afterKept, *until(points, "recent"), "and so is everything after it")
	for _, p := range points {
		assert.Equal(t, p.Point.Name == "kept", p.Kept, p.Point.Name)
	}
}

func TestWhatKeptPointsHoldBack(t *testing.T) {
	backups := nightlies(10)

	held := walgkit.HeldBack(profile, backups, 7, []*entity.ManagedRestorePoint{point("p", 8, 30)}, retentionNow)

	assert.Equal(t, &managedmodel.HeldBack{Backups: 2, Bytes: 200}, held)
	assert.Nil(t, walgkit.HeldBack(profile, backups, 7, nil, retentionNow), "retention's own keep nothing back")
}

func TestARestorePointsName(t *testing.T) {
	for name, valid := range map[string]bool{
		"before-v42":            true,
		"release_2026.10.18":    true,
		strings.Repeat("a", 63): true,
		strings.Repeat("a", 64): false, // Postgres keeps 63 bytes
		"hp-test-1":             false, // HivePaaS's own
		"-leading":              false,
		"with space":            false,
		"quote'd":               false, // it goes into SQL
		"":                      false,
	} {
		assert.Equal(t, valid, ValidPointName(name), name)
	}
}

// A point is listed once the WAL it is in is archived: walg-mark waits twice the
// archive timeout for it, and prints its LSN.
func TestMarkingARestorePoint(t *testing.T) {
	managed := &entity.AppManagedSettings{Database: &entity.ManagedDatabase{
		ArchiveTimeout: timeutil.Duration(90 * time.Second)}}

	mark, err := New().MarkRestorePoint(managed, "before-v42")

	assert.NoError(t, err)
	assert.Equal(t, managedmodel.MarkCommand{Command: walgkit.ScriptDir + "/walg-mark before-v42 180",
		Timeout: 3 * time.Minute}, mark)
	_, err = New().MarkRestorePoint(managed, "x'; DROP TABLE t; --")
	assert.Error(t, err)

	lsn, err := New().MarkResult("Switching WAL...\nLSN=0/3000060\nSEGMENT=000000010000000000000003\n")
	assert.NoError(t, err)
	assert.Equal(t, "0/3000060", lsn)
	_, err = New().MarkResult("LSN=not-an-lsn\n")
	assert.Error(t, err)
}

// A deployment's point is stale once no base backup kept starts before it; a
// person's never is, and nothing is stale when the run listed no backups.
func TestStaleDeployPoints(t *testing.T) {
	deploy := func(name string, daysAgo int) *entity.ManagedRestorePoint {
		p := point(name, daysAgo, 0)
		p.Origin = entity.ManagedPointOriginDeploy
		return p
	}
	managed := &entity.AppManagedSettings{Database: &entity.ManagedDatabase{RetainFull: 3,
		RestorePoints: []*entity.ManagedRestorePoint{
			deploy("deploy-old", 5), deploy("deploy-new", 1), point("person-old", 5, 0)}}}
	listed := ""
	for _, b := range nightlies(7)[4:] { // what retention kept: the three newest
		listed += b.Name + "@" + b.At.Format(time.RFC3339) + "@100 "
	}

	assert.Equal(t, []string{"deploy-old"},
		New().StaleDeployPoints(managed, map[string]string{walgkit.OutBackups: listed}))
	assert.Nil(t, New().StaleDeployPoints(managed, map[string]string{}), "no list, nothing stale")
	assert.Nil(t, New().StaleDeployPoints(&entity.AppManagedSettings{}, map[string]string{walgkit.OutBackups: listed}))
}
