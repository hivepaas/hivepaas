package composeserviceimpl

import (
	"bytes"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
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
	// missingFiles are the compose files and env files the file reads through
	// an include or extends that the request lacks: nothing more is read until
	// it has them.
	missingFiles []*composeservice.FileNeed
	// includeEnvFiles are the env files includes name, as files of the
	// request's.
	includeEnvFiles []string
	// dir is the scratch directory the project was read in, gone since:
	// compose-go makes a path of an included or extended file absolute in it.
	dir string
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
	if err = r.checkIncludes(); err != nil {
		return nil, err
	}
	if err = r.readVariables(req); err != nil {
		return nil, err
	}
	if r.missingFiles = r.missingIncludeEnvFiles(); len(r.missingFiles) > 0 {
		return r, nil
	}
	// The files an include or extends reads are known once compose-go has read
	// them: their variables are the file's too, and with them it reads again.
	for len(r.missing) == 0 && r.project == nil {
		project, sb, loadErr := r.load(ctx, req)
		if sb == nil {
			return nil, loadErr
		}
		if r.loadedVariables(sb) {
			r.classifyVariables(req)
			continue
		}
		if loadErr != nil {
			if missing := sb.missingFiles(); len(missing) > 0 {
				for _, p := range missing {
					r.missingFiles = append(r.missingFiles, &composeservice.FileNeed{Path: p, As: composeservice.NeedCompose})
				}
				return r, nil
			}
			return nil, loadErr
		}
		r.project, r.dir = project, sb.dir
	}
	if r.project == nil {
		return r, nil
	}
	r.localPaths()
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

// checkIncludes refuses an include compose-go would read a file of the
// server's through. compose-go reads an include's env files itself, from the
// disk and not through the loader: each has to be a file of the request's,
// relative to the file naming it, as written - a value holding a variable is
// only a path once interpolated, and is refused. Its project_directory is
// refused too: the files an included file names are read from its own
// directory, as the loader reads them. Every file compose-go may load is the
// compose file or one of the request's, so each is checked.
func (r *read) checkIncludes() error {
	if err := r.checkIncludesOf("", composeFileName, r.raw); err != nil {
		return err
	}
	for _, name := range slices.Sorted(maps.Keys(r.files)) {
		if !strings.Contains(string(r.files[name]), "include") {
			continue
		}
		var doc map[string]any
		if yaml.Unmarshal(r.files[name], &doc) != nil {
			continue
		}
		base := path.Dir(name)
		if base == "." {
			base = ""
		}
		if err := r.checkIncludesOf(base, name, doc); err != nil {
			return err
		}
	}
	return nil
}

// checkIncludesOf checks the includes of one file, whose directory is base,
// and keeps the env files they name.
func (r *read) checkIncludesOf(base, file string, doc map[string]any) error {
	entries, _ := doc["include"].([]any)
	for _, entry := range entries {
		fields, _ := entry.(map[string]any)
		if _, found := fields["project_directory"]; found {
			return hperrors.Wrap(hperrors.ErrComposeInvalid).WithExtraDetail(
				"%s: include: project_directory is not supported - an included file's directory is its project's", file)
		}
		var envFiles []any
		switch value := fields["env_file"].(type) {
		case nil:
		case []any:
			envFiles = value
		default:
			envFiles = []any{value}
		}
		for _, p := range envFiles {
			s, ok := p.(string)
			if s == "/dev/null" { // compose-go reads nothing for it
				continue
			}
			if !ok || strings.Contains(s, "$") || path.IsAbs(s) || filepath.IsAbs(s) {
				return hperrors.Wrap(hperrors.ErrComposeFilePath).WithExtraDetail("%s: include: %v", file, p)
			}
			rel, inside := cleanPath(path.Join(base, filepath.ToSlash(strings.TrimSpace(s))))
			if !inside {
				return hperrors.Wrap(hperrors.ErrComposeFilePath).WithExtraDetail("%s: include: %s", file, s)
			}
			if !slices.Contains(r.includeEnvFiles, rel) {
				r.includeEnvFiles = append(r.includeEnvFiles, rel)
			}
		}
	}
	return nil
}

// missingIncludeEnvFiles are the env files includes name that the request
// lacks: compose-go cannot read the file without them.
func (r *read) missingIncludeEnvFiles() []*composeservice.FileNeed {
	var out []*composeservice.FileNeed
	for _, rel := range r.includeEnvFiles {
		if _, given := r.files[rel]; !given {
			out = append(out, &composeservice.FileNeed{Path: rel, As: composeservice.NeedEnvFile})
		}
	}
	return out
}

// localPaths makes the paths compose-go made absolute in the scratch
// directory - those of included files - relative to the compose file again,
// as the converter reads every path.
func (r *read) localPaths() {
	for name, svc := range r.project.Services {
		for i := range svc.Volumes {
			if svc.Volumes[i].Type == types.VolumeTypeBind {
				svc.Volumes[i].Source = r.local(svc.Volumes[i].Source)
			}
		}
		for i := range svc.EnvFiles {
			svc.EnvFiles[i].Path = r.local(svc.EnvFiles[i].Path)
		}
		r.project.Services[name] = svc
	}
	for name, secret := range r.project.Secrets {
		secret.File = r.local(secret.File)
		r.project.Secrets[name] = secret
	}
	for name, config := range r.project.Configs {
		config.File = r.local(config.File)
		r.project.Configs[name] = config
	}
}

// local is a path of the compose file's as the converter reads it: relative
// to the compose file. compose-go makes the relative paths of an included
// file absolute, in the scratch directory, and one there is relative again.
// One elsewhere in HivePaaS's own data directory - where a relative path
// climbing out of the scratch directory lands - is no path a compose file
// mounts: it reads as one leaving the compose file's directory. Any other is as
// written.
func (r *read) local(p string) string {
	if r.dir == "" || !filepath.IsAbs(p) {
		return p
	}
	if rel, err := filepath.Rel(r.dir, p); err == nil {
		if cleaned, ok := cleanPath(rel); ok {
			return cleaned
		}
	}
	data := filepath.Dir(fileutil.AppTempDir())
	if rel, err := filepath.Rel(data, p); err == nil && rel != ".." && !strings.HasPrefix(rel, "../") {
		return "../" + filepath.Base(p)
	}
	return p
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
	for _, rel := range r.includeEnvFiles {
		r.addValues(rel)
	}
	r.markers = newSecretMarkers()
	r.classifyVariables(req)
	return nil
}

// addValues adds the values of an env file of the request's an include reads
// to those of the variables it does not already give - the .env's and the
// review's come first, as compose's own environment does - and says whether
// it added any.
func (r *read) addValues(rel string) bool {
	content, given := r.files[rel]
	if !given {
		return false
	}
	values, err := dotenv.ParseWithLookup(bytes.NewReader(content), func(name string) (string, bool) {
		value, ok := r.values[name]
		return value, ok
	})
	if err != nil {
		return false
	}
	added := false
	for name, value := range values {
		if _, known := r.values[name]; !known {
			r.values[name] = value
			added = true
		}
	}
	return added
}

// loadedVariables adds the variables of the compose files compose-go read
// through an include or extends, and the values of the .env beside each, to
// the file's; true when it found any, and the file is read again with them.
func (r *read) loadedVariables(sb *sandbox) bool {
	found := false
	for _, rel := range sb.loadedFiles() {
		var doc map[string]any
		if yaml.Unmarshal(r.files[rel], &doc) == nil && doc != nil {
			for name, v := range template.ExtractVariables(doc, template.DefaultPattern) {
				if _, known := r.variables[name]; !known {
					r.variables[name] = v
					found = true
				}
			}
		}
		if r.addValues(path.Join(path.Dir(rel), ".env")) {
			found = true
		}
	}
	return found
}

// classifyVariables says of each variable whether its value is kept as an env
// secret, and which required ones have none: nothing more is read until they
// have.
func (r *read) classifyVariables(req *composeservice.ConvertReq) {
	r.secret, r.plain, r.missing = map[string]bool{}, map[string]bool{}, nil
	for _, name := range slices.Sorted(maps.Keys(r.variables)) {
		if v := req.Variables[name]; v != nil && v.Secret != nil {
			r.secret[name] = *v.Secret
			r.plain[name] = !*v.Secret && secretByName(name)
		} else {
			r.secret[name] = secretByName(name)
		}
		if r.variables[name].Required && r.values[name] == "" {
			r.missing = append(r.missing, name)
		}
	}
}

// secretWords are the words a variable's name holds to be taken for a secret.
var secretWords = []string{"PASSWORD", "PASSWD", "PASS", "SECRET", "TOKEN", "KEY", "PRIVATE", "CREDENTIALS", "SALT"}

// pathWords end a variable's name that holds where a secret is rather than
// the secret: POSTGRES_PASSWORD_FILE, SSL_KEY_PATH.
var pathWords = []string{"FILE", "PATH", "DIR"}

// secretByName says whether a variable's name reads as a secret's: one of
// its words, split at underscores, is one of secretWords - and its last is
// none of pathWords.
func secretByName(name string) bool {
	words := strings.Split(strings.ToUpper(name), "_")
	if slices.Contains(pathWords, words[len(words)-1]) {
		return false
	}
	return slices.ContainsFunc(words, func(word string) bool { return slices.Contains(secretWords, word) })
}

// load has compose-go read the file in a scratch directory, which holds the
// request's files it asks for (sandbox), removed after. The directory is one
// of the day's in the app's data directory: should the process die before
// removing it, the system cleanup removes the day's directory a few days on.
func (r *read) load(ctx context.Context, req *composeservice.ConvertReq) (*types.Project, *sandbox, error) {
	dir, err := fileutil.CreateTempDirInAppPath("", "compose-*", 0)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	sb := newSandbox(dir, r.files)
	for _, rel := range r.includeEnvFiles {
		if err = sb.write(rel); err != nil {
			return nil, sb, err
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
		o.ResourceLoaders = []loader.ResourceLoader{sb}
	}, loader.WithUnsupportedAttributesCheck(unsupportedPatterns(), r.noteUnsupported))
	if err != nil {
		return nil, sb, hperrors.Wrap(hperrors.ErrComposeInvalid).WithExtraDetail("%s",
			strings.ReplaceAll(err.Error(), dir, "."))
	}
	return project, sb, nil
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
