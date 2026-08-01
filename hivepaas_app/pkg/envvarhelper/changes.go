package envvarhelper

import (
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func Equal(ev1, ev2 []*entity.EnvVar) bool {
	if len(ev1) != len(ev2) {
		return false
	}
	for i := range ev1 {
		if !ev1[i].Equal(ev2[i]) {
			return false
		}
	}
	return true
}

func CalcContentChanges(ev1, ev2 []*entity.EnvVar) (runtimeChange, sharedChange, buildChange bool) {
	s1 := sortEnvVars(ev1)
	s2 := sortEnvVars(ev2)

	r1, sh1, b1 := filterEnvVars(s1)
	r2, sh2, b2 := filterEnvVars(s2)

	return !Equal(r1, r2), !Equal(sh1, sh2), !Equal(b1, b2)
}

func sortEnvVars(vars []*entity.EnvVar) []*entity.EnvVar {
	if len(vars) == 0 {
		return nil
	}
	sorted := make([]*entity.EnvVar, len(vars))
	copy(sorted, vars)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i] == nil || sorted[j] == nil {
			return sorted[i] != nil
		}
		return sorted[i].Key < sorted[j].Key
	})
	return sorted
}

func filterEnvVars(vars []*entity.EnvVar) (runtime, shared, build []*entity.EnvVar) {
	for _, v := range vars {
		if v == nil {
			continue
		}
		if !v.IsBuild {
			runtime = append(runtime, v)
		}
		if v.IsShared {
			shared = append(shared, v)
		}
		if v.IsBuild {
			build = append(build, v)
		}
	}
	return runtime, shared, build
}
