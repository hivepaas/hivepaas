#!/usr/bin/env bash
# Formats and lints a Go file right after Claude Code edits it, so a formatting or lint problem
# comes back in the same turn instead of surfacing on a later manual run.
#
# Wired up as a PostToolUse hook for Edit|Write. Anything that is not a Go file inside a Go module
# is a no-op, so this stays silent in other projects.

set -uo pipefail

payload=$(cat)
file=$(printf '%s' "$payload" | jq -r '.tool_input.file_path // empty' 2>/dev/null)

[[ -n "${file}" && "${file}" == *.go && -f "${file}" ]] || exit 0

dir=$(cd "$(dirname "${file}")" 2>/dev/null && pwd) || exit 0

# golangci-lint resolves .golangci.yaml from the module root and quietly falls back to its own
# defaults when started anywhere else, so locate the root instead of trusting the current dir.
root="${dir}"
while [[ "${root}" != "/" && ! -f "${root}/go.mod" ]]; do
    root=$(dirname "${root}")
done
[[ -f "${root}/go.mod" ]] || exit 0

# gofmt is enabled as a golangci-lint formatter in this repo's .golangci.yaml, so `--fix` below
# already applies it. goimports stays off there ("SO SLOW"), which is why gci - not goimports -
# is what orders imports.
command -v golangci-lint >/dev/null 2>&1 || exit 0

# Lint the one package that was touched, not the tree: ~1s instead of minutes.
if [[ "${dir}" == "${root}" ]]; then
    package="."
else
    package="./${dir#"${root}"/}"
fi

output=$(cd "${root}" && golangci-lint run --fix --timeout=90s "${package}" 2>&1)
status=$?

[[ ${status} -eq 0 ]] && exit 0

# Exit 2 puts this in front of Claude. That is the point: it fixes the finding now, instead of
# handing over a file that fails `make lint`.
{
    echo "golangci-lint found issues in ${package} (auto-fixable ones were already applied):"
    echo "${output}"
} >&2
exit 2
