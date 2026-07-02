package entityutil

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func SliceToIDMap[T entity.IDEntity, S ~[]T](entities S) map[string]T {
	result := make(map[string]T, len(entities))
	for _, ent := range entities {
		result[ent.GetID()] = ent
	}
	return result
}

func SliceToNameMap[T entity.NamedEntity, S ~[]T](entities S, caseSensitive bool) map[string]T {
	result := make(map[string]T, len(entities))
	for _, ent := range entities {
		if caseSensitive {
			result[ent.GetName()] = ent
		} else {
			result[strings.ToLower(ent.GetName())] = ent
		}
	}
	return result
}
