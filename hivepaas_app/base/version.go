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

var StableVersion = &ReleaseInfo{
	ReleaseDate:       timeutil.Date(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
	AppVersion:        "v0.1.0",
	AppImage:          "hivepaas/hivepaas-dev:0.1.0",
	AgentImage:        "hivepaas/hivepaas-agent-dev:0.1.0",
	RedisImage:        "redis:8.6-alpine",
	DbImage:           "postgres:18.3-alpine",
	TraefikImage:      "traefik:v3.7",
	VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
	VlagentImage:      "victoriametrics/vlagent:v1.52.0",
	RegistryImage:     "ghcr.io/project-zot/zot:v2.1.21",
	FunctionRuntimes:  functionRuntimesV1(),

	BlockMajorUpgrade: []string{HivepaasDbKey},

	Templates: &TemplatesRef{
		Repo:        "hivepaas/app-templates",
		Commit:      "b495ab97464767c5c936a1bb4a8381fb680047f5",
		IndexSHA256: "89cbe671b6ddf01d544eee20875d54a328e6e9f69d9c418a7fb9eb9cd8db0b13",
	},
}

const BetaVersionCode = VersionCodeV1

// TODO: update these info later
var BetaVersion = &ReleaseInfo{
	ReleaseDate:       timeutil.Date(time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)),
	AppVersion:        "v1.0.0-beta1",
	AppImage:          "hivepaas/hivepaas-dev:0.1.0",
	AgentImage:        "hivepaas/hivepaas-agent-dev:0.1.0",
	RedisImage:        "redis:8.6-alpine",
	DbImage:           "postgres:18.3-alpine",
	TraefikImage:      "traefik:v3.7",
	VictoriaLogsImage: "victoriametrics/victoria-logs:v1.52.0",
	VlagentImage:      "victoriametrics/vlagent:v1.52.0",
	RegistryImage:     "ghcr.io/project-zot/zot:v2.1.21",
	FunctionRuntimes:  functionRuntimesV1(),

	BlockMajorUpgrade: []string{HivepaasDbKey},
}

// functionRuntimesV1 are the images of function-runtimes v1.2.0, a release of
// the function contract v1: hivepaas-runtime call, which hands a scheduled
// call to the function's serve.
func functionRuntimesV1() map[string]string {
	return map[string]string{
		"node24": "ghcr.io/hivepaas/function-runtime-node24:1.2.0" +
			"@sha256:dd905254ef4cd7b88bb73d37f75fcaa9b700c3ed521bcb8e3eac424e2229d779",
		"bun1": "ghcr.io/hivepaas/function-runtime-bun1:1.2.0" +
			"@sha256:f950403f2c15f9d353f9567799f0aec3b9065d67eb7dfdf43690fc8e9388ddeb",
		"python313": "ghcr.io/hivepaas/function-runtime-python313:1.2.0" +
			"@sha256:db4db45c8f28701d33ae7ec60c4f334bc6b25b8d733639d5dbc5584f309ce6ea",
		"go127": "ghcr.io/hivepaas/function-runtime-go127:1.2.0" +
			"@sha256:44f0507bbead6dfe0bbb4ca9511dc0354968ed86985d6787916887c2326cb946",
		"go127-build": "ghcr.io/hivepaas/function-runtime-go127-build:1.2.0" +
			"@sha256:0dcdffd4d417296b4848f58442aab018a8a96ca5080ff052b5a7d6540b59c9eb",
	}
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
	ReleaseDate timeutil.Date `json:"releaseDate"`
	AppVersion  string        `json:"appVersion"`
	AppImage    string        `json:"appImage"`
	// AgentImage is the agent's, the service on every node the app works
	// through. It is released with the app and moves with it; a release that
	// names none leaves the agent as it is.
	AgentImage        string `json:"agentImage,omitempty"`
	RedisImage        string `json:"redisImage"`
	DbImage           string `json:"dbImage"`
	TraefikImage      string `json:"traefikImage"`
	VictoriaLogsImage string `json:"victoriaLogsImage"`
	VlagentImage      string `json:"vlagentImage"`
	// RegistryImage is the registry HivePaaS runs for itself, when it runs one.
	// The configuration HivePaaS writes for it is the configuration this version
	// of zot accepts, so a bump is the trigger to re-check that.
	RegistryImage string `json:"registryImage"`
	// FunctionRuntimes are the images functions are built on, by runtime
	// (node24, go127; go127-build for the image Go functions compile in). A
	// function built under a release uses the image this release names, and
	// keeps it until it is deployed again. The images are released from the
	// repository hivepaas/function-runtimes.
	FunctionRuntimes map[string]string `json:"functionRuntimes,omitempty"`

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

	// NotesURL is where the release's notes are read: what changed, and what to
	// know before moving to it. Absent means the release published none.
	NotesURL string `json:"notesUrl,omitempty"`

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
