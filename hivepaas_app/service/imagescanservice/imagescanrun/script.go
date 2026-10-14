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

// Script is what the scan's container runs: Trivy on the image, then trivy
// version, for the DB's date a standalone report does not carry, and both in
// one JSON object on the standard output - imagescanservice.ScanOutput. It
// exits with Trivy's code.
func Script(req *RunReq, source base.ImageScanSource, insecure bool) string {
	cacheDir := executil.ArgQuote(req.CacheDir)
	args := []string{
		"trivy", "image",
		"--cache-dir", cacheDir,
		"--scanners", "vuln",
		"--list-all-pkgs=false",
		"--disable-telemetry",
		"--skip-version-check",
		"--no-progress",
		"--format", "json",
		"--timeout", req.Timeout.String(),
		"$skip",
	}
	if req.DBRepository != "" {
		args = append(args, "--db-repository", executil.ArgQuote(req.DBRepository))
	}
	if req.JavaDBRepository != "" {
		args = append(args, "--java-db-repository", executil.ArgQuote(req.JavaDBRepository))
	}
	if source == base.ImageScanSourceNode {
		args = append(args, "--image-src", "docker", req.ImageID)
	} else {
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
