package specmodel

import (
	"fmt"
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// keySeparator joins a type to a name when checking uniqueness. It is a NUL
// byte so that no name can forge a collision by containing the separator.
const keySeparator = "\x00"

// DeriveSettingKeys assigns every setting the key a spec refers to it by,
// returning setting id -> key.
//
// The name is used verbatim rather than slugified. Slugifying collapses
// distinctions that matter: SlugifyAsKey turns both "*.localhost" and
// "localhost" into "localhost", and both "*.dev.hivepaas.com" and
// "dev.hivepaas.com" into "dev_hivepaas_com" - four different certificates,
// two keys. Keys only ever appear inside YAML, which quotes whatever it has to.
//
// Duplicates are broken by kind first, because kind is already part of identity
// in this system - one default is allowed per (scope, type, kind) - and only
// then by an index.
//
// Ordering is by created_at then id so that two exports of unchanged data
// produce the same keys regardless of the order rows came back in.
func DeriveSettingKeys(settings []*entity.Setting) map[string]string {
	ordered := make([]*entity.Setting, len(settings))
	copy(ordered, settings)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
			return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
		}
		return ordered[i].ID < ordered[j].ID
	})

	nameCount := make(map[string]int, len(ordered))
	for _, s := range ordered {
		nameCount[string(s.Type)+keySeparator+s.Name]++
	}

	keys := make(map[string]string, len(ordered))
	taken := make(map[string]bool, len(ordered))

	for _, s := range ordered {
		candidate := s.Name
		if nameCount[string(s.Type)+keySeparator+s.Name] > 1 && s.Kind != "" {
			candidate = s.Name + "@" + s.Kind
		}

		key := candidate
		for i := 2; taken[string(s.Type)+keySeparator+key]; i++ {
			key = fmt.Sprintf("%s#%d", candidate, i)
		}

		taken[string(s.Type)+keySeparator+key] = true
		keys[s.ID] = key
	}
	return keys
}
