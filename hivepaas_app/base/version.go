package base

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	VersionCodeV1 = "v000001" // Date: 2026-01-01
)

const CurrentVersion = VersionCodeV1

const StableVersionCode = VersionCodeV1

const BetaVersionCode = VersionCodeV1

// The repositories the app's and the agent's images are released to.
const (
	AppImageRepo   = "ghcr.io/hivepaas/hivepaas"
	AgentImageRepo = "ghcr.io/hivepaas/hivepaas-agent"
)

// BetaVersion is the beta release this binary is: release.json's, as it was
// built. One file says what a release runs; see compiledRelease.
var BetaVersion = compiledRelease(hivepaas.ReleaseJSON, "beta", nil)

// StableVersion is the stable release this binary is. There has been none yet:
// until release.json names one, the stable channel runs the beta's images under
// the version it has always had.
var StableVersion = compiledRelease(hivepaas.ReleaseJSON, "stable", &ReleaseInfo{
	ReleaseDate: timeutil.Date(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
	AppVersion:  "v0.1.0",
})

// compiledRelease is a channel's release in a release file, as this binary runs
// it. A channel the file does not name is unreleased, the beta's images under
// unreleased's version and date.
//
// The app's and the agent's images are named by the version: the file pins
// them by digest once they are built, which is after this binary is, and the
// updater reads them from the release info it fetches, never from here. Neither
// are the templates read from here.
//
// A file that cannot be read stops the binary: it would otherwise run images
// nobody chose.
func compiledRelease(data []byte, channel string, unreleased *ReleaseInfo) *ReleaseInfo {
	var file map[string]*ReleaseInfo
	if err := json.Unmarshal(data, &file); err != nil {
		panic(fmt.Sprintf("release.json: %v", err))
	}
	release := file[channel]
	if release == nil {
		beta := file["beta"]
		if unreleased == nil || beta == nil {
			panic("release.json names no " + channel + " release")
		}
		copied := *beta
		copied.ReleaseDate, copied.AppVersion = unreleased.ReleaseDate, unreleased.AppVersion
		copied.NotesURL = ""
		copied.FunctionRuntimes = maps.Clone(beta.FunctionRuntimes)
		copied.BlockMajorUpgrade = slices.Clone(beta.BlockMajorUpgrade)
		release = &copied
	}
	version := strings.TrimPrefix(release.AppVersion, "v")
	release.AppImage = AppImageRepo + ":" + version
	release.AgentImage = AgentImageRepo + ":" + version
	release.Templates = nil
	return release
}

// ReleaseInfo is what a release says to run: an entry of release.json, which an
// installation fetches to find out a newer one exists, and which the binary is
// built with (BetaVersion, StableVersion).
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
	// OBIImage is OBI's: the eBPF instrumentation each node's agent runs while
	// apps' routes and calls are on - not a service, a container the agent
	// starts. An agent runs the OBI of the release it is built with, so an
	// update moves OBI by moving the agent; a release that names none leaves
	// it on obi.DefaultImage.
	OBIImage string `json:"obiImage,omitempty"`
	// RegistryImage is the registry HivePaaS runs for itself, when it runs one.
	// The configuration HivePaaS writes for it is the configuration this version
	// of zot accepts, so a bump is the trigger to re-check that.
	RegistryImage string `json:"registryImage"`
	// PlaceholderImage is what a new app runs until its first deployment: an
	// image whose own command waits and stops on a signal, so the app is given
	// none, and the image set later runs its own.
	PlaceholderImage string `json:"placeholderImage"`
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
