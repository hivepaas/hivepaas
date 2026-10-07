package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// rawGitHub is where a server fetches the templates from: the index at the
// pinned commit, which it refuses unless it hashes to the pin.
const rawGitHub = "https://raw.githubusercontent.com"

// checkTemplates checks the templates pin of each channel of a release: the
// index.json of the pinned repository at the pinned commit hashes to its
// indexSha256. A pin that does not leaves every server that updates without
// templates, so it is caught here, before the release is signed. It answers
// what it checked, as `channel: repo@commit`.
func checkTemplates(release []byte, fetch func(url string) ([]byte, error)) ([]string, error) {
	var channels map[string]struct {
		Templates *base.TemplatesRef `json:"templates"`
	}
	if err := json.Unmarshal(release, &channels); err != nil {
		return nil, fmt.Errorf("reading the release: %w", err)
	}
	names := make([]string, 0, len(channels))
	for name := range channels {
		names = append(names, name)
	}
	slices.Sort(names)

	var checked []string
	for _, name := range names {
		pin := channels[name].Templates
		if pin == nil {
			continue
		}
		url := fmt.Sprintf("%s/%s/%s/index.json", rawGitHub, pin.Repo, pin.Commit)
		index, err := fetch(url)
		if err != nil {
			return checked, fmt.Errorf("%s: the templates' index at %s: %w", name, pin.Commit, err)
		}
		sum := sha256.Sum256(index)
		if got := hex.EncodeToString(sum[:]); got != pin.IndexSHA256 {
			return checked, fmt.Errorf("%s: the templates' index at %s hashes to %s, not the pinned %s: "+
				"pin it with `go run ./tools/apptemplate pin <app-templates checkout>` at that commit",
				name, pin.Commit, got, pin.IndexSHA256)
		}
		checked = append(checked, fmt.Sprintf("%s: %s@%s", name, pin.Repo, pin.Commit))
	}
	return checked, nil
}

// fetchHTTP gets a URL's body, refusing anything but a 200.
func fetchHTTP(url string) ([]byte, error) {
	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Get(url) //nolint:noctx // a command run by hand, once
	if err != nil {
		return nil, err //nolint:wrapcheck // the caller says what was fetched
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body) //nolint:wrapcheck // the caller says what was fetched
}
