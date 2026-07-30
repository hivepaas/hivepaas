package envvarhelper

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

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
