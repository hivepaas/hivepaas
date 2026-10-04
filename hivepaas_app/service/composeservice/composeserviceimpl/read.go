package composeserviceimpl

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/compose-spec/compose-go/v2/dotenv"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/template"
	"github.com/compose-spec/compose-go/v2/tree"
	"github.com/compose-spec/compose-go/v2/types"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
)

const (
	composeMaxBytes = 1 << 20
	// fileMaxBytes is Docker's limit for a config, which a file becomes.
	fileMaxBytes  = 500 << 10
	filesMaxBytes = 5 << 20
	filesMax      = 100
	servicesMax   = 50

	// composeFileName is the compose file's name in the scratch directory.
	composeFileName = "compose.yaml"
	// placeholderName names the project while it is read, unless the file names
	// it: compose-go needs one, and nothing here uses it.
	placeholderName = "hivepaas-compose"
)

// read is a compose file read: the project compose-go made of it, and what
// reading it took and found.
type read struct {
	project *types.Project
	// raw is the file as YAML, before anything was done to it.
	raw map[string]any
	// variables are those the file uses, by name; values those given - by the
	// .env, then the review - and secret those kept as env secrets.
	variables map[string]template.Variable
	values    map[string]string
	secret    map[string]bool
	// plain are the variables whose names read as a secret's that the review
	// says are not: a variable they fill is not kept as a secret either.
	plain map[string]bool
	// files are the files the request carries, by their cleaned path.
	files map[string][]byte
	// markers stands in for a secret variable's value while the file is read
	// (secretMarkers).
	markers *secretMarkers
	// unsupported are what compose-go found that HivePaaS does not carry, by
	// service.
	unsupported map[string][]string
	// missing are the required variables with no value.
	missing []string
}

// readCompose reads a compose file the way `docker compose config` would, and
// touches nothing on the server: compose-go loads it in an empty scratch
// directory holding only the request's files, from which a path that is
// absolute or leaves it cannot load anything; it reads no env_file or
// label_file itself, and its environment is the request's alone.
func readCompose(ctx context.Context, req *composeservice.ConvertReq) (*read, error) {
	files, err := cleanFiles(req.Files)
	if err != nil {
		return nil, err
	}
	if len(req.Compose) > composeMaxBytes {
		return nil, hperrors.Wrap(hperrors.ErrComposeTooBig)
	}
	r := &read{files: files, unsupported: map[string][]string{}}
	if err = yaml.Unmarshal([]byte(req.Compose), &r.raw); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrComposeInvalid).WithExtraDetail("%s", err.Error())
	}
	if r.raw == nil {
		return nil, hperrors.Wrap(hperrors.ErrComposeNoServices)
	}
	if err = r.readVariables(req); err != nil {
		return nil, err
	}
	if len(r.missing) > 0 {
		return r, nil
	}
	if r.project, err = r.load(ctx, req); err != nil {
		return nil, err
	}
	if len(r.project.Services) > servicesMax {
		return nil, hperrors.Wrap(hperrors.ErrComposeTooBig).WithExtraDetail("%d services, %d at most",
			len(r.project.Services), servicesMax)
	}
	return r, nil
}

// cleanFiles keys the request's files by their path as compose writes it,
// relative to the compose file: a path that is absolute or leaves that
// directory is refused.
func cleanFiles(files map[string][]byte) (map[string][]byte, error) {
	if len(files) > filesMax {
		return nil, hperrors.Wrap(hperrors.ErrComposeTooBig).WithExtraDetail("%d files, %d at most", len(files), filesMax)
	}
	out := make(map[string][]byte, len(files))
	total := 0
	for name, content := range files {
		cleaned, ok := cleanPath(name)
		if !ok {
			return nil, hperrors.Wrap(hperrors.ErrComposeFilePath).WithExtraDetail("%s", name)
		}
		if len(content) > fileMaxBytes {
			return nil, hperrors.Wrap(hperrors.ErrComposeTooBig).WithExtraDetail("%s", name)
		}
		total += len(content)
		out[cleaned] = content
	}
	if total > filesMaxBytes {
		return nil, hperrors.Wrap(hperrors.ErrComposeTooBig)
	}
	return out, nil
}

// cleanPath is a path relative to the compose file as one key: `./a/../b` is
// `b`. false for one that is absolute or leaves the directory.
func cleanPath(p string) (string, bool) {
	p = filepath.ToSlash(strings.TrimSpace(p))
	if p == "" || path.IsAbs(p) || filepath.IsAbs(p) {
		return "", false
	}
	cleaned := path.Clean(p)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false
	}
	return cleaned, true
}

