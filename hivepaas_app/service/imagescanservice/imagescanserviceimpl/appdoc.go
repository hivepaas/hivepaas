package imagescanserviceimpl

import (
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

const (
	// defaultTrivyImage is what runs when the release names no Trivy image - a
	// release file from before the field existed.
	defaultTrivyImage = "aquasec/trivy:0.75.0"

	// scanEnv is the hivepaas project's environment both apps live in: its
	// network is theirs alone, and the runner reaches the server on it by the
	// server's key.
	scanEnv = "scanning"

	serverAppName = "Trivy server"
	serverAppNote = "Holds the vulnerability DB image scanning matches against. It is configured in system settings."
	runnerAppName = "Trivy runner"
	runnerAppNote = "Scans the images of the node it runs on, one task on every node. It is configured in " +
		"system settings."

	serverPort = 4954

	// tokenKey is the secret both apps hold the server's token in, and tokenPath
	// where it is mounted: a token in an argument or an environment variable
	// would be in the service's spec for anybody who inspects it.
	tokenKey  = "TRIVY_TOKEN"
	tokenPath = "/run/hivepaas/trivy-token" //nolint:gosec // a path, not a credential
	// secretFileMode is the mode a secret file is mounted with.
	secretFileMode = "0444"

	// The server keeps the DB and the layers it analyzed in a volume of its
	// node; one that moves to another node fetches the DB again, in under a
	// minute.
	serverDataDir    = "/data"
	serverDataVolume = "hivepaas-trivy-data"
	// The runner keeps Trivy's Java DB in a volume of each node: the jar
	// analyzer runs where the image is, and the DB is a 935 MiB download.
	runnerCacheDir    = "/root/.cache/trivy"
	runnerCacheVolume = "hivepaas-image-scan-cache"

	kindEngine = "trivy"
)

// serverURL is where the runner reaches the server: the server's key is its
// name on the environment's network.
func serverURL() string {
	return fmt.Sprintf("http://%s:%d", base.HivepaasTrivyKey, serverPort)
}

// serverCommand is the server's command line after `sh`: the token read from
// its file, the DB fetched from a mirror when one is named.
func serverCommand(dbRepository string) string {
	script := fmt.Sprintf(`exec trivy server --listen 0.0.0.0:%d --cache-dir %s --token "$(cat %s)"`,
		serverPort, serverDataDir, tokenPath)
	if dbRepository != "" {
		script += " --db-repository " + executil.ArgQuote(dbRepository)
	}
	return "-c " + executil.ArgQuote(script)
}

// runnerCommand keeps the runner up between scans, which are commands run in
// it: a shell that waits, and leaves at once when told to stop.
func runnerCommand() string {
	return "-c " + executil.ArgQuote(`trap "exit 0" TERM INT; while :; do sleep 3600 & wait $!; done`)
}

// appDocInput is what an app's document needs that is decided elsewhere.
type appDocInput struct {
	Image      string
	Entrypoint string
	Command    string
	Token      string
	Resources  systemappservice.Resources
}

// buildAppDoc is the document an app of image scanning is built from - the
// same a template's app is, so it cannot drift from what an app can be. What a
// document may not say - a volume of the node, one task on every node - is
// customizeSpec's.
func buildAppDoc(in appDocInput) (*specmodel.AppDoc, error) {
	deployment := &specmodel.Deployment{
		Source: map[string]any{
			"activeMethod": string(base.DeploymentMethodImage),
			"imageSource":  map[string]any{"image": in.Image},
			"entrypoint":   in.Entrypoint,
			"command":      in.Command,
		},
		// The server answers once its DB is there, which on a first start is
		// a download; the runner only waits. Neither is checked.
		Container: &specmodel.Container{Healthcheck: &specmodel.Healthcheck{Enabled: false}},
		Resources: &specmodel.Resources{
			Limits: &specmodel.ResourceLimits{
				CPUs:   in.Resources.CPULimit,
				Memory: unit.DataSize(in.Resources.MemoryLimit),
			},
			Capabilities: &specmodel.Capabilities{OomScoreAdj: base.OomScoreAdjSystemAddon},
		},
	}
	settings := map[string]any{
		specmodel.SingletonBlockName(base.SettingTypeAppKind): map[string]any{
			"category": string(base.AppCategoryWebapp),
			"engine":   kindEngine,
			"webapp":   map[string]any{},
		},
		specmodel.CollectionBlockName(base.SettingTypeSecret): map[string]any{
			tokenKey: map[string]any{
				"value": in.Token,
				"swarmRef": map[string]any{
					"file": map[string]any{"name": tokenPath, "mode": secretFileMode},
				},
			},
		},
	}
	doc := &specmodel.AppDoc{Deployment: deployment, Settings: settings}
	if err := specmodel.CheckBuildable(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return doc, nil
}

// customizeServer gives the server the volume of its node it keeps the DB in.
func customizeServer(spec *swarm.ServiceSpec) error {
	addVolume(spec, serverDataVolume, serverDataDir)
	return nil
}

// customizeRunner makes the runner one task on every node, with the volume of
// the node it keeps the Java DB in. The node's Docker socket is the app's
// Docker API access, in host mode: see ensureHostAccess.
func customizeRunner(spec *swarm.ServiceSpec) error {
	// Exactly one mode: the daemon rejects a spec carrying two.
	spec.Mode = swarm.ServiceMode{Global: &swarm.GlobalService{}}
	addVolume(spec, runnerCacheVolume, runnerCacheDir)
	return nil
}

func addVolume(spec *swarm.ServiceSpec, volume, target string) {
	container := spec.TaskTemplate.ContainerSpec
	for _, m := range container.Mounts {
		if m.Target == target {
			return
		}
	}
	container.Mounts = append(container.Mounts, mount.Mount{Type: mount.TypeVolume, Source: volume, Target: target})
}

// imageWithoutDigest is a reference as it is shown: the digest a reference
// may be pinned with says nothing to a person.
func imageWithoutDigest(ref string) string {
	if at := strings.LastIndex(ref, "@"); at > 0 {
		return ref[:at]
	}
	return ref
}
