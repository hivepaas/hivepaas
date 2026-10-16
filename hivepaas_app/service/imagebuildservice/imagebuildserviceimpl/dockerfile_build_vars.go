package imagebuildserviceimpl

import (
	"maps"
	"regexp"
	"slices"
	"strings"
)

// argNameRe is a name ARG takes; another one is still a build argument, which
// the generated Dockerfile does not declare.
var argNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// withBuildVariables gives a generated Dockerfile's steps the build's variables.
//
// A generated Dockerfile declares only its own options with ARG, so a variable
// passed as a build argument - an API URL compiled into a frontend - would not
// reach its install and build steps, and a secret one, given as a BuildKit
// secret, would reach none. Each stage declares the plain ones after its FROM,
// and each RUN mounts the secret ones as environment variables: never an ARG,
// whose value the image's history keeps.
func withBuildVariables(dockerfile string, envVars map[string]*string, secretEnvVars map[string]string) string {
	args := slices.DeleteFunc(slices.Sorted(maps.Keys(envVars)), func(name string) bool {
		return !argNameRe.MatchString(name)
	})
	secrets := slices.Sorted(maps.Keys(secretEnvVars))
	if len(args) == 0 && len(secrets) == 0 {
		return dockerfile
	}

	mounts := make([]string, 0, len(secrets))
	for _, name := range secrets {
		mounts = append(mounts, "--mount=type=secret,id="+name+",env="+name)
	}

	lines := strings.Split(dockerfile, "\n")
	out := make([]string, 0, len(lines))
	// The instruction a line continues, and whether the next line continues it
	instruction, continued := "", false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		if !continued {
			keyword, _, _ := strings.Cut(trimmed, " ")
			instruction = strings.ToUpper(keyword)
			if instruction == "RUN" && len(mounts) > 0 {
				indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				line = indent + "RUN " + strings.Join(mounts, " ") + " " + strings.TrimSpace(trimmed[len(keyword):])
			}
		}
		out = append(out, line)
		continued = strings.HasSuffix(trimmed, "\\")
		if !continued && instruction == "FROM" {
			for _, name := range args {
				out = append(out, "ARG "+name)
			}
		}
	}
	return strings.Join(out, "\n")
}