// readVariables finds the variables the file uses, and what they are given:
// the .env's values, then the review's, over them. A required one with neither
// is missing, and nothing more is read until it has one.
func (r *read) readVariables(req *composeservice.ConvertReq) error {
	r.variables = template.ExtractVariables(r.raw, template.DefaultPattern)
	given := func(name string) (string, bool) {
		if v := req.Variables[name]; v != nil && v.Value != nil {
			return *v.Value, true
		}
		return "", false
	}
	values, err := dotenv.ParseWithLookup(strings.NewReader(req.DotEnv), given)
	if err != nil {
		return hperrors.Wrap(hperrors.ErrComposeInvalid).WithExtraDetail(".env: %s", err.Error())
	}
	for name, v := range req.Variables {
		if v != nil && v.Value != nil {
			values[name] = *v.Value
		}
	}
	r.values = values

	r.secret, r.plain = map[string]bool{}, map[string]bool{}
	for _, name := range slices.Sorted(maps.Keys(r.variables)) {
		if v := req.Variables[name]; v != nil && v.Secret != nil {
			r.secret[name] = *v.Secret
			r.plain[name] = !*v.Secret && secretByName(name)
		} else {
			r.secret[name] = secretByName(name)
		}
		if r.variables[name].Required && values[name] == "" {
			r.missing = append(r.missing, name)
		}
	}
	r.markers = newSecretMarkers()
	return nil
}

// secretWords are the words a variable's name holds to be taken for a secret.
var secretWords = []string{"PASSWORD", "PASSWD", "PASS", "SECRET", "TOKEN", "KEY", "PRIVATE", "CREDENTIALS", "SALT"}

// secretByName says whether a variable's name reads as a secret's: one of
// its words, split at underscores, is one of secretWords.
func secretByName(name string) bool {
	for word := range strings.SplitSeq(strings.ToUpper(name), "_") {
		if slices.Contains(secretWords, word) {
			return true
		}
	}
	return false
}

// load has compose-go read the file in a scratch directory holding only the
// request's files, removed after.
func (r *read) load(ctx context.Context, req *composeservice.ConvertReq) (*types.Project, error) {
	dir, err := os.MkdirTemp("", "hivepaas-compose-*")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	for name, content := range r.files {
		target := filepath.Join(dir, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil { //nolint:mnd // owner only
			return nil, hperrors.Wrap(err)
		}
		if err = os.WriteFile(target, content, 0o600); err != nil { //nolint:mnd // owner only
			return nil, hperrors.Wrap(err)
		}
	}

	environment := types.Mapping{}
	for name, value := range r.values {
		environment[name] = value
		if marked, ok := r.marked(name, value); ok {
			environment[name] = marked
		}
	}
	// One with neither a value nor a default is empty either way; given as
	// such, compose-go does not log that it is not set - an issue says so.
	for name, v := range r.variables {
		if _, given := environment[name]; !given && v.DefaultValue == "" && !v.Required {
			environment[name] = ""
		}
	}
	details := types.ConfigDetails{
		WorkingDir:  dir,
		ConfigFiles: []types.ConfigFile{{Filename: filepath.Join(dir, composeFileName), Content: []byte(req.Compose)}},
		Environment: environment,
	}
	project, err := loader.LoadWithContext(ctx, details, func(o *loader.Options) {
		o.SetProjectName(placeholderName, false)
		o.SkipResolveEnvironment = true
		o.SkipResolveLabels = true
		o.ResolvePaths = false
		o.Profiles = req.Profiles
		o.ResourceLoaders = []loader.ResourceLoader{sandboxLoader{dir: dir}}
	}, loader.WithUnsupportedAttributesCheck(unsupportedPatterns(), r.noteUnsupported))
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrComposeInvalid).WithExtraDetail("%s", strings.ReplaceAll(err.Error(), dir, "."))
	}
	return project, nil
}

// sandboxLoader is the only loader compose-go has: it accepts every path, so
// that neither a remote one nor its own local one is asked, and loads none
// but a file of the scratch directory.
type sandboxLoader struct {
	dir string
}

func (l sandboxLoader) Accept(string) bool { return true }

func (l sandboxLoader) Load(_ context.Context, p string) (string, error) {
	rel := p
	if filepath.IsAbs(p) {
		var err error
		if rel, err = filepath.Rel(l.dir, p); err != nil {
			return "", hperrors.Wrap(hperrors.ErrComposeFilePath).WithExtraDetail("%s", p)
		}
	}
	cleaned, ok := cleanPath(rel)
	if !ok {
		return "", hperrors.Wrap(hperrors.ErrComposeFilePath).WithExtraDetail("%s", p)
	}
	target := filepath.Join(l.dir, filepath.FromSlash(cleaned))
	if _, err := os.Stat(target); err != nil {
		return "", hperrors.Wrap(hperrors.ErrComposeInvalid).WithExtraDetail("%s is not among the files given", cleaned)
	}
	return target, nil
}

