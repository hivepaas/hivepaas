package mariadb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// A point is marked by walg-mark, which closes and archives the binary log it
// is in: it waits two archive timeouts for that, as Postgres's does.
func TestMarkingARestorePoint(t *testing.T) {
	managed := &entity.AppManagedSettings{Engine: "mariadb", EngineAPI: 1,
		Database: &entity.ManagedDatabase{ArchiveTimeout: timeutil.Duration(90 * time.Second)}}

	command, err := New().MarkRestorePoint(managed, "before-v42")

	assert.NoError(t, err)
	assert.Equal(t, walgkit.ScriptDir+"/walg-mark before-v42 180", command.Command)
	assert.Equal(t, 3*time.Minute, command.Timeout)

	for _, name := range []string{"", "a;b", "hp-test-1", "-x", string(make([]byte, 64))} {
		_, err = New().MarkRestorePoint(managed, name)
		assert.Error(t, err, "%q", name)
	}
}

// What walg-mark prints is read back as the point's position: its GTID, which a
// restore stops at whatever server's binary logs the archive holds by then.
func TestTheMarkPrintsThePosition(t *testing.T) {
	for stdout, want := range map[string]string{
		"Flushed.\nPOSITION=mariadb-bin.000007:1234\nGTID=0-1-58\n": "0-1-58",
		"POSITION=mariadb-bin.000007:1234\nGTID=0-1-58,7-2-1094\n":  "0-1-58,7-2-1094",
		"GTID=0-2-9\nPOSITION=mariadb-bin-db-r1.000002:340\n":       "0-2-9",
		// No transaction yet, or a GTID it cannot be: the file and offset.
		"POSITION=mariadb-bin.000001:328\nGTID=\n":            "mariadb-bin.000001:328",
		"POSITION=mariadb-bin.000001:328\nGTID=0-1-1; drop\n": "mariadb-bin.000001:328",
	} {
		position, err := New().MarkResult(stdout)
		assert.NoError(t, err, "%q", stdout)
		assert.Equal(t, want, position, "%q", stdout)
	}

	for _, stdout := range []string{"", "GTID=0-1-x\n", "POSITION=0/3000060\n", "POSITION=x.000001:1; drop\n"} {
		_, err := New().MarkResult(stdout)
		assert.Error(t, err, "%q", stdout)
	}
}
