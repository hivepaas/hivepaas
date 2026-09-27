package taskschedjobexec

import (
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// dataBackupOutputs is what a data backup step tells the steps after it.
func dataBackupOutputs(result *entity.SchedJobDataBackupResult) map[string]string {
	if result == nil {
		return nil
	}
	return map[string]string{
		"SNAPSHOT_ID":         result.SnapshotID,
		"SNAPSHOT_SIZE_BYTES": strconv.FormatInt(result.SizeBytes, 10),
	}
}