func (l sandboxLoader) Dir(p string) string {
	return filepath.Dir(p)
}

// unsupportedFields are what a service may say that HivePaaS does not carry to
// its app: Swarm services have none of them, or HivePaaS does not set it.
var unsupportedFields = []string{
	"privileged", "devices", "network_mode", "pid", "ipc", "uts", "userns_mode", "cgroup", "cgroup_parent",
	"runtime", "isolation", "platform", "cpu_shares", "cpuset", "cpu_count", "cpu_percent", "cpu_period",
	"cpu_quota", "cpu_rt_period", "cpu_rt_runtime", "blkio_config", "oom_kill_disable", "storage_opt",
	"volumes_from", "external_links", "mac_address", "domainname", "post_start", "pre_stop", "pre_start",
	"develop", "pull_policy", "device_cgroup_rules", "credential_spec", "annotations", "volume_driver",
	"label_file", "use_api_socket", "models", "provider", "attach", "deploy.update_config",
	"deploy.rollback_config", "deploy.endpoint_mode", "deploy.placement.max_replicas_per_node",
}

func unsupportedPatterns() []loader.UnsupportedAttributePattern {
	out := make([]loader.UnsupportedAttributePattern, 0, len(unsupportedFields))
	for _, field := range unsupportedFields {
		out = append(out, loader.UnsupportedAttributePattern{
			Path: tree.NewPath("services", tree.PathMatchAll).Next(field),
		})
	}
	return out
}

// noteUnsupported keeps what the check found by service: services.<name>.<field>.
func (r *read) noteUnsupported(found []loader.UnsupportedAttribute) {
	for _, attribute := range found {
		parts := attribute.Path.Parts()
		if len(parts) < 3 || parts[0] != "services" { //nolint:mnd // services.<name>.<field>
			continue
		}
		field := strings.Join(parts[2:], ".")
		if !slices.Contains(r.unsupported[parts[1]], field) {
			r.unsupported[parts[1]] = append(r.unsupported[parts[1]], field)
		}
	}
}

// secretMarkers stand in for secret variables' values while the file is read,
// so that what reads one afterwards knows it read a secret: an environment
// value refers to the env secret, anything else is given the value. A marker
// holds a nonce of this read, and characters no file writes.
type secretMarkers struct {
	nonce string
}

const (
	markerOpen  = ""
	markerClose = ""
)

func newSecretMarkers() *secretMarkers {
	nonce := make([]byte, 8) //nolint:mnd // bytes
	_, _ = rand.Read(nonce)
	return &secretMarkers{nonce: hex.EncodeToString(nonce)}
}

// The kinds of marker: a secret variable's, and a plain one's - one whose name
// reads as a secret's that the review says is not.
const (
	secretKind = ":"
	plainKind  = "="
)

func (m *secretMarkers) of(name string) string {
	return fmt.Sprintf("%s%s%s%s%s", markerOpen, m.nonce, secretKind, name, markerClose)
}

func (m *secretMarkers) plainOf(name string) string {
	return fmt.Sprintf("%s%s%s%s%s", markerOpen, m.nonce, plainKind, name, markerClose)
}

// marked is what a variable is read as: a marker for a secret one or a plain
// one with a value; false for any other, read as its value.
func (r *read) marked(name, value string) (string, bool) {
	switch {
	case value == "":
		return "", false
	case r.secret[name]:
		return r.markers.of(name), true
	case r.plain[name]:
		return r.markers.plainOf(name), true
	}
	return "", false
}

// replace writes each secret marker of s as what with makes of its
// variable's name; names are the variables it found.
func (m *secretMarkers) replace(s string, with func(name string) string) (string, []string) {
	return m.replaceKind(s, secretKind, with)
}

// replacePlain writes each plain marker of s as its variable's value.
func (m *secretMarkers) replacePlain(s string, values map[string]string) (string, []string) {
	return m.replaceKind(s, plainKind, func(name string) string { return values[name] })
}

func (m *secretMarkers) replaceKind(s, kind string, with func(name string) string) (string, []string) {
	if !strings.Contains(s, markerOpen) {
		return s, nil
	}
	var b strings.Builder
	var names []string
	prefix := markerOpen + m.nonce + kind
	for {
		start := strings.Index(s, prefix)
		if start < 0 {
			b.WriteString(s)
			return b.String(), names
		}
		end := strings.Index(s[start:], markerClose)
		if end < 0 {
			b.WriteString(s)
			return b.String(), names
		}
		name := s[start+len(prefix) : start+end]
		b.WriteString(s[:start])
		b.WriteString(with(name))
		names = append(names, name)
		s = s[start+end+len(markerClose):]
	}
}
