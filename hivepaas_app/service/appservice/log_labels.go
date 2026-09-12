package appservice

import (
	"maps"
	"sort"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// The container labels that identify an app in its own logs.
//
// With the json-file driver's `labels` option, the daemon copies these into an
// `attrs` block on every line it writes. The daemon writes them, not the app,
// which is what makes them safe to scope a log query by: nothing a container
// prints can change them, and ApplyUserLabels refuses any user-supplied label
// under the hivepaas. prefix, so container settings cannot either.
//
// They must be container labels. Swarm does not pass service labels down to
// containers, so a service label named in `labels` produces no attrs at all -
// measured, not assumed.
const (
	LabelLogAppID        = "hivepaas.app.id"
	LabelLogProjectID    = "hivepaas.project.id"
	LabelLogProjectEnvID = "hivepaas.projectEnv.id"

	// logDriverJSONFile is the only driver whose files the collector can read.
	// `local`, the default HivePaaS gives new apps, writes a binary format under
	// local-logs/ and no *-json.log at all.
	logDriverJSONFile = "json-file"
)

// WithAppLogLabels returns labels with the app's identity stamped in.
//
// It returns a new map rather than writing into the one given. A cloned app's
// ContainerSpec starts as a shallow copy of its source's, so writing in place
// would stamp the clone's identity onto the source app too.
func WithAppLogLabels(labels map[string]string, app *entity.App) map[string]string {
	out := make(map[string]string, len(labels)+3) //nolint:mnd // the three identity labels
	maps.Copy(out, labels)
	out[LabelLogAppID] = app.ID
	out[LabelLogProjectID] = app.ProjectID
	out[LabelLogProjectEnvID] = app.ProjectEnvID
	return out
}

// IsLogDriverCollectible reports whether a service with this driver writes logs
// the collector can read.
//
// No driver means the daemon's default, which is json-file on a stock daemon
// and on every node this was checked against. A node configured otherwise in
// daemon.json would make that assumption wrong for services that name no driver.
func IsLogDriverCollectible(d *swarm.Driver) bool {
	return d == nil || d.Name == "" || d.Name == logDriverJSONFile
}

// WithLogLabelsOption returns the driver with its `labels` option naming the
// identity labels, so the daemon copies them into every line.
//
// A driver the collector cannot read is returned unchanged: the option would do
// nothing for it. Any labels the operator already listed are kept. The result is
// a new value, for the same reason WithAppLogLabels returns a new map.
func WithLogLabelsOption(d *swarm.Driver) *swarm.Driver {
	if d == nil || d.Name != logDriverJSONFile {
		return d
	}

	names := map[string]bool{LabelLogAppID: true, LabelLogProjectID: true, LabelLogProjectEnvID: true}
	for _, n := range strings.Split(d.Options["labels"], ",") {
		if n = strings.TrimSpace(n); n != "" {
			names[n] = true
		}
	}
	list := make([]string, 0, len(names))
	for n := range names {
		list = append(list, n)
	}
	// Sorted so the same app always produces the same spec; an unstable one would
	// make every update look like a change and restart the app for nothing.
	sort.Strings(list)

	opts := make(map[string]string, len(d.Options)+1)
	maps.Copy(opts, d.Options)
	opts["labels"] = strings.Join(list, ",")
	return &swarm.Driver{Name: d.Name, Options: opts}
}
