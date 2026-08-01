package envvarhelper

import (
	"sort"

	"github.com/cespare/xxhash/v2"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func CalcHashes(vars []*entity.EnvVar) (runtimeHash, sharedHash, buildTimeHash uint64) {
	if len(vars) == 0 {
		return 0, 0, 0
	}

	// 1. Clone & sort input by var key (name)
	sortedVars := make([]*entity.EnvVar, len(vars))
	copy(sortedVars, vars)
	sort.Slice(sortedVars, func(i, j int) bool {
		if sortedVars[i] == nil || sortedVars[j] == nil {
			return sortedVars[i] != nil
		}
		return sortedVars[i].Key < sortedVars[j].Key
	})

	runtimeHasher := xxhash.New()
	sharedHasher := xxhash.New()
	buildTimeHasher := xxhash.New()

	var runtimeCount, sharedCount, buildCount int

	// 2. Loop vars and calculate 3 hash types
	for _, v := range sortedVars {
		if v == nil {
			continue
		}

		entry := v.Key + "=" + v.Value + "\n"

		// Runtime vars: !v.IsBuild && !v.IsShared
		if !v.IsBuild && !v.IsShared {
			_, _ = runtimeHasher.WriteString(entry)
			runtimeCount++
		}

		// Shared vars: v.IsShared
		if v.IsShared {
			_, _ = sharedHasher.WriteString(entry)
			sharedCount++
		}

		// Build-time vars: v.IsBuild
		if v.IsBuild {
			_, _ = buildTimeHasher.WriteString(entry)
			buildCount++
		}
	}

	if runtimeCount > 0 {
		runtimeHash = runtimeHasher.Sum64()
	}
	if sharedCount > 0 {
		sharedHash = sharedHasher.Sum64()
	}
	if buildCount > 0 {
		buildTimeHash = buildTimeHasher.Sum64()
	}
	return runtimeHash, sharedHash, buildTimeHash
}
