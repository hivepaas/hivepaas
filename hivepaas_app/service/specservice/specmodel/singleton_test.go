package specmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// These two are the only types in a real installation whose settings carry no
// name at all. Treated as collections, their key would be the empty string.
func TestTypesWithNoNameAreSingletons(t *testing.T) {
	assert.True(t, IsSingletonType(base.SettingTypeEnvVar))
	assert.True(t, IsSingletonType(base.SettingTypeAppRouting))
}

func TestNamedCollectionTypesAreNotSingletons(t *testing.T) {
	for _, typ := range []base.SettingType{
		base.SettingTypeSSLCert,
		base.SettingTypeSecret,
		base.SettingTypeSSHKey,
		base.SettingTypeConfigFile,
		base.SettingTypeOAuth,
	} {
		assert.False(t, IsSingletonType(typ), "%v holds many per scope", typ)
	}
}

// Both call GetSingle, which is the usual singleton signal, but both pass extra
// filters: notification adds `is_default = TRUE`, sched-job adds a kind and a
// target. Both hold several rows per scope - the development installation has
// two notifications and four scheduled jobs in the global scope alone - so
// classifying either as a singleton would export one row and lose the rest.
func TestNotificationAndSchedJobAreCollections(t *testing.T) {
	assert.False(t, IsSingletonType(base.SettingTypeNotification))
	assert.False(t, IsSingletonType(base.SettingTypeSchedJob))
	assert.Equal(t, "notifications", CollectionBlockName(base.SettingTypeNotification))
	assert.Equal(t, "schedJobs", CollectionBlockName(base.SettingTypeSchedJob))
}

// Every setting type that has a parser must be classified exactly once. A type
// in neither map has no block name and would be refused at assembly; a type in
// both is ambiguous.
//
// The list comes from entity's parser registry rather than from these two maps,
// which is the whole point: iterating the maps would make this test unable to
// notice a type missing from both.
func TestEverySettingTypeIsClassifiedExactlyOnce(t *testing.T) {
	for _, typ := range entity.AllParsedSettingTypes() {
		inSingleton := SingletonBlockName(typ) != ""
		inCollection := CollectionBlockName(typ) != ""
		assert.True(t, inSingleton || inCollection, "%v has no block name", typ)
		assert.False(t, inSingleton && inCollection, "%v is classified twice", typ)
	}
}

func TestBlockNamesAreUnique(t *testing.T) {
	seen := map[string]base.SettingType{}
	for typ, name := range singletonBlockNames {
		prev, dup := seen[name]
		assert.False(t, dup, "block name %q used by both %v and %v", name, prev, typ)
		seen[name] = typ
	}
	for typ, name := range collectionBlockNames {
		prev, dup := seen[name]
		assert.False(t, dup, "block name %q used by both %v and %v", name, prev, typ)
		seen[name] = typ
	}
}
