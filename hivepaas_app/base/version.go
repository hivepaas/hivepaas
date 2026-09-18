package base

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	VersionCodeV1 = "v000001" // Date: 2026-01-01
)

const CurrentVersion = VersionCodeV1

const StableVersionCode = VersionCodeV1

// TODO: update these info later
var StableVersion = &ReleaseInfo{
	ReleaseDate:       timeutil.Date(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
	AppVersion:        "v0.1.0",
	AppImage:          "hivepaas/hivepaas-dev:0.1.0",
	RedisImage:        "redis:8.6-alpine",
	DbImage:           "postgres:18.3-alpine",
	TraefikImage:      "traefik:v3.7",
	VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
	VlagentImage:      "victoriametrics/vlagent:v1.52.0",

	BlockMajorUpgrade: []string{HivepaasDbKey},

	Templates: &TemplatesRef{
		Repo:        "hivepaas/app-templates",
		Commit:      "814917c4414b73cfe742904dd779b871a086e7d1",
		IndexSHA256: "375f4ae6c03d4a771e261aeef6b07798dd8d0e0db8e3a6ebf269ab7ff51e8604",
	},
}

const BetaVersionCode = VersionCodeV1

// TODO: update these info later
var BetaVersion = &ReleaseInfo{
	ReleaseDate:       timeutil.Date(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
	AppVersion:        "v0.1.0-beta1",
	AppImage:          "hivepaas/hivepaas-dev:0.1.0",
	RedisImage:        "redis:8.6-alpine",
	DbImage:           "postgres:18.3-alpine",
	TraefikImage:      "traefik:v3.7",
	VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
	VlagentImage:      "victoriametrics/vlagent:v1.52.0",

	BlockMajorUpgrade: []string{HivepaasDbKey},
}

// ReleaseInfo is what a release says to run. It is mirrored by release.json,
// which is the copy an installation fetches to find out a newer one exists; the
// values below are the copy compiled into the binary, and the two have to be
// changed together.
//
// The logging images are two, not one: the backend and the collector are
// separate upstream repositories. They release in step today and still get a
// field each, because a version derived from the other's tag would be wrong in
// silence on the first release where they do not.
type ReleaseInfo struct {
	ReleaseDate       timeutil.Date `json:"releaseDate"`
	AppVersion        string        `json:"appVersion"`
	AppImage          string        `json:"appImage"`
	RedisImage        string        `json:"redisImage"`
	DbImage           string        `json:"dbImage"`
	TraefikImage      string        `json:"traefikImage"`
	VictoriaLogsImage string        `json:"victoriaLogsImage"`
	VlagentImage      string        `json:"vlagentImage"`

	// BlockMajorUpgrade names the components whose image may not cross a major
	// version in this release, by the same keys as HivepaasDbKey and friends.
	//
	// Every component can be listed; the default is that none are. Crossing a
	// major only matters where a component owns durable state it might convert,
	// and where that conversion is not something this update knows how to do or
	// to undo - swarm's rollback restores the image, never the data underneath.
	//
	// It is declared by the release rather than compiled in because the release
	// is where the answer is actually known. By the time a release names a new
	// major, whoever cut it has read the upstream notes; nobody earlier could.
	//
	// `db` is listed because postgres will not start on a data directory written
	// by a different major - it needs pg_upgrade or a dump and reload, neither of
	// which the updater performs. Removing it from the list is how that work,
	// when it exists, is switched on.
	BlockMajorUpgrade []string `json:"blockMajorUpgrade,omitempty"`

	// Templates pins the app templates this release offers. It is only ever read
	// from release info, never from the copy compiled in: templates are released
	// on their own schedule, and the binary has no use for a pin it cannot fetch.
	// Absent means the release offers no templates.
	Templates *TemplatesRef `json:"templates,omitempty"`
}

// TemplatesRef pins a revision of the app templates repository.
//
// A commit alone does not pin content for us: the files are fetched over HTTPS
// from GitHub, and nothing on this side can check a git object id against what
// comes back. The pin is therefore a chain of sha256 hashes, rooted in the signed
// release info. IndexSHA256 is the hash of the repository's index.json at
// Commit, and the index lists every template file with its own sha256. A file
// whose hash does not match the index, or an index whose hash does not match
// this, is refused.
type TemplatesRef struct {
	// Repo is the GitHub repository, as owner/name.
	Repo string `json:"repo"`
	// Commit is the full 40-character commit sha the files are read at.
	Commit string `json:"commit"`
	// IndexSHA256 is the hex sha256 of index.json at Commit.
	IndexSHA256 string `json:"indexSha256"`
}
