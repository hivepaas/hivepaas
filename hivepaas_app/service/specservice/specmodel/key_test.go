package specmodel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func cert(id, name, kind string, createdAt time.Time) *entity.Setting {
	return &entity.Setting{
		ID: id, Name: name, Kind: kind, Type: base.SettingTypeSSLCert,
		Scope: base.ObjectScopeGlobal, CreatedAt: createdAt,
	}
}

// The five certificates seeded into a real development installation. Three have
// unique names and must keep them untouched; only the genuinely ambiguous pair
// gains a suffix.
//
// This is also the case that rules out slugifying: SlugifyAsKey collapses
// "*.dev.hivepaas.com" and "dev.hivepaas.com" both to "dev_hivepaas_com", and
// "*.localhost" and "localhost" both to "localhost".
func TestDeriveSettingKeysOnRealCertificateNames(t *testing.T) {
	t0 := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 9, 15, 14, 27, 54, 0, time.UTC)

	keys := DeriveSettingKeys([]*entity.Setting{
		cert("01JAB9XED0GTXBSQDFVYAJ8WM2", "*.dev.hivepaas.com", "self-signed", t0),
		cert("01JAB9XED0GTXBSQDFVYAJ8WM1", "*.localhost", "letsencrypt", t0),
		cert("01JAB9XED0GTXBSQDFVYAJ8WM4", "dev.hivepaas.com", "self-signed", t0),
		cert("01JAB9XED0GTXBSQDFVYAJ8WM3", "localhost", "letsencrypt", t0),
		cert("01M2JQF72SWV5NCVATKD1WV536", "localhost", "self-signed", t1),
	})

	assert.Equal(t, "*.dev.hivepaas.com", keys["01JAB9XED0GTXBSQDFVYAJ8WM2"])
	assert.Equal(t, "*.localhost", keys["01JAB9XED0GTXBSQDFVYAJ8WM1"])
	assert.Equal(t, "dev.hivepaas.com", keys["01JAB9XED0GTXBSQDFVYAJ8WM4"])
	assert.Equal(t, "localhost@letsencrypt", keys["01JAB9XED0GTXBSQDFVYAJ8WM3"])
	assert.Equal(t, "localhost@self-signed", keys["01M2JQF72SWV5NCVATKD1WV536"])
}

// ssh-key, notification and basic-auth all carry an empty kind, so a duplicate
// name there cannot be broken by kind and falls through to an index.
func TestDeriveSettingKeysFallsThroughToIndexWhenKindIsEmpty(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mk := func(id, name string, at time.Time) *entity.Setting {
		return &entity.Setting{ID: id, Name: name, Type: base.SettingTypeSSHKey, CreatedAt: at}
	}

	keys := DeriveSettingKeys([]*entity.Setting{
		mk("k2", "deploy", t0.Add(time.Hour)),
		mk("k1", "deploy", t0),
		mk("k3", "deploy", t0.Add(2*time.Hour)),
	})

	// Ordered by created_at, so k1 is the unsuffixed one.
	assert.Equal(t, "deploy", keys["k1"])
	assert.Equal(t, "deploy#2", keys["k2"])
	assert.Equal(t, "deploy#3", keys["k3"])
}

// Two settings of different types never collide with each other.
func TestDeriveSettingKeysIsScopedByType(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	keys := DeriveSettingKeys([]*entity.Setting{
		{ID: "a", Name: "main", Type: base.SettingTypeSSHKey, CreatedAt: t0},
		{ID: "b", Name: "main", Type: base.SettingTypeSecret, CreatedAt: t0},
	})
	assert.Equal(t, "main", keys["a"])
	assert.Equal(t, "main", keys["b"])
}

// The same input must always produce the same keys, whatever order the database
// returned the rows in.
func TestDeriveSettingKeysIsDeterministic(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	a := &entity.Setting{ID: "z", Name: "dup", Type: base.SettingTypeSSHKey, CreatedAt: t0}
	b := &entity.Setting{ID: "y", Name: "dup", Type: base.SettingTypeSSHKey, CreatedAt: t0}

	forward := DeriveSettingKeys([]*entity.Setting{a, b})
	reversed := DeriveSettingKeys([]*entity.Setting{b, a})
	assert.Equal(t, forward, reversed)
	// Same created_at, so the id breaks the tie: "y" sorts before "z".
	assert.Equal(t, "dup", forward["y"])
	assert.Equal(t, "dup#2", forward["z"])
}

// The input slice must not be reordered underneath the caller.
func TestDeriveSettingKeysDoesNotMutateItsInput(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	input := []*entity.Setting{
		{ID: "z", Name: "b", Type: base.SettingTypeSSHKey, CreatedAt: t0.Add(time.Hour)},
		{ID: "y", Name: "a", Type: base.SettingTypeSSHKey, CreatedAt: t0},
	}
	DeriveSettingKeys(input)
	assert.Equal(t, "z", input[0].ID, "the caller's ordering must survive")
}
