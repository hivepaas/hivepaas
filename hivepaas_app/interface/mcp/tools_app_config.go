package mcp

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// maxBundleFile is the most of one file in an export bundle that is read. An
// app's document is a few kilobytes.
const maxBundleFile = 4 << 20

var errNoAppDocument = errors.New("mcp: the export has no document for the app")

type appConfig struct {
	App string `json:"app"`
	// YAML is the app's document in the configuration spec, the format
	// System settings › Configuration exports and imports.
	YAML string `json:"yaml"`
}

func getAppConfigTool() Tool {
	return readTool("get_app_config", "Read an app's configuration",
		"Answers an app's whole configuration as YAML, in the format HivePaaS exports and imports: "+
			"its source, env vars, mounts, routing, resources, health check and every other setting. "+
			"Secret values are never in it; a secret is shown by name only.",
		func(ctx context.Context, call *Call, in appInput) (appConfig, error) {
			ref, err := in.resolve(ctx, call)
			if err != nil {
				return appConfig{}, err
			}
			// omit is the one mode that reveals nothing, and needs no capability.
			resp, err := call.Do(ctx, http.MethodPost, ref.path("/spec/export"), nil,
				map[string]string{"secretsMode": "omit"})
			if err != nil {
				return appConfig{}, err
			}
			doc, err := appDocument(resp.Body, ref.AppKey)
			if err != nil {
				return appConfig{}, err
			}
			return appConfig{App: ref.AppKey,
				YAML: cutText(doc, logBudget, "get_app answers the app in brief")}, nil
		})
}

// appDocument finds the app's own document in an export bundle: a gzipped tar
// whose env document, projects/<project>/envs/<env>.yaml, holds its apps by key.
func appDocument(bundle []byte, appKey string) (string, error) {
	gz, err := gzip.NewReader(bytes.NewReader(bundle))
	if err != nil {
		return "", fmt.Errorf("mcp: reading the export: %w", err)
	}
	defer func() { _ = gz.Close() }()
	archive := tar.NewReader(gz)
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return "", errNoAppDocument
		}
		if err != nil {
			return "", fmt.Errorf("mcp: reading the export: %w", err)
		}
		parts := strings.Split(path.Clean(header.Name), "/")
		if len(parts) != 4 || parts[0] != "projects" || parts[2] != "envs" || path.Ext(parts[3]) != ".yaml" {
			continue
		}
		content, err := io.ReadAll(io.LimitReader(archive, maxBundleFile))
		if err != nil {
			return "", fmt.Errorf("mcp: reading the export: %w", err)
		}
		if doc, found, err := appNode(content, appKey); err != nil || found {
			return doc, err
		}
	}
}

// appNode is the YAML of apps.<appKey> in an env document, in the order the
// export wrote it.
func appNode(envDoc []byte, appKey string) (string, bool, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(envDoc, &root); err != nil {
		return "", false, fmt.Errorf("mcp: reading the export: %w", err)
	}
	if len(root.Content) == 0 {
		return "", false, nil
	}
	apps := mappingValue(root.Content[0], "apps")
	if apps == nil {
		return "", false, nil
	}
	app := mappingValue(apps, appKey)
	if app == nil {
		return "", false, nil
	}
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2) //nolint:mnd
	if err := enc.Encode(app); err != nil {
		return "", false, fmt.Errorf("mcp: writing the app's document: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", false, fmt.Errorf("mcp: writing the app's document: %w", err)
	}
	return b.String(), true, nil
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
