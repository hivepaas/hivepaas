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

// What walg-mark prints is read back as the point's position: the binary log
// and the offset of its end, which a restore replays to.
func TestTheMarkPrintsThePosition(t *testing.T) {
	position, err := New().MarkResult("Flushed.\nPOSITION=mariadb-bin.000007:1234\nGTID=0-1-58\n")
	assert.NoError(t, err)
	assert.Equal(t, "mariadb-bin.000007:1234", position)

	for _, stdout := range []string{"", "GTID=0-1-58\n", "POSITION=0/3000060\n", "POSITION=x.000001:1; drop\n"} {
		_, err = New().MarkResult(stdout)
		assert.Error(t, err, "%q", stdout)
	}
}
