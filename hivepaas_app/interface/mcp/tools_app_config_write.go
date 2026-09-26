package mcp

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// A configuration change goes through the configuration spec's own import,
// which already plans before it applies: validate answers what would change
// and a planHash, and apply refuses a hash that no longer matches. The tool
// exports the app, puts the model's document in place of the app's, and asks
// the env's import about that one app. See the phase 2 spec, §2.

const (
	codeSpecPlanChanged = "ERR_SPEC_IMPORT_PLAN_CHANGED"
	severityBlocked     = "blocked"
	actionUpdate        = "update"
	actionCreate        = "create"
	actionSkip          = "skip"
)

type updateConfigInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the app's key, name or id"`
	YAML    string `json:"yaml" jsonschema:"the app's whole document as get_app_config answers it, edited"`
}

type importIssue struct {
	Severity string         `json:"severity"`
	Code     string         `json:"code"`
	Path     string         `json:"path,omitempty"`
	Hint     string         `json:"hint,omitempty"`
	Detail   map[string]any `json:"detail,omitempty"`
}

type changedNode struct {
	Path    string   `json:"path"`
	Kind    string   `json:"kind"`
	Action  string   `json:"action"`
	Changes []string `json:"changes,omitempty"`
}

type configPlan struct {
	App     string `json:"app"`
	Project string `json:"project"`
	Env     string `json:"env"`
	// Action is update, unchanged, or skip when an issue keeps the app out.
	Action   string        `json:"action"`
	Changes  []string      `json:"changes"`
	Restart  bool          `json:"restart"`
	Redeploy bool          `json:"redeploy"`
	Issues   []importIssue `json:"issues,omitempty"`
	Notes    []importIssue `json:"notes,omitempty"`
	// AlsoChanged are other objects the change brings along: a setting the
	// document now refers to that the env does not have.
	AlsoChanged []changedNode `json:"alsoChanged,omitempty"`
}

type apiPlanNode struct {
	Path     string        `json:"path"`
	Kind     string        `json:"kind"`
	Key      string        `json:"key"`
	Selected bool          `json:"selected"`
	Action   string        `json:"action"`
	Changes  []string      `json:"changes"`
	Restart  bool          `json:"restart"`
	Deploy   bool          `json:"deploy"`
	Issues   []importIssue `json:"issues"`
	Notes    []importIssue `json:"notes"`
}

type apiImportPlan struct {
	Nodes    []apiPlanNode `json:"nodes"`
	PlanHash string        `json:"planHash"`
}

func planUpdateAppConfigTool() Tool {
	return planTool("plan_update_app_config", "Plan a configuration change",
		"Plans changing an app's configuration: give its whole document, as get_app_config answers it, "+
			"with your edits. Answers each setting that would change, whether the app restarts or is "+
			"deployed again - a change to its source deploys it - and anything the import would refuse or "+
			"clear. A secret's value is never in the document; a secret it adds is created empty. Nothing "+
			"changes until apply_plan.",
		NeedWrite, &applier{refused: configRefused, result: configResult,
			follow: "get_app_status shows the app restart; get_app_config reads what it has now."},
		func(ctx context.Context, call *Call, in updateConfigInput) (configPlan, *storedPlan, error) {
			doc, err := parseAppDocument(in.YAML)
			if err != nil {
				return configPlan{}, nil, err
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return configPlan{}, nil, err
			}
			exported, err := call.Do(ctx, http.MethodPost, ref.path("/spec/export"), nil,
				map[string]string{"secretsMode": "omit"})
			if err != nil {
				return configPlan{}, nil, err
			}
			bundle, err := replaceAppDocument(exported.Body, ref.AppKey, doc)
			if err != nil {
				return configPlan{}, nil, err
			}
			appPath := "projects/" + ref.ProjectKey + "/envs/" + ref.Env + "/apps/" + ref.AppKey
			body := map[string]any{"bundle": bundle,
				"selection": map[string]any{"include": []string{appPath}},
				"options":   map[string]any{"existing": actionUpdate, "deployChangedSource": true}}
			var resp struct {
				Data apiImportPlan `json:"data"`
			}
			if err = call.Post(ctx, ref.envRef.path("/spec/import/validate"), body, &resp); err != nil {
				return configPlan{}, nil, err
			}
			out, ok := makeConfigPlan(ref, appPath, &resp.Data)
			if !ok {
				return out, nil, nil
			}
			body["planHash"] = resp.Data.PlanHash
			body["acceptIssues"] = len(out.Issues) > 0
			raw, err := json.Marshal(body)
			if err != nil {
				return configPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.envRef.path("/spec/import/apply"), Body: raw,
				Summary: fmt.Sprintf("change %s in %s/%s: %s", ref.AppKey, ref.ProjectKey, ref.Env,
					strings.Join(out.Changes, ", "))}, nil
		})
}

// makeConfigPlan reads the import's plan for the app, and says whether it can
// be applied: it changes something, and nothing selected is blocked.
func makeConfigPlan(ref *appRef, appPath string, plan *apiImportPlan) (configPlan, bool) {
	out := configPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Action: "unchanged",
		Changes: []string{}}
	changes, blocked := false, false
	for _, node := range plan.Nodes {
		if !node.Selected {
			continue
		}
		for _, issue := range node.Issues {
			if issue.Severity == severityBlocked {
				blocked = true
			}
		}
		if node.Path == appPath {
			out.Action, out.Restart, out.Redeploy = node.Action, node.Restart, node.Deploy
			out.Issues, out.Notes = node.Issues, node.Notes
			if node.Changes != nil {
				out.Changes = node.Changes
			}
			changes = changes || node.Action == actionUpdate
			continue
		}
		out.Issues = append(out.Issues, node.Issues...)
		if node.Action == actionUpdate || node.Action == actionCreate {
			out.AlsoChanged = append(out.AlsoChanged, changedNode{Path: node.Path, Kind: node.Kind,
				Action: node.Action, Changes: node.Changes})
			changes = true
		}
	}
	return out, changes && !blocked && out.Action != actionSkip
}

