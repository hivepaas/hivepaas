package sysupdateserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// The major has to appear as a path segment of its own. Rewriting anything else
// would point postgres at a directory nobody chose.
func TestRepointPGDATAMovesTheMajorSegment(t *testing.T) {
	got, err := repointPGDATA("/var/lib/postgresql/18/docker", 18, 19)

	assert.NoError(t, err)
	assert.Equal(t, "/var/lib/postgresql/19/docker", got)
}

func TestRepointPGDATARefusesAPathWithoutTheMajor(t *testing.T) {
	_, err := repointPGDATA("/var/lib/postgresql/data", 18, 19)

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

func TestRepointPGDATARefusesWhenTheServiceSetsNone(t *testing.T) {
	_, err := repointPGDATA("", 18, 19)

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

// `18` inside a longer segment is not the major, and rewriting it would build a
// path to nowhere.
func TestRepointPGDATALeavesALookalikeSegmentAlone(t *testing.T) {
	_, err := repointPGDATA("/var/lib/postgresql/pg18data/docker", 18, 19)

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

func TestClientMajorVersionReadsPgRestoreOutput(t *testing.T) {
	got, ok := clientMajorVersion("pg_restore (PostgreSQL) 18.6\n")
	assert.True(t, ok)
	assert.Equal(t, 18, got)

	got, ok = clientMajorVersion("pg_restore (PostgreSQL) 19beta2")
	assert.True(t, ok)
	assert.Equal(t, 19, got)

	_, ok = clientMajorVersion("")
	assert.False(t, ok)
}

func TestPGDataEnvIsReadAndReplacedInPlace(t *testing.T) {
	env := []string{"POSTGRES_DB=hivepaas", "PGDATA=/var/lib/postgresql/18/docker"}

	assert.Equal(t, "/var/lib/postgresql/18/docker", pgDataEnv(env))

	env = setPGDataEnv(env, "/var/lib/postgresql/19/docker")
	assert.Equal(t, []string{"POSTGRES_DB=hivepaas", "PGDATA=/var/lib/postgresql/19/docker"}, env)

	// A service that never set it gets it appended rather than losing the change.
	env = setPGDataEnv([]string{"POSTGRES_DB=hivepaas"}, "/x")
	assert.Equal(t, []string{"POSTGRES_DB=hivepaas", "PGDATA=/x"}, env)
}

// The new cluster goes on a volume of its own. Suffixes do not accumulate: an
// install already upgraded once moves from _19 to _20, not to _19_20.
func TestRepointDbVolumeNamesTheVolumeForTheMajor(t *testing.T) {
	assert.Equal(t, "hivepaas_db_19", repointDbVolume("hivepaas_db", 19))
	assert.Equal(t, "hivepaas_db_20", repointDbVolume("hivepaas_db_19", 20))
	// A name whose last segment is not a number keeps all of itself.
	assert.Equal(t, "pg_data_19", repointDbVolume("pg_data", 19))
}

func TestDbVolumeNameReadsTheMountAtTheDataPath(t *testing.T) {
	got, err := dbVolumeName([]mount.Mount{
		{Type: mount.TypeBind, Source: "/etc/thing", Target: "/etc/thing"},
		{Type: mount.TypeVolume, Source: "hivepaas_db", Target: dbVolumeMountTarget},
	})

	assert.NoError(t, err)
	assert.Equal(t, "hivepaas_db", got)
}

// An upgrade moves the database onto a volume of its own, and there is no
// equivalent of that for a host path somebody chose.
func TestDbVolumeNameRefusesABindMount(t *testing.T) {
	_, err := dbVolumeName([]mount.Mount{
		{Type: mount.TypeBind, Source: "/srv/pgdata", Target: dbVolumeMountTarget},
	})

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

func TestDbVolumeNameRefusesAServiceWithNoDataMount(t *testing.T) {
	_, err := dbVolumeName([]mount.Mount{{Type: mount.TypeVolume, Source: "x", Target: "/elsewhere"}})

	assert.ErrorIs(t, err, hperrors.ErrUnsupported)
}

func TestSetDbVolumeNameLeavesOtherMountsAlone(t *testing.T) {
	mounts := []mount.Mount{
		{Type: mount.TypeBind, Source: "/etc/thing", Target: "/etc/thing"},
		{Type: mount.TypeVolume, Source: "hivepaas_db", Target: dbVolumeMountTarget},
	}

	setDbVolumeName(mounts, "hivepaas_db_19")

	assert.Equal(t, "/etc/thing", mounts[0].Source)
	assert.Equal(t, "hivepaas_db_19", mounts[1].Source)
}
