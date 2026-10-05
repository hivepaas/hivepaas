package backupsnapshotuc

import (
	"time"

	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// snapshotFilter is what a person narrows a view to.
type snapshotFilter struct {
	ID      string
	RepoIDs []string
	AppIDs  []string
	// Tags are key:value, all of which a snapshot must carry.
	Tags []string
	From time.Time
	// Before is the end of the range, left out.
	Before time.Time
	Search string
}

const (
	// hasTagIn is a snapshot carrying one of the given tags.
	hasTagIn = "EXISTS (SELECT 1 FROM tags AS t WHERE t.object_id = setting.id AND t.deleted_at IS NULL " +
		"AND t.tag IN (?))"
	hasTag = "EXISTS (SELECT 1 FROM tags AS t WHERE t.object_id = setting.id AND t.deleted_at IS NULL " +
		"AND t.tag = ?)"
	// ownedByNoLiveApp is a snapshot no app that still exists owns.
	ownedByNoLiveApp = "setting.ref_id IN (?) AND NOT EXISTS (SELECT 1 FROM tags AS t " +
		"JOIN apps AS a ON t.tag = '" + entity.DataBackupTagApp + ":' || a.id " +
		"WHERE t.object_id = setting.id AND t.deleted_at IS NULL AND a.deleted_at IS NULL)"
	// snapshotTime is when a snapshot was taken: its setting's created_at
	// (entity.BackupSnapshot), which idx_settings_backup_snapshot_time orders
	// by repository - the data's time, a text, could not be indexed so.
	snapshotTime = "setting.created_at"
	nothing      = "1=0"
)

// snapshotQueryOpts lists the snapshots of a view's reach, narrowed by the
// filter, newest first. The filter only narrows: a repository or an app outside
// the reach lists nothing.
func snapshotQueryOpts(reach *snapshotReach, filter *snapshotFilter) []bunex.SelectQueryOption {
	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.type = ?", base.SettingTypeBackupSnapshot),
		// An expression: SelectOrder would quote it as a column name.
		func(q *bun.SelectQuery) *bun.SelectQuery { return q.OrderExpr(snapshotTime + " DESC") },
	}

	repoIDs := reach.repoIDs
	if len(filter.RepoIDs) > 0 {
		repoIDs = gofn.Intersection(repoIDs, filter.RepoIDs)
	}
	if len(repoIDs) == 0 || (len(reach.appIDs) == 0 && len(reach.ownerRepoIDs) == 0) {
		return append(opts, bunex.SelectWhere(nothing))
	}
	opts = append(opts, bunex.SelectWhere("setting.ref_id IN (?)", bunex.List(repoIDs)))

	var owned []bunex.SelectQueryOption
	if len(reach.appIDs) > 0 {
		owned = append(owned, bunex.SelectWhereOr(hasTagIn, bunex.List(appTags(reach.appIDs))))
	}
	if len(reach.ownerRepoIDs) > 0 {
		owned = append(owned, bunex.SelectWhereOr(ownedByNoLiveApp, bunex.List(reach.ownerRepoIDs)))
	}
	opts = append(opts, bunex.SelectWhereGroup(owned...))

	if filter.ID != "" {
		opts = append(opts, bunex.SelectWhere("setting.id = ?", filter.ID))
	}
	if len(filter.AppIDs) > 0 {
		opts = append(opts, bunex.SelectWhere(hasTagIn, bunex.List(appTags(filter.AppIDs))))
	}
	for _, tag := range filter.Tags {
		opts = append(opts, bunex.SelectWhere(hasTag, tag))
	}
	if !filter.From.IsZero() {
		opts = append(opts, bunex.SelectWhere(snapshotTime+" >= ?", filter.From))
	}
	if !filter.Before.IsZero() {
		opts = append(opts, bunex.SelectWhere(snapshotTime+" < ?", filter.Before))
	}
	if filter.Search != "" {
		keyword := bunex.MakeLikeOpStr(filter.Search, true)
		opts = append(opts, bunex.SelectWhereGroup(
			bunex.SelectWhere("setting.name ILIKE ?", keyword),
			bunex.SelectWhereOr("setting.data->>'description' ILIKE ?", keyword),
		))
	}
	return opts
}

func appTags(appIDs []string) []string {
	tags := make([]string, 0, len(appIDs))
	for _, id := range appIDs {
		tags = append(tags, entity.DataBackupTagApp+":"+id)
	}
	return tags
}