// parseAppDocument reads the model's document: one YAML mapping, the app.
func parseAppDocument(text string) (*yaml.Node, error) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(text), &root); err != nil {
		return nil, &InputError{Message: "yaml is not YAML: " + err.Error()}
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, &InputError{Message: "yaml is the app's document: one mapping, as get_app_config answers it"}
	}
	return root.Content[0], nil
}

type bundleEntry struct {
	header  *tar.Header
	content []byte
}

// replaceAppDocument is an exported bundle with the app's document in its env
// document replaced, every file otherwise as it was, in the same order.
func replaceAppDocument(bundle []byte, appKey string, doc *yaml.Node) ([]byte, error) {
	entries, err := readBundle(bundle)
	if err != nil {
		return nil, err
	}
	replaced := false
	for _, entry := range entries {
		parts := strings.Split(path.Clean(entry.header.Name), "/")
		if len(parts) != 4 || parts[0] != "projects" || parts[2] != "envs" || path.Ext(parts[3]) != ".yaml" {
			continue
		}
		content, found, err := replaceInEnvDocument(entry.content, appKey, doc)
		if err != nil {
			return nil, err
		}
		if found {
			entry.content, replaced = content, true
		}
	}
	if !replaced {
		return nil, errNoAppDocument
	}
	return writeBundle(entries)
}

func replaceInEnvDocument(envDoc []byte, appKey string, doc *yaml.Node) ([]byte, bool, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(envDoc, &root); err != nil {
		return nil, false, fmt.Errorf("mcp: reading the export: %w", err)
	}
	if len(root.Content) == 0 {
		return nil, false, nil
	}
	apps := mappingValue(root.Content[0], "apps")
	if apps == nil || apps.Kind != yaml.MappingNode {
		return nil, false, nil
	}
	for i := 0; i+1 < len(apps.Content); i += 2 {
		if apps.Content[i].Value == appKey {
			apps.Content[i+1] = doc
			var b bytes.Buffer
			enc := yaml.NewEncoder(&b)
			enc.SetIndent(2) //nolint:mnd
			if err := enc.Encode(&root); err != nil {
				return nil, false, fmt.Errorf("mcp: writing the env document: %w", err)
			}
			if err := enc.Close(); err != nil {
				return nil, false, fmt.Errorf("mcp: writing the env document: %w", err)
			}
			return b.Bytes(), true, nil
		}
	}
	return nil, false, nil
}

func readBundle(bundle []byte) ([]*bundleEntry, error) {
	gz, err := gzip.NewReader(bytes.NewReader(bundle))
	if err != nil {
		return nil, fmt.Errorf("mcp: reading the export: %w", err)
	}
	defer func() { _ = gz.Close() }()
	archive := tar.NewReader(gz)
	var entries []*bundleEntry
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return entries, nil
		}
		if err != nil {
			return nil, fmt.Errorf("mcp: reading the export: %w", err)
		}
		content, err := io.ReadAll(io.LimitReader(archive, maxBundleFile))
		if err != nil {
			return nil, fmt.Errorf("mcp: reading the export: %w", err)
		}
		entries = append(entries, &bundleEntry{header: header, content: content})
	}
}

func writeBundle(entries []*bundleEntry) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	archive := tar.NewWriter(gz)
	for _, entry := range entries {
		header := *entry.header
		header.Size = int64(len(entry.content))
		if err := archive.WriteHeader(&header); err != nil {
			return nil, fmt.Errorf("mcp: writing the bundle: %w", err)
		}
		if _, err := archive.Write(entry.content); err != nil {
			return nil, fmt.Errorf("mcp: writing the bundle: %w", err)
		}
	}
	if err := archive.Close(); err != nil {
		return nil, fmt.Errorf("mcp: writing the bundle: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("mcp: writing the bundle: %w", err)
	}
	return buf.Bytes(), nil
}

// configRefused says a configuration plan was refused because the app changed
// since: the import's hash no longer matches what it would do now.
func configRefused(apiErr *APIError) error {
	if apiErr.Code == codeSpecPlanChanged {
		return planMoved("the app's configuration")
	}
	return nil
}

// configResult keeps of the import's answer what it did to each object, not the
// whole plan again.
func configResult(data json.RawMessage) (any, error) {
	var applied struct {
		Plan *struct {
			Nodes []struct {
				Path    string `json:"path"`
				Action  string `json:"action"`
				Outcome string `json:"outcome"`
			} `json:"nodes"`
		} `json:"plan"`
		Deployments []any `json:"deployments"`
	}
	if err := json.Unmarshal(data, &applied); err != nil {
		return nil, fmt.Errorf("mcp: decoding the answer: %w", err)
	}
	type outcome struct {
		Path    string `json:"path"`
		Action  string `json:"action"`
		Outcome string `json:"outcome,omitempty"`
	}
	out := struct {
		Outcomes    []outcome `json:"outcomes"`
		Deployments []any     `json:"deployments,omitempty"`
	}{Outcomes: []outcome{}, Deployments: applied.Deployments}
	if applied.Plan != nil {
		for _, node := range applied.Plan.Nodes {
			if node.Outcome != "" || node.Action == actionUpdate || node.Action == actionCreate {
				out.Outcomes = append(out.Outcomes, outcome(node))
			}
		}
	}
	return out, nil
}
