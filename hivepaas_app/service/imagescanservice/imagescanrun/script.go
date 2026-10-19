package imagescanrun

import (
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
)

// javaDBMaxAgeDays is how old the Java DB on the volume may be before Trivy
// is let renew it: it is a 935 MiB download, and Trivy's own is three days.
const javaDBMaxAgeDays = 7

// Script is what the scan's container runs: Trivy on the image, or on an
// artifact's file as a root filesystem, then trivy version, for the DB's date a
// standalone report does not carry, and both in one JSON object on the standard
// output - imagescanservice.ScanOutput. It exits with Trivy's code. An SBOM is
// the packages alone: no scanner, so no DB.
func Script(req *RunReq, source base.ImageScanSource, insecure bool) string {
	cacheDir := executil.ArgQuote(req.CacheDir)
	target := "image"
	if source == base.ImageScanSourceArtifact {
		target = "rootfs"
	}
	args := []string{
		"trivy", target,
		"--cache-dir", cacheDir,
		"--disable-telemetry",
		"--skip-version-check",
		"--no-progress",
		"--timeout", req.Timeout.String(),
	}
	if req.Format == FormatCycloneDX {
		args = append(args, "--format", "cyclonedx")
	} else {
		args = append(args, "--scanners", "vuln", "--list-all-pkgs=false", "--format", "json", "$skip")
		if req.DBRepository != "" {
			args = append(args, "--db-repository", executil.ArgQuote(req.DBRepository))
		}
		if req.JavaDBRepository != "" {
			args = append(args, "--java-db-repository", executil.ArgQuote(req.JavaDBRepository))
		}
	}
	switch source {
	case base.ImageScanSourceArtifact:
		args = append(args, artifactTarget)
	case base.ImageScanSourceNode:
		args = append(args, "--image-src", "docker", req.ImageID)
	case base.ImageScanSourceRegistry:
		args = append(args, "--image-src", "remote", "--platform", req.Platform)
		if insecure {
			args = append(args, "--insecure")
		}
		args = append(args, executil.ArgQuote(req.RemoteRef))
	}
	return fmt.Sprintf(`skip=""
if [ -n "$(find %[1]s/java-db/metadata.json -mtime -%[2]d 2>/dev/null)" ]; then skip=--skip-java-db-update; fi
%[3]s > /tmp/report.json
code=$?
version=$(trivy version --format json --cache-dir %[1]s 2>/dev/null)
[ -n "$version" ] || version=null
printf '{"exitCode":%%d,"version":%%s,"report":' "$code" "$version"
if [ "$code" -eq 0 ]; then cat /tmp/report.json; else printf null; fi
printf '}\n'
exit "$code"
`, cacheDir, javaDBMaxAgeDays, strings.Join(args, " "))
}
