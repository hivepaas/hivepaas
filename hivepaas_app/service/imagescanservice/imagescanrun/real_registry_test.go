package imagescanrun

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/services/docker"
)

// TestReadingFromARealRegistry runs a scan on a node that lacks the image, as
// a scan on another node than the image's does: Trivy reads it from a registry
// that asks for a password, with the credentials the scan copies in. It needs
// a node of its own - it starts a registry and pushes and removes an image -
// and runs where the default Docker socket is that node's, such as inside the
// dashboard's e2e dind:
//
//	GOOS=linux GOARCH=<dind's> CGO_ENABLED=0 go test -c -o imagescanrun.test .
//	docker cp imagescanrun.test hp-e2e-dind:/hp/
//	docker exec -e HP_TEST_SCAN_REGISTRY=1 hp-e2e-dind /hp/imagescanrun.test \
//	  -test.run TestReadingFromARealRegistry -test.v
//
// (dind's /tmp is emptied by HivePaaS's own cleanup.)
//
// It reads the image's SBOM, which needs no vulnerability DB: what it tests is
// how the image is read, which a scan of its vulnerabilities shares.
func TestReadingFromARealRegistry(t *testing.T) {
	if os.Getenv("HP_TEST_SCAN_REGISTRY") != "1" {
		t.Skip("set HP_TEST_SCAN_REGISTRY=1 to run it on a node of its own; see its comment")
	}
	const (
		image    = "alpine:3.11"
		registry = "hp-scan-registry"
		port     = "15000"
		repo     = "hp-scan/alpine"
		user     = "scanner"
		password = "s3cret-for-the-test"
	)
	ctx := context.Background()

	// A registry that asks for a password.
	authDir := t.TempDir()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	must(t, assert.NoError(t, err))
	must(t, assert.NoError(t, os.WriteFile(filepath.Join(authDir, "htpasswd"),
		[]byte(user+":"+string(hash)+"\n"), 0o600)))
	dockerCLI(t, "run", "-d", "--name", registry, "-p", port+":5000", "-v", authDir+":/auth",
		"-e", "REGISTRY_AUTH=htpasswd", "-e", "REGISTRY_AUTH_HTPASSWD_REALM=hp-scan",
		"-e", "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd", "registry:2")
	t.Cleanup(func() {
		_, _ = runDocker("rm", "-f", "-v", registry)
		_, _ = runDocker("rmi", "registry:2")
	})

	// The image goes to it, and leaves the node.
	local := "localhost:" + port + "/" + repo + ":3.11"
	dockerCLI(t, "pull", image)
	imageID := strings.TrimSpace(dockerCLI(t, "image", "inspect", "--format", "{{.Id}}", image))
	dockerCLI(t, "tag", image, local)
	login := exec.CommandContext(ctx, "docker", "login", "localhost:"+port, "-u", user, "--password-stdin")
	login.Stdin = strings.NewReader(password)
	out, err := login.CombinedOutput()
	must(t, assert.NoError(t, err, "%s", out))
	t.Cleanup(func() { _, _ = runDocker("logout", "localhost:"+port) })
	// The digest is what the push answers: a node pushes the manifest of its
	// own platform, not the index the image was pulled by.
	digest := ""
	for range 10 { // the registry takes a moment to listen
		out, err := runDocker("push", local)
		if err == nil {
			if m := pushedDigest.FindStringSubmatch(out); m != nil {
				digest = m[1]
			}
			break
		}
		if strings.Contains(out, "unauthorized") {
			break
		}
		time.Sleep(time.Second)
	}
	must(t, assert.NotEmpty(t, digest, "pushing %s", local))
	dockerCLI(t, "rmi", local, image)
	_, err = runDocker("image", "inspect", imageID)
	must(t, assert.Error(t, err, "the node no longer has the image"))

	// Trivy's container reaches the registry at the bridge's gateway: a private
	// address, which it reads over plain HTTP.
	gateway := strings.TrimSpace(dockerCLI(t, "network", "inspect", "bridge", "--format",
		"{{(index .IPAM.Config 0).Gateway}}"))
	host := gateway + ":" + port
	auth := func(pass string) []byte {
		return []byte(fmt.Sprintf(`{"auths":{%q:{"auth":%q}}}`, host,
			base64.StdEncoding.EncodeToString([]byte(user+":"+pass))))
	}

	manager, err := docker.New()
	must(t, assert.NoError(t, err))
	cache := "hp-scan-registry-trivy"
	t.Cleanup(func() { _, _ = runDocker("volume", "rm", cache) })
	req := &RunReq{
		Format:       FormatCycloneDX,
		TrivyImage:   "aquasec/trivy:0.75.0",
		ImageID:      imageID,
		ImageNode:    "another",
		RemoteRef:    host + "/" + repo + "@" + digest,
		Platform:     runtime.GOOS + "/" + runtime.GOARCH,
		DockerConfig: auth(password),
		Volume:       mount.Mount{Type: mount.TypeVolume, Source: cache},
		CacheDir:     DataDir + "/hivepaas/scanning/trivy",
		NanoCPUs:     1e9,
		MemoryBytes:  512 << 20,
		Timeout:      5 * time.Minute,
	}
	resp, err := (&Runner{Docker: manager}).Run(ctx, req)
	must(t, assert.NoError(t, err))
	assert.Equal(t, base.ImageScanSourceRegistry, resp.Source)
	must(t, assert.Zero(t, resp.ExitCode, "trivy: %s", resp.Log))
	var envelope struct {
		Report struct {
			BOMFormat  string `json:"bomFormat"`
			Components []struct {
				Name string `json:"name"`
			} `json:"components"`
		} `json:"report"`
	}
	must(t, assert.NoError(t, json.Unmarshal(resp.Output, &envelope), "%s", resp.Output))
	assert.Equal(t, "CycloneDX", envelope.Report.BOMFormat)
	names := make([]string, 0, len(envelope.Report.Components))
	for _, c := range envelope.Report.Components {
		names = append(names, c.Name)
	}
	assert.Contains(t, names, "apk-tools", "the image's packages, read from the registry")

	// The wrong password is the registry's refusal, in Trivy's words.
	req.DockerConfig = auth("not-the-password")
	resp, err = (&Runner{Docker: manager}).Run(ctx, req)
	must(t, assert.NoError(t, err))
	assert.NotZero(t, resp.ExitCode)
	assert.Contains(t, strings.ToLower(resp.Log), "unauthorized")
}

// pushedDigest is the digest in what docker push answers.
var pushedDigest = regexp.MustCompile(`digest: (sha256:[0-9a-f]{64})`)

// must stops the test when a check failed.
func must(t *testing.T, ok bool) {
	t.Helper()
	if !ok {
		t.FailNow()
	}
}

// dockerCLI runs the node's docker CLI, and fails the test when it fails.
func dockerCLI(t *testing.T, args ...string) string {
	t.Helper()
	out, err := runDocker(args...)
	must(t, assert.NoError(t, err, "docker %s: %s", strings.Join(args, " "), out))
	return out
}

func runDocker(args ...string) (string, error) {
	out, err := exec.CommandContext(context.Background(), "docker", args...).CombinedOutput()
	return string(out), err
}
