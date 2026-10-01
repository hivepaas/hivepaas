package imagebuildserviceimpl

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// checkSecretsNotDeclaredAsArg refuses a Dockerfile that declares one of the
// build's secret variables with ARG.
//
// The secret does not reach such an ARG - it is given as a BuildKit secret, not
// as a build argument - so the build would go on without it and fail somewhere
// later, or worse, succeed with an empty value. And a build argument is what
// leaks: its value is written into the image's history, which anyone who can
// pull the image can read. Saying so before the build starts is cheaper than
// either.
func checkSecretsNotDeclaredAsArg(data *imageBuildData) error {
	if len(data.SecretEnvVars) == 0 {
		return nil
	}
	content, err := os.ReadFile(filepath.Join(data.CheckoutDir, data.Dockerfile.Path))
	if err != nil {
		return hperrors.Wrap(err)
	}
	names := secretsDeclaredAsArg(string(content), slices.Collect(maps.Keys(data.SecretEnvVars)))
	if len(names) > 0 {
		return hperrors.Wrap(hperrors.ErrBuildSecretDeclaredAsArg).
			WithParam("Names", strings.Join(names, ", ")).
			WithParam("Example", names[0])
	}
	return nil
}

// secretsDeclaredAsArg is which of secrets a Dockerfile declares with ARG,
// sorted. It reads the instruction as Docker does: the keyword in any case,
// several names on one line, defaults after '=', lines continued with a
// backslash, comments skipped. Names are matched exactly, as Docker does.
func secretsDeclaredAsArg(dockerfile string, secrets []string) []string {
	var found []string
	for _, line := range logicalLines(dockerfile) {
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.EqualFold(fields[0], "ARG") { //nolint:mnd // the keyword and a name
			continue
		}
		for _, field := range fields[1:] {
			name, _, _ := strings.Cut(field, "=")
			if slices.Contains(secrets, name) && !slices.Contains(found, name) {
				found = append(found, name)
			}
		}
	}
	slices.Sort(found)
	return found
}

// logicalLines joins the lines a backslash continues and drops comments.
func logicalLines(dockerfile string) []string {
	var lines []string
	var current strings.Builder
	for _, raw := range strings.Split(dockerfile, "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "\\") {
			current.WriteString(strings.TrimSuffix(line, "\\"))
			current.WriteString(" ")
			continue
		}
		current.WriteString(line)
		lines = append(lines, current.String())
		current.Reset()
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}
